package provider

import "context"

// Message represents a single turn in a conversation with an LLM.
type Message struct {
	// Role identifies the sender of the message (e.g., "system", "user", "assistant", "tool").
	Role string
	// Content is the text payload of the message.
	Content string
	// ToolCalls contains requests for tool execution, typically sent by the assistant.
	ToolCalls []ToolCall
	// ToolCallID is the unique identifier for a specific tool execution request.
	ToolCallID string
	// ToolName is the name of the tool associated with a tool result.
	ToolName string
	// ToolError indicates if the tool execution failed.
	ToolError bool
}

// ToolCall represents a request from the model to execute a specific tool.
type ToolCall struct {
	ID   string
	Name string
	Args map[string]any
}

// ToolResult contains the output of a tool execution to be returned to the model.
type ToolResult struct {
	CallID  string
	Output  string
	IsError bool
}

// CompletionRequest defines the input parameters for a model completion.
type CompletionRequest struct {
	Messages  []Message
	Tools     []ToolDefinition // Tools is nil if no tools are available.
	Model     string
	MaxTokens int
}

// CompletionResponse holds the output and metadata from a model completion.
type CompletionResponse struct {
	Content      string
	ToolCalls    []ToolCall
	StopReason   string // StopReason can be "end_turn", "tool_use", or "max_tokens".
	InputTokens  int
	OutputTokens int
}

// ToolDefinition defines a tool that the model can choose to call.
type ToolDefinition struct {
	Name        string
	Description string
	Parameters  map[string]ParameterDef
	Required    []string
}

// ParameterDef describes a single parameter for a tool, following JSON Schema structure.
type ParameterDef struct {
	Type        string // Type can be "string", "integer", "boolean", or "array".
	Description string
	Enum        []string // Enum is optional and restricts the parameter to specific values.
}

// Provider defines the interface for interacting with different LLM backends.
type Provider interface {
	// Complete sends a completion request to the provider and returns the model's response.
	Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error)
	// ValidateKey checks if the configured API key is valid for the provider.
	ValidateKey(ctx context.Context) error
	// Name returns the unique identifier for the provider.
	Name() string
}
