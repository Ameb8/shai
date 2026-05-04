package provider

import (
	"context"
	"fmt"
	"shai/internal/config"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiProvider struct {
	client *genai.Client
	model  string
}

func NewGeminiProvider(ctx context.Context, cfg config.ProviderConfig, modelName string) (*GeminiProvider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("gemini API key is required")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.APIKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %w", err)
	}

	if modelName == "" {
		modelName = cfg.DefaultModel
	}
	if modelName == "" {
		modelName = "gemini-2.0-flash" // Hardcoded fallback as per design
	}

	return &GeminiProvider{
		client: client,
		model:  modelName,
	}, nil
}

func (p *GeminiProvider) Name() string {
	return "gemini"
}

func (p *GeminiProvider) ValidateKey(ctx context.Context) error {
	// A simple way to validate is to list models or do a tiny completion
	iter := p.client.ListModels(ctx)
	_, err := iter.Next()
	return err
}

func (p *GeminiProvider) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	model := p.client.GenerativeModel(p.model)
	
	// Convert tools if provided
	if len(req.Tools) > 0 {
		var toolDefs []*genai.Tool
		var functions []*genai.FunctionDeclaration
		for _, t := range req.Tools {
			props := make(map[string]*genai.Schema)
			for name, pDef := range t.Parameters {
				props[name] = &genai.Schema{
					Type:        mapStringToType(pDef.Type),
					Description: pDef.Description,
				}
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
			role := "user"
			if msg.Role == "assistant" {
				role = "model"
			}
			history = append(history, &genai.Content{
				Parts: []genai.Part{genai.Text(msg.Content)},
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

	// Map stop reason
	switch candidate.FinishReason {
	case genai.FinishReasonStop:
		response.StopReason = "end_turn"
	case genai.FinishReasonSafety:
		return CompletionResponse{}, fmt.Errorf("response blocked by safety filters")
	default:
		// Check for tool use
		if len(response.ToolCalls) > 0 {
			response.StopReason = "tool_use"
		} else {
			response.StopReason = "end_turn"
		}
	}

	return response, nil
}

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
