package mockprovider

import (
	"context"
	"strings"
	"testing"

	"github.com/ameb8/shai/internal/provider"
)

// Turn represents one scripted response from the mock LLM.
type Turn struct {
	// ToolCalls is set when this turn should simulate the LLM requesting tool use.
	// StopReason will be set to "tool_use" in the response.
	// Must be non-empty if FinalContent is empty and Err is nil.
	ToolCalls []provider.ToolCall

	// FinalContent is set when this turn should simulate the LLM producing its
	// final answer. StopReason will be set to "end_turn" in the response.
	// Must be non-empty if ToolCalls is nil and Err is nil.
	FinalContent string

	// Assertions contains optional checks to run against the CompletionRequest
	// received for this turn, before the scripted response is returned.
	Assertions TurnAssertions

	// Err, if non-nil, causes Complete to return this error for this turn
	// instead of a scripted response. ToolCalls and FinalContent are ignored.
	Err error
}

// TurnAssertions checks that the agent constructed the request correctly before this turn fires.
type TurnAssertions struct {
	// MessageCount asserts the exact number of messages in the request.
	// Zero means no assertion.
	MessageCount int

	// ContainsRole asserts that at least one message with this role exists.
	ContainsRole string

	// LastMessageRole asserts the role of the final message in the slice.
	LastMessageRole string

	// LastMessageContains asserts a substring present in the last message's Content.
	LastMessageContains string

	// ToolResultPresent asserts that at least one message with Role "tool" exists.
	ToolResultPresent bool

	// ToolsOffered asserts that the Tools slice in the request is non-nil and non-empty.
	ToolsOffered bool
}

// MockProvider is a scripted sequence player for LLM interactions.
type MockProvider struct {
	script []Turn
	turn   int
	Calls  []provider.CompletionRequest
	t      *testing.T
}

// New creates a MockProvider that will play through the given script in order.
func New(t *testing.T, script []Turn) *MockProvider {
	for i, turn := range script {
		if turn.Err != nil {
			continue
		}
		hasToolCalls := len(turn.ToolCalls) > 0
		hasFinalContent := turn.FinalContent != ""

		if hasToolCalls && hasFinalContent {
			t.Fatalf("Turn %d: both ToolCalls and FinalContent are set", i)
		}
		if !hasToolCalls && !hasFinalContent {
			t.Fatalf("Turn %d: neither ToolCalls nor FinalContent is set", i)
		}
		if hasToolCalls {
			for j, tc := range turn.ToolCalls {
				if tc.ID == "" {
					t.Fatalf("Turn %d, ToolCall %d: empty ID", i, j)
				}
			}
		}
	}

	return &MockProvider{
		script: script,
		t:      t,
	}
}

// Complete satisfies the provider.Provider interface.
func (m *MockProvider) Complete(ctx context.Context, req provider.CompletionRequest) (provider.CompletionResponse, error) {
	m.Calls = append(m.Calls, req)

	if m.turn >= len(m.script) {
		m.t.Fatalf("Complete called more times than there are turns in the script (turn index %d, script length %d)", m.turn, len(m.script))
	}

	turn := m.script[m.turn]
	m.turn++

	if turn.Err != nil {
		return provider.CompletionResponse{}, turn.Err
	}

	m.runAssertions(req, turn.Assertions)

	if len(turn.ToolCalls) > 0 {
		return provider.CompletionResponse{
			ToolCalls:  turn.ToolCalls,
			StopReason: "tool_use",
		}, nil
	}

	return provider.CompletionResponse{
		Content:    turn.FinalContent,
		StopReason: "end_turn",
	}, nil
}

func (m *MockProvider) runAssertions(req provider.CompletionRequest, a TurnAssertions) {
	if a.MessageCount > 0 && len(req.Messages) != a.MessageCount {
		m.t.Errorf("MessageCount: expected %d, got %d", a.MessageCount, len(req.Messages))
	}

	if a.ContainsRole != "" {
		found := false
		for _, msg := range req.Messages {
			if msg.Role == a.ContainsRole {
				found = true
				break
			}
		}
		if !found {
			m.t.Errorf("ContainsRole: role %q not found in messages", a.ContainsRole)
		}
	}

	if len(req.Messages) > 0 {
		lastMsg := req.Messages[len(req.Messages)-1]
		if a.LastMessageRole != "" && lastMsg.Role != a.LastMessageRole {
			m.t.Errorf("LastMessageRole: expected %q, got %q", a.LastMessageRole, lastMsg.Role)
		}
		if a.LastMessageContains != "" && !strings.Contains(lastMsg.Content, a.LastMessageContains) {
			m.t.Errorf("LastMessageContains: content %q does not contain %q", lastMsg.Content, a.LastMessageContains)
		}
	} else if a.LastMessageRole != "" || a.LastMessageContains != "" {
		m.t.Errorf("LastMessage assertions failed: no messages in request")
	}

	if a.ToolResultPresent {
		found := false
		for _, msg := range req.Messages {
			if msg.Role == "tool" {
				found = true
				break
			}
		}
		if !found {
			m.t.Errorf("ToolResultPresent: no message with role \"tool\" found")
		}
	}

	if a.ToolsOffered && len(req.Tools) == 0 {
		m.t.Errorf("ToolsOffered: expected tools to be offered, but none were")
	}
}

// ValidateKey satisfies the provider.Provider interface.
func (m *MockProvider) ValidateKey(ctx context.Context) error {
	return nil
}

// Name satisfies the provider.Provider interface.
func (m *MockProvider) Name() string {
	return "mock"
}

// AssertExhausted asserts that all scripted turns were consumed.
func (m *MockProvider) AssertExhausted() {
	if m.turn != len(m.script) {
		m.t.Errorf("AssertExhausted: expected %d turns, but only %d were consumed", len(m.script), m.turn)
	}
}
