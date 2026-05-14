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
	grokDefaultModel = "grok-4.20-reasoning"
	grokDefaultURL   = "https://api.x.ai/v1/chat/completions"
)

// GrokProvider implements the Provider interface for the x.AI Grok API.
type GrokProvider struct {
	apiKey string
	model  string
	client *http.Client
	url    string
}

// NewGrokProvider initializes a new GrokProvider with the given configuration and model.
// It returns an error if the API key is missing.
func NewGrokProvider(_ context.Context, cfg config.ProviderConfig, modelName string) (*GrokProvider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("grok API key is required")
	}

	// Resolve the model name using configuration defaults or package constants.
	if modelName == "" {
		modelName = cfg.DefaultModel
	}
	if modelName == "" {
		modelName = grokDefaultModel
	}

	url := cfg.BaseURL
	if url == "" {
		url = grokDefaultURL
	}

	return &GrokProvider{
		apiKey: cfg.APIKey,
		model:  modelName,
		client: &http.Client{Timeout: 120 * time.Second},
		url:    url,
	}, nil
}

// Name returns the identifier for this provider.
func (p *GrokProvider) Name() string {
	return "grok"
}

// ValidateKey checks the validity of the API key by sending a minimal completion request.
func (p *GrokProvider) ValidateKey(ctx context.Context) error {
	_, err := p.Complete(ctx, CompletionRequest{
		Messages: []Message{
			{Role: "user", Content: "ping"},
		},
		MaxTokens: 1,
	})
	return err
}

// Complete sends a completion request to the Grok API and returns the response.
// It handles mapping between internal models and Grok-specific API structures.
func (p *GrokProvider) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	model := req.Model
	if model == "" {
		model = p.model
	}

	// Construct the request body using mapped messages and tools.
	body := openAIChatRequest{
		Model:     model,
		Messages:  mapOpenAIMessages(req.Messages),
		Tools:     mapOpenAITools(req.Tools),
		MaxTokens: req.MaxTokens,
		Stream:    false,
	}

	var response openAIChatResponse
	if err := p.do(ctx, body, &response); err != nil {
		return CompletionResponse{}, err
	}
	if len(response.Choices) == 0 {
		return CompletionResponse{}, fmt.Errorf("grok returned no choices")
	}

	// Map the API response back to the internal CompletionResponse structure.
	choice := response.Choices[0]
	result := CompletionResponse{
		Content:      choice.Message.Content,
		ToolCalls:    mapProviderToolCalls(choice.Message.ToolCalls),
		InputTokens:  response.Usage.PromptTokens,
		OutputTokens: response.Usage.CompletionTokens,
	}

	// Determine the stop reason based on the API's finish reason.
	if len(result.ToolCalls) > 0 || choice.FinishReason == "tool_calls" {
		result.StopReason = "tool_use"
	} else if choice.FinishReason == "length" {
		result.StopReason = "max_tokens"
	} else {
		result.StopReason = "end_turn"
	}

	return result, nil
}

// do executes an HTTP request to the Grok API and decodes the response into the output object.
func (p *GrokProvider) do(ctx context.Context, body openAIChatRequest, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	// Prepare the POST request with necessary authorization and content-type headers.
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

	// Read and limit the response size to prevent memory exhaustion.
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("grok API error %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("failed to decode grok response: %w", err)
	}

	return nil
}
