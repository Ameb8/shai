package integration

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/ameb8/shai/internal/agent"
	"github.com/ameb8/shai/internal/parser"
	"github.com/ameb8/shai/internal/provider"
	"github.com/ameb8/shai/internal/testutil/mockprovider"
)

// TestAgentComplete_DirectAnswer verifies that the agent correctly handles a direct
// response from the LLM without any tool calls.
func TestAgentComplete_DirectAnswer(t *testing.T) {
	finalContent := "echo hello\n# print hello"
	// Configure a mock provider with a single direct answer turn.
	script := []mockprovider.Turn{
		{
			FinalContent: finalContent,
		},
	}
	mock := mockprovider.New(t, script)
	messages := buildMessages(t, "say hello")

	// Execute the completion loop.
	resp, err := agent.Complete(context.Background(), mock, messages)

	if err != nil {
		t.Fatalf("agent.Complete failed: %v", err)
	}
	// Validate the stop reason and final content.
	if resp.StopReason != "end_turn" {
		t.Errorf("expected StopReason \"end_turn\", got %q", resp.StopReason)
	}
	if resp.Content != finalContent {
		t.Errorf("expected Content %q, got %q", finalContent, resp.Content)
	}
	mock.AssertExhausted()
}

// TestAgentComplete_SingleToolCall ensures the agent can successfully execute a
// single tool and incorporate its result into the final response.
func TestAgentComplete_SingleToolCall(t *testing.T) {
	// Define a sequence where the agent first calls a tool, then provides the final answer.
	script := []mockprovider.Turn{
		{
			ToolCalls: []provider.ToolCall{
				{
					ID:   "call_1",
					Name: "run_query",
					Args: map[string]any{
						"executable": "echo",
						"args":       []any{"hello"},
					},
				},
			},
		},
		{
			FinalContent: "echo hello\n# command output was hello",
			Assertions: mockprovider.TurnAssertions{
				ToolResultPresent: true,
				LastMessageRole:   "tool",
			},
		},
	}
	mock := mockprovider.New(t, script)
	messages := buildMessages(t, "run echo hello")

	resp, err := agent.Complete(context.Background(), mock, messages)

	if err != nil {
		t.Fatalf("agent.Complete failed: %v", err)
	}
	// Ensure the loop terminated correctly after tool execution.
	if resp.StopReason != "end_turn" {
		t.Errorf("expected StopReason \"end_turn\", got %q", resp.StopReason)
	}
	mock.AssertExhausted()
	if len(mock.Calls) != 2 {
		t.Errorf("expected 2 calls to provider, got %d", len(mock.Calls))
	}
}

// TestAgentComplete_MultiTurnToolCalls validates that the agent can handle multiple
// sequential tool calls before reaching a final conclusion.
func TestAgentComplete_MultiTurnToolCalls(t *testing.T) {
	// Simulate multiple turns of tool usage followed by a terminal response.
	script := []mockprovider.Turn{
		{
			ToolCalls: []provider.ToolCall{
				{
					ID:   "call_1",
					Name: "run_query",
					Args: map[string]any{
						"executable": "echo",
						"args":       []any{"first"},
					},
				},
			},
		},
		{
			ToolCalls: []provider.ToolCall{
				{
					ID:   "call_2",
					Name: "run_query",
					Args: map[string]any{
						"executable": "echo",
						"args":       []any{"second"},
					},
				},
			},
			Assertions: mockprovider.TurnAssertions{
				ToolResultPresent: true,
			},
		},
		{
			FinalContent: "echo done\n# all turns completed",
			Assertions: mockprovider.TurnAssertions{
				ToolResultPresent: true,
			},
		},
	}
	mock := mockprovider.New(t, script)
	messages := buildMessages(t, "run two echoes")

	_, err := agent.Complete(context.Background(), mock, messages)

	if err != nil {
		t.Fatalf("agent.Complete failed: %v", err)
	}
	// Confirm all expected provider turns were executed.
	if len(mock.Calls) != 3 {
		t.Errorf("expected 3 calls to provider, got %d", len(mock.Calls))
	}
	mock.AssertExhausted()
}

// TestAgentComplete_ToolIterationLimit verifies that the agent terminates with an
// error if it exceeds the maximum allowed tool execution turns.
func TestAgentComplete_ToolIterationLimit(t *testing.T) {
	// Construct a script that triggers the maximum number of tool calls allowed.
	var script []mockprovider.Turn
	for i := 0; i < 10; i++ {
		script = append(script, mockprovider.Turn{
			ToolCalls: []provider.ToolCall{
				{
					ID:   "call",
					Name: "run_query",
					Args: map[string]any{
						"executable": "echo",
						"args":       []any{"loop"},
					},
				},
			},
		})
	}

	mock := mockprovider.New(t, script)
	messages := buildMessages(t, "loop forever")

	_, err := agent.Complete(context.Background(), mock, messages)

	// Verify that the iteration limit was enforced and reported.
	if err == nil {
		t.Fatal("expected error due to tool loop limit, got nil")
	}
	if !strings.Contains(err.Error(), "tool loop exceeded") {
		t.Errorf("expected error to contain \"tool loop exceeded\", got %q", err.Error())
	}
	if len(mock.Calls) != 10 {
		t.Errorf("expected 10 calls to provider, got %d", len(mock.Calls))
	}
	mock.AssertExhausted()
}

// TestAgentComplete_ProviderError ensures that errors returned by the LLM provider
// are correctly propagated to the caller.
func TestAgentComplete_ProviderError(t *testing.T) {
	sentinelErr := errors.New("provider failure")
	// Setup a turn that immediately returns an error.
	script := []mockprovider.Turn{
		{
			Err: sentinelErr,
		},
	}
	mock := mockprovider.New(t, script)
	messages := buildMessages(t, "trigger error")

	_, err := agent.Complete(context.Background(), mock, messages)

	// Confirm the error is the expected sentinel.
	if err == nil {
		t.Fatal("expected error from provider, got nil")
	}
	if !errors.Is(err, sentinelErr) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

// TestAgentComplete_MalformedResponse checks how the agent handles LLM output that
// does not strictly adhere to the expected format.
func TestAgentComplete_MalformedResponse(t *testing.T) {
	// Provide output that lacks the standard # explanation markers.
	malformedContent := "this is not a valid output contract"
	script := []mockprovider.Turn{
		{
			FinalContent: malformedContent,
		},
	}
	mock := mockprovider.New(t, script)
	messages := buildMessages(t, "give bad response")

	resp, err := agent.Complete(context.Background(), mock, messages)
	if err != nil {
		t.Fatalf("agent.Complete failed: %v", err)
	}

	// Parse the response to verify fallback behavior.
	parsed, err := parser.Parse(resp.Content)
	if err != nil {
		t.Fatalf("parser.Parse failed: %v", err)
	}

	// Verify that malformed content is treated as a command without an explanation.
	if parsed.Command != malformedContent {
		t.Errorf("expected Command %q, got %q", malformedContent, parsed.Command)
	}
	if len(parsed.Explanation) != 0 {
		t.Errorf("expected 0 explanation lines, got %d", len(parsed.Explanation))
	}
}
