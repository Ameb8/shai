package provider

import (
	"encoding/json"
)

// openAIChatRequest defines the structure for a chat completion request to an OpenAI-compatible API.
type openAIChatRequest struct {
	Model     string          `json:"model"`
	Messages  []openAIMessage `json:"messages"`
	Tools     []openAITool    `json:"tools,omitempty"`
	MaxTokens int             `json:"max_tokens,omitempty"`
	Stream    bool            `json:"stream"`
}

// openAIMessage represents a single message in an OpenAI-compatible chat completion request.
type openAIMessage struct {
	Role       string              `json:"role"`
	Content    string              `json:"content,omitempty"`
	ToolCalls  []openAIToolCall    `json:"tool_calls,omitempty"`
	ToolCallID string              `json:"tool_call_id,omitempty"`
}

// openAITool defines a tool that the model can call during a completion.
type openAITool struct {
	Type     string           `json:"type"`
	Function openAIFunction   `json:"function"`
}

// openAIFunction describes a function that a tool can execute.
type openAIFunction struct {
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Parameters  openAIJSONSchema `json:"parameters"`
}

// openAIJSONSchema defines the structure of parameters for a tool function.
type openAIJSONSchema struct {
	Type        string                      `json:"type"`
	Properties  map[string]openAIJSONSchema `json:"properties,omitempty"`
	Items       *openAIJSONSchema           `json:"items,omitempty"`
	Required    []string                    `json:"required,omitempty"`
	Description string                      `json:"description,omitempty"`
	Enum        []string                    `json:"enum,omitempty"`
}

// openAIToolCall represents a tool call initiated by the model.
type openAIToolCall struct {
	ID       string                 `json:"id,omitempty"`
	Type     string                 `json:"type,omitempty"`
	Function openAIToolCallFunction `json:"function"`
}

// openAIToolCallFunction contains the details of a function being called.
type openAIToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// openAIChatResponse defines the structure for a chat completion response from an OpenAI-compatible API.
type openAIChatResponse struct {
	Choices []struct {
		FinishReason string        `json:"finish_reason"`
		Message      openAIMessage `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

// mapOpenAIMessages converts internal message structures to OpenAI-compatible message structures.
func mapOpenAIMessages(messages []Message) []openAIMessage {
	result := make([]openAIMessage, 0, len(messages))
	for _, msg := range messages {
		mapped := openAIMessage{
			Role:       msg.Role,
			Content:    msg.Content,
			ToolCallID: msg.ToolCallID,
		}
		if msg.Role == "assistant" {
			mapped.ToolCalls = mapOpenAIToolCalls(msg.ToolCalls)
		}
		result = append(result, mapped)
	}
	return result
}

// mapOpenAITools converts internal tool definitions to OpenAI-compatible tool structures.
func mapOpenAITools(tools []ToolDefinition) []openAITool {
	if len(tools) == 0 {
		return nil
	}

	result := make([]openAITool, 0, len(tools))
	for _, tool := range tools {
		result = append(result, openAITool{
			Type: "function",
			Function: openAIFunction{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters: openAIJSONSchema{
					Type:       "object",
					Properties: mapOpenAIParameters(tool.Parameters),
					Required:   tool.Required,
				},
			},
		})
	}
	return result
}

// mapOpenAIParameters converts internal parameter definitions to OpenAI-compatible JSON schema structures.
func mapOpenAIParameters(params map[string]ParameterDef) map[string]openAIJSONSchema {
	result := make(map[string]openAIJSONSchema, len(params))
	for name, param := range params {
		schema := openAIJSONSchema{
			Type:        param.Type,
			Description: param.Description,
			Enum:        param.Enum,
		}
		if param.Type == "array" {
			schema.Items = &openAIJSONSchema{Type: "string"}
		}
		result[name] = schema
	}
	return result
}

// mapOpenAIToolCalls converts internal tool call structures to OpenAI-compatible tool call structures.
func mapOpenAIToolCalls(calls []ToolCall) []openAIToolCall {
	if len(calls) == 0 {
		return nil
	}

	result := make([]openAIToolCall, 0, len(calls))
	for _, call := range calls {
		encodedArgs, err := json.Marshal(call.Args)
		if err != nil {
			encodedArgs = []byte("{}")
		}
		result = append(result, openAIToolCall{
			ID:   call.ID,
			Type: "function",
			Function: openAIToolCallFunction{
				Name:      call.Name,
				Arguments: string(encodedArgs),
			},
		})
	}
	return result
}

// mapProviderToolCalls converts OpenAI-compatible tool call structures back to internal tool call structures.
func mapProviderToolCalls(calls []openAIToolCall) []ToolCall {
	if len(calls) == 0 {
		return nil
	}

	result := make([]ToolCall, 0, len(calls))
	for _, call := range calls {
		args := map[string]any{}
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
