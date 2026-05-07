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

	return &GrokProvider{
		apiKey: cfg.APIKey,
		model:  modelName,
		client: &http.Client{Timeout: 120 * time.Second},
		url:    grokDefaultURL,
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
	body := grokChatRequest{
		Model:     model,
		Messages:  mapGrokMessages(req.Messages),
		Tools:     mapGrokTools(req.Tools),
		MaxTokens: req.MaxTokens,
		Stream:    false,
	}

	var response grokChatResponse
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
func (p *GrokProvider) do(ctx context.Context, body grokChatRequest, out any) error {
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

// grokChatRequest defines the structure for a chat completion request to the Grok API.
type grokChatRequest struct {
	Model     string        `json:"model"`
	Messages  []grokMessage `json:"messages"`
	Tools     []grokTool    `json:"tools,omitempty"`
	MaxTokens int           `json:"max_tokens,omitempty"`
	Stream    bool          `json:"stream"`
}

// grokMessage represents a single message in a Grok chat completion request.
type grokMessage struct {
	Role       string         `json:"role"`
	Content    string         `json:"content,omitempty"`
	ToolCalls  []grokToolCall `json:"tool_calls,omitempty"`
	ToolCallID string         `json:"tool_call_id,omitempty"`
}

// grokTool defines a tool that the model can call during a completion.
type grokTool struct {
	Type     string       `json:"type"`
	Function grokFunction `json:"function"`
}

// grokFunction describes a function that a tool can execute.
type grokFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  grokJSONSchema `json:"parameters"`
}

// grokJSONSchema defines the structure of parameters for a tool function.
type grokJSONSchema struct {
	Type        string                    `json:"type"`
	Properties  map[string]grokJSONSchema `json:"properties,omitempty"`
	Items       *grokJSONSchema           `json:"items,omitempty"`
	Required    []string                  `json:"required,omitempty"`
	Description string                    `json:"description,omitempty"`
	Enum        []string                  `json:"enum,omitempty"`
}

// grokToolCall represents a tool call initiated by the model.
type grokToolCall struct {
	ID       string               `json:"id,omitempty"`
	Type     string               `json:"type,omitempty"`
	Function grokToolCallFunction `json:"function"`
}

// grokToolCallFunction contains the details of a function being called.
type grokToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// grokChatResponse defines the structure for a chat completion response from the Grok API.
type grokChatResponse struct {
	Choices []struct {
		FinishReason string      `json:"finish_reason"`
		Message      grokMessage `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

// mapGrokMessages converts internal message structures to Grok-specific message structures.
func mapGrokMessages(messages []Message) []grokMessage {
	result := make([]grokMessage, 0, len(messages))
	for _, msg := range messages {
		mapped := grokMessage{
			Role:       msg.Role,
			Content:    msg.Content,
			ToolCallID: msg.ToolCallID,
		}
		// Attach tool calls if the message is from the assistant.
		if msg.Role == "assistant" {
			mapped.ToolCalls = mapGrokToolCalls(msg.ToolCalls)
		}
		result = append(result, mapped)
	}
	return result
}

// mapGrokTools converts internal tool definitions to Grok-specific tool structures.
func mapGrokTools(tools []ToolDefinition) []grokTool {
	if len(tools) == 0 {
		return nil
	}

	result := make([]grokTool, 0, len(tools))
	for _, tool := range tools {
		result = append(result, grokTool{
			Type: "function",
			Function: grokFunction{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters: grokJSONSchema{
					Type:       "object",
					Properties: mapGrokParameters(tool.Parameters),
					Required:   tool.Required,
				},
			},
		})
	}
	return result
}

// mapGrokParameters converts internal parameter definitions to Grok-specific JSON schema structures.
func mapGrokParameters(params map[string]ParameterDef) map[string]grokJSONSchema {
	result := make(map[string]grokJSONSchema, len(params))
	for name, param := range params {
		schema := grokJSONSchema{
			Type:        param.Type,
			Description: param.Description,
			Enum:        param.Enum,
		}
		// Handle array types by specifying the items schema.
		if param.Type == "array" {
			schema.Items = &grokJSONSchema{Type: "string"}
		}
		result[name] = schema
	}
	return result
}

// mapGrokToolCalls converts internal tool call structures to Grok-specific tool call structures.
func mapGrokToolCalls(calls []ToolCall) []grokToolCall {
	if len(calls) == 0 {
		return nil
	}

	result := make([]grokToolCall, 0, len(calls))
	for _, call := range calls {
		// Serialize tool arguments to JSON, defaulting to empty object on error.
		encodedArgs, err := json.Marshal(call.Args)
		if err != nil {
			encodedArgs = []byte("{}")
		}
		result = append(result, grokToolCall{
			ID:   call.ID,
			Type: "function",
			Function: grokToolCallFunction{
				Name:      call.Name,
				Arguments: string(encodedArgs),
			},
		})
	}
	return result
}

// mapProviderToolCalls converts Grok-specific tool call structures back to internal tool call structures.
func mapProviderToolCalls(calls []grokToolCall) []ToolCall {
	if len(calls) == 0 {
		return nil
	}

	result := make([]ToolCall, 0, len(calls))
	for _, call := range calls {
		args := map[string]any{}
		// Deserialize function arguments from JSON string into a map.
		if call.Function.Arguments != "" {
			if err := json.Unmarshal([]byte(call.Function.Arguments), &args); err != nil {
				args = map[string]any{}
			}
		}
		result = append(result, ToolCall{
			ID:   call.ID,
			Name: call.Function.Name,
			Args: args,
		})
	}
	return result
}
