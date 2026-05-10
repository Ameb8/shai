package agent

import (
	"context"
	"testing"

	"github.com/ameb8/shai/internal/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProvider implements the provider.Provider interface for testing.
type MockProvider struct {
	mock.Mock
}

// Complete mocks the LLM completion call.
func (m *MockProvider) Complete(ctx context.Context, req provider.CompletionRequest) (provider.CompletionResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(provider.CompletionResponse), args.Error(1)
}

// ValidateKey mocks the provider's API key validation.
func (m *MockProvider) ValidateKey(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Name returns the provider's name.
func (m *MockProvider) Name() string {
	return "mock"
}

// TestComplete_SimpleEndTurn verifies that a simple completion without tool
// use is returned immediately.
func TestComplete_SimpleEndTurn(t *testing.T) {
	mp := new(MockProvider)
	ctx := context.Background()
	messages := []provider.Message{{Role: "user", Content: "hi"}}

	expectedResp := provider.CompletionResponse{
		Content:    "hello",
		StopReason: "end_turn",
	}

	mp.On("Complete", ctx, mock.MatchedBy(func(req provider.CompletionRequest) bool {
		return len(req.Messages) == 1 && req.Messages[0].Content == "hi"
	})).Return(expectedResp, nil)

	resp, err := Complete(ctx, mp, messages)

	assert.NoError(t, err)
	assert.Equal(t, expectedResp, resp)
	mp.AssertExpectations(t)
}

// TestComplete_ExceedsMaxIterations verifies that the loop terminates and
// returns an error if the tool usage exceeds the iteration limit.
func TestComplete_ExceedsMaxIterations(t *testing.T) {
	mp := new(MockProvider)
	ctx := context.Background()
	messages := []provider.Message{{Role: "user", Content: "do something"}}

	toolCallResp := provider.CompletionResponse{
		StopReason: "tool_use",
		ToolCalls: []provider.ToolCall{
			{Name: "get_os_info"},
		},
	}

	// Always return tool_use to trigger max iterations
	mp.On("Complete", ctx, mock.Anything).Return(toolCallResp, nil)

	resp, err := Complete(ctx, mp, messages)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tool loop exceeded")
	assert.Empty(t, resp.Content)
	// Should have been called MaxToolIterations times
	mp.AssertNumberOfCalls(t, "Complete", MaxToolIterations)
}

// TestComplete_ToolDispatchError verifies that the agent handles tool dispatch
// errors correctly and continues the conversation with the error information.
func TestComplete_ToolDispatchError(t *testing.T) {
	mp := new(MockProvider)
	ctx := context.Background()
	messages := []provider.Message{{Role: "user", Content: "run invalid"}}

	toolCallResp := provider.CompletionResponse{
		StopReason: "tool_use",
		ToolCalls: []provider.ToolCall{
			{Name: "run_query", Args: map[string]any{"executable": "invalid_cmd"}},
		},
	}

	finalResp := provider.CompletionResponse{
		Content:    "failed to run",
		StopReason: "end_turn",
	}

	// First call returns tool_use
	mp.On("Complete", ctx, mock.MatchedBy(func(req provider.CompletionRequest) bool {
		return len(req.Messages) == 1
	})).Return(toolCallResp, nil).Once()

	// Second call returns final response after tool error
	mp.On("Complete", ctx, mock.MatchedBy(func(req provider.CompletionRequest) bool {
		return len(req.Messages) == 3 && req.Messages[2].Role == "tool" && req.Messages[2].ToolError == true
	})).Return(finalResp, nil).Once()

	resp, err := Complete(ctx, mp, messages)

	assert.NoError(t, err)
	assert.Equal(t, finalResp, resp)
	mp.AssertExpectations(t)
}
