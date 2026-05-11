package mockprovider

import (
	"context"
	"errors"
	"testing"

	"github.com/ameb8/shai/internal/provider"
	"github.com/stretchr/testify/assert"
)

// TestMockProvider_Complete verifies that the mock provider correctly plays back
// a multi-turn scripted conversation, including tool calls and final responses.
func TestMockProvider_Complete(t *testing.T) {
	script := []Turn{
		{
			ToolCalls: []provider.ToolCall{
				{ID: "call-1", Name: "test_tool", Args: map[string]any{"arg": 1}},
			},
			Assertions: TurnAssertions{
				MessageCount: 1,
				ContainsRole: "user",
			},
		},
		{
			FinalContent: "final answer",
			Assertions: TurnAssertions{
				LastMessageRole:     "tool",
				LastMessageContains: "result",
			},
		},
	}

	m := New(t, script)
	ctx := context.Background()

	// Execute the first turn and verify that the provider requests a tool call.
	req1 := provider.CompletionRequest{
		Messages: []provider.Message{{Role: "user", Content: "hello"}},
	}
	resp1, err := m.Complete(ctx, req1)
	assert.NoError(t, err)
	assert.Equal(t, "tool_use", resp1.StopReason)
	assert.Len(t, resp1.ToolCalls, 1)
	assert.Equal(t, "call-1", resp1.ToolCalls[0].ID)

	// Execute the second turn and verify that the provider returns the final response.
	req2 := provider.CompletionRequest{
		Messages: []provider.Message{
			{Role: "user", Content: "hello"},
			{Role: "tool", Content: "result: success"},
		},
	}
	resp2, err := m.Complete(ctx, req2)
	assert.NoError(t, err)
	assert.Equal(t, "end_turn", resp2.StopReason)
	assert.Equal(t, "final answer", resp2.Content)

	m.AssertExhausted()
	assert.Len(t, m.Calls, 2)
}

// TestMockProvider_Error verifies that the mock provider correctly returns
// scripted errors during a completion request.
func TestMockProvider_Error(t *testing.T) {
	expectedErr := errors.New("provider failure")
	script := []Turn{
		{Err: expectedErr},
	}

	m := New(t, script)
	resp, err := m.Complete(context.Background(), provider.CompletionRequest{})
	assert.ErrorIs(t, err, expectedErr)
	assert.Empty(t, resp.Content)
	m.AssertExhausted()
}
