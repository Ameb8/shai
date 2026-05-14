package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ameb8/shai/internal/config"
)

const (
	// mistralDefaultModel is the model used if no model is specified in the config or request.
	mistralDefaultModel = "mistral-small-latest"
	// mistralDefaultURL is the endpoint for Mistral's chat completion API.
	mistralDefaultURL = "https://api.mistral.ai/v1/chat/completions"
)

// MistralProvider implements the provider.Provider interface using the Mistral AI API.
type MistralProvider struct {
	apiKey string
	model  string
	client *http.Client
	url    string
}

// NewMistralProvider initializes a new Mistral provider with the given
// configuration and model name.
func NewMistralProvider(_ context.Context, cfg config.ProviderConfig, modelName string) (*MistralProvider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("mistral API key is required")
	}

	if modelName == "" {
		modelName = cfg.DefaultModel
	}
	if modelName == "" {
		modelName = mistralDefaultModel
	}

	url := cfg.BaseURL
	if url == "" {
		url = mistralDefaultURL
	}

	return &MistralProvider{
		apiKey: cfg.APIKey,
		model:  modelName,
		client: &http.Client{Timeout: 120 * time.Second},
		url:    url,
	}, nil
}

// Name returns the provider identifier "mistral".
func (p *MistralProvider) Name() string {
	return "mistral"
}

// ValidateKey checks if the API key is valid by sending a minimal completion request.
func (p *MistralProvider) ValidateKey(ctx context.Context) error {
	_, err := p.Complete(ctx, CompletionRequest{
		Messages:  []Message{{Role: "user", Content: "ping"}},
		MaxTokens: 1,
	})
	return err
}

// Complete sends a chat completion request to the Mistral API.
func (p *MistralProvider) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	model := req.Model
	if model == "" {
		model = p.model
	}

	body := mistralChatRequest{
		Model:     model,
		Messages:  mapGrokMessages(req.Messages),
		Tools:     mapGrokTools(req.Tools),
		MaxTokens: req.MaxTokens,
	}

	var response mistralChatResponse
	if err := p.do(ctx, body, &response); err != nil {
		return CompletionResponse{}, err
	}
	if len(response.Choices) == 0 {
		return CompletionResponse{}, fmt.Errorf("mistral returned no choices")
	}

	choice := response.Choices[0]
	result := CompletionResponse{
		Content:      choice.Message.Content,
		ToolCalls:    mapProviderToolCalls(choice.Message.ToolCalls),
		InputTokens:  response.Usage.PromptTokens,
		OutputTokens: response.Usage.CompletionTokens,
	}

	if len(result.ToolCalls) > 0 || choice.FinishReason == "tool_calls" {
		result.StopReason = "tool_use"
	} else if choice.FinishReason == "length" {
		result.StopReason = "max_tokens"
	} else {
		result.StopReason = "end_turn"
	}

	return result, nil
}

// do performs the HTTP POST request to the Mistral API and decodes the response.
func (p *MistralProvider) do(ctx context.Context, body mistralChatRequest, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("mistral API error %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("failed to decode mistral response: %w", err)
	}

	return nil
}

// mistralChatRequest defines the JSON structure for a Mistral chat completion request.
type mistralChatRequest struct {
	Model     string        `json:"model"`
	Messages  []grokMessage `json:"messages"`
	Tools     []grokTool    `json:"tools,omitempty"`
	MaxTokens int           `json:"max_tokens,omitempty"`
}

// mistralChatResponse defines the JSON structure for a Mistral chat completion response.
type mistralChatResponse struct {
	Choices []struct {
		FinishReason string      `json:"finish_reason"`
		Message      grokMessage `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}
