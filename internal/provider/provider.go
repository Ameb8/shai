package provider

import "context"

type Message struct {
	Role    string // "system" | "user" | "assistant" | "tool"
	Content string
}

type ToolCall struct {
	ID   string
	Name string
	Args map[string]any
}

type ToolResult struct {
	CallID  string
	Output  string
	IsError bool
}

type CompletionRequest struct {
	Messages  []Message
	Tools     []ToolDefinition // nil if no tools needed
	Model     string
	MaxTokens int
}

type CompletionResponse struct {
	Content      string
	ToolCalls    []ToolCall
	StopReason   string // "end_turn" | "tool_use" | "max_tokens"
	InputTokens  int
	OutputTokens int
}

type ToolDefinition struct {
	Name        string
	Description string
	Parameters  map[string]ParameterDef
	Required    []string
}

type ParameterDef struct {
	Type        string   // "string" | "integer" | "boolean" | "array"
	Description string
	Enum        []string // optional
}

type Provider interface {
	Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error)
	ValidateKey(ctx context.Context) error
	Name() string
}
