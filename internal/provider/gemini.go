package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ameb8/shai/internal/config"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GeminiProvider implements the provider.Provider interface using the Google
// Generative AI (Gemini) API.
type GeminiProvider struct {
	client *genai.Client
	model  string
	url    string
}

// NewGeminiProvider initializes a new Gemini provider with the given configuration
// and model name.
func NewGeminiProvider(ctx context.Context, cfg config.ProviderConfig, modelName string) (*GeminiProvider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("gemini API key is required")
	}

	opts := []option.ClientOption{option.WithAPIKey(cfg.APIKey)}
	url := cfg.BaseURL
	if url != "" {
		opts = append(opts, option.WithEndpoint(url))
	}

	client, err := genai.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %w", err)
	}

	if modelName == "" {
		modelName = cfg.DefaultModel
	}
	if modelName == "" {
		modelName = "gemini-2.0-flash" // Hardcoded fallback
	}

	return &GeminiProvider{
		client: client,
		model:  modelName,
		url:    url,
	}, nil
}

// Name returns the provider identifier "gemini".
func (p *GeminiProvider) Name() string {
	return "gemini"
}

// ValidateKey checks if the provided API key is valid by attempting to list
// available models.
func (p *GeminiProvider) ValidateKey(ctx context.Context) error {
	// A simple way to validate is to list models or do a tiny completion
	iter := p.client.ListModels(ctx)
	_, err := iter.Next()
	return err
}

// Complete sends a completion request to the Gemini API, handling tool
// definitions and chat history.
func (p *GeminiProvider) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	model := p.client.GenerativeModel(p.model)

	// Convert tools if provided
	if len(req.Tools) > 0 {
		var toolDefs []*genai.Tool
		var functions []*genai.FunctionDeclaration
		for _, t := range req.Tools {
			props := make(map[string]*genai.Schema)
			for name, pDef := range t.Parameters {
				props[name] = mapParameterToSchema(pDef)
				if len(pDef.Enum) > 0 {
					props[name].Enum = pDef.Enum
				}
			}
			functions = append(functions, &genai.FunctionDeclaration{
				Name:        t.Name,
				Description: t.Description,
				Parameters: &genai.Schema{
					Type:       genai.TypeObject,
					Properties: props,
					Required:   t.Required,
				},
			})
		}
		toolDefs = append(toolDefs, &genai.Tool{
			FunctionDeclarations: functions,
		})
		model.Tools = toolDefs
	}

	cs := model.StartChat()

	// Separate system instructions from chat history
	var systemInstructions []genai.Part
	var history []*genai.Content

	for _, msg := range req.Messages {
		if msg.Role == "system" {
			systemInstructions = append(systemInstructions, genai.Text(msg.Content))
		} else {
			role := mapMessageRole(msg.Role)
			parts := messageParts(msg)
			history = append(history, &genai.Content{
				Parts: parts,
				Role:  role,
			})
		}
	}

	if len(systemInstructions) > 0 {
		model.SystemInstruction = &genai.Content{
			Parts: systemInstructions,
		}
	}

	// The last message is the current prompt
	if len(history) == 0 {
		return CompletionResponse{}, fmt.Errorf("no user messages provided")
	}

	lastMsg := history[len(history)-1]
	cs.History = history[:len(history)-1]

	resp, err := cs.SendMessage(ctx, lastMsg.Parts...)
	if err != nil {
		return CompletionResponse{}, err
	}

	if len(resp.Candidates) == 0 {
		return CompletionResponse{}, fmt.Errorf("no candidates in response")
	}

	candidate := resp.Candidates[0]
	var response CompletionResponse

	for _, part := range candidate.Content.Parts {
		if text, ok := part.(genai.Text); ok {
			response.Content += string(text)
		} else if fnCall, ok := part.(genai.FunctionCall); ok {
			response.ToolCalls = append(response.ToolCalls, ToolCall{
				ID:   "", // Gemini doesn't use IDs for function calls in the same way
				Name: fnCall.Name,
				Args: fnCall.Args,
			})
		}
	}

	if len(response.ToolCalls) > 0 {
		response.StopReason = "tool_use"
		return response, nil
	}

	switch candidate.FinishReason {
	case genai.FinishReasonSafety:
		return CompletionResponse{}, fmt.Errorf("response blocked by safety filters")
	case genai.FinishReasonStop:
		response.StopReason = "end_turn"
	default:
		response.StopReason = "end_turn"
	}

	return response, nil
}

// mapMessageRole converts internal message roles to Gemini-specific roles.
func mapMessageRole(role string) string {
	switch role {
	case "assistant":
		return "model"
	case "tool":
		return "function"
	default:
		return "user"
	}
}

// messageParts converts an internal Message into a slice of Gemini content parts.
func messageParts(msg Message) []genai.Part {
	var parts []genai.Part
	if msg.Content != "" && msg.Role != "tool" {
		parts = append(parts, genai.Text(msg.Content))
	}

	for _, call := range msg.ToolCalls {
		parts = append(parts, genai.FunctionCall{
			Name: call.Name,
			Args: call.Args,
		})
	}

	if msg.Role == "tool" {
		response := map[string]any{}
		if err := json.Unmarshal([]byte(msg.Content), &response); err != nil {
			response["result"] = msg.Content
		}
		if msg.ToolError {
			response["is_error"] = true
		}
		parts = append(parts, genai.FunctionResponse{
			Name:     msg.ToolName,
			Response: response,
		})
	}

	if len(parts) == 0 {
		parts = append(parts, genai.Text(""))
	}

	return parts
}

// mapParameterToSchema converts an internal parameter definition to a Gemini schema.
func mapParameterToSchema(pDef ParameterDef) *genai.Schema {
	schema := &genai.Schema{
		Type:        mapStringToType(pDef.Type),
		Description: pDef.Description,
	}
	if strings.EqualFold(pDef.Type, "array") {
		schema.Items = &genai.Schema{Type: genai.TypeString}
	}
	return schema
}

// mapStringToType converts a string type name to a genai.Type.
func mapStringToType(s string) genai.Type {
	switch strings.ToLower(s) {
	case "string":
		return genai.TypeString
	case "integer":
		return genai.TypeInteger
	case "boolean":
		return genai.TypeBoolean
	case "array":
		return genai.TypeArray
	case "object":
		return genai.TypeObject
	default:
		return genai.TypeString
	}
}
