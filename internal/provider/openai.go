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
	openaiDefaultModel = "gpt-4o"
	openaiDefaultURL   = "https://api.openai.com/v1/chat/completions"
)

// OpenAIProvider implements the Provider interface for the OpenAI API.
type OpenAIProvider struct {
	apiKey string
	model  string
	client *http.Client
	url    string
}

// NewOpenAIProvider initializes a new OpenAIProvider with the given configuration and model.
func NewOpenAIProvider(_ context.Context, cfg config.ProviderConfig, modelName string) (*OpenAIProvider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("openai API key is required")
	}

	if modelName == "" {
		modelName = cfg.DefaultModel
	}
	if modelName == "" {
		modelName = openaiDefaultModel
	}

	url := cfg.BaseURL
	if url == "" {
		url = openaiDefaultURL
	}

	return &OpenAIProvider{
		apiKey: cfg.APIKey,
		model:  modelName,
		client: &http.Client{Timeout: 120 * time.Second},
		url:    url,
	}, nil
}

// Name returns the identifier for this provider.
func (p *OpenAIProvider) Name() string {
	return "openai"
}

// ValidateKey checks the validity of the API key by sending a minimal completion request.
func (p *OpenAIProvider) ValidateKey(ctx context.Context) error {
	_, err := p.Complete(ctx, CompletionRequest{
		Messages: []Message{
			{Role: "user", Content: "ping"},
		},
		MaxTokens: 1,
	})
	return err
}

// Complete sends a completion request to the OpenAI API and returns the response.
func (p *OpenAIProvider) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	model := req.Model
	if model == "" {
		model = p.model
	}

	body := openAIChatRequest{
		Model:     model,
		Messages:  mapOpenAIMessages(req.Messages),
		Tools:     mapOpenAITools(req.Tools),
		MaxTokens: req.MaxTokens,
	}

	var response openAIChatResponse
	if err := p.do(ctx, body, &response); err != nil {
		return CompletionResponse{}, err
	}
	if len(response.Choices) == 0 {
		return CompletionResponse{}, fmt.Errorf("openai returned no choices")
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

// do executes an HTTP request to the OpenAI API.
func (p *OpenAIProvider) do(ctx context.Context, body openAIChatRequest, out any) error {
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
		return fmt.Errorf("openai API error %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("failed to decode openai response: %w", err)
	}

	return nil
}
