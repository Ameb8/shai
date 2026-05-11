# Mock Provider Design: In-Process Go Integration Tests

## Why

The `agent.Complete` loop already accepts a `provider.Provider` interface, making it naturally testable without real LLM calls. A mock provider implementation allows integration tests to:

- Script deterministic, multi-turn tool-use sequences
- Assert that the agent feeds correct message history back to the provider each turn
- Verify the tool iteration limit is enforced
- Test parser and output contract compliance against controlled responses
- Run fast and offline with no API keys or external dependencies

The mock is not a general-purpose stub — it is a **scripted sequence player**. Each test scenario defines exactly how many turns the LLM takes, what it does each turn (tool call or final answer), and optionally what it asserts about the messages it receives.

---

## The Injection Gap

`agent.Complete(ctx, p, messages)` already accepts a `Provider` interface — no changes needed there.

However, `cmd/query.go`'s `runQuery` constructs the provider internally via `provider.NewProvider()` and passes it directly to `agent.Complete`. There is currently no injection point for tests to substitute a mock provider without going through the real registry.

To enable in-process CLI-level integration tests (testing the full `runQuery` path), a **provider injection seam** must be added to `cmd/query.go`. The implementing agent should determine the least-invasive way to do this given the existing cobra command structure — for example, a package-level override variable, a functional option on the command, or a constructor that accepts an optional provider. The seam must not affect production behavior when no override is set.

Tests that only exercise `agent.Complete` directly (not the full cobra path) do not require this change.

---

## Package Location

```
internal/testutil/mockprovider/
    mock_provider.go
    mock_provider_test.go   // unit tests for the mock itself
```

Placing it under `internal/testutil/` keeps it clearly non-production. The `mockprovider` sub-package avoids polluting a flat `testutil` namespace as more test utilities are added.

---

## Core Types

### `Turn`

Represents one scripted response from the mock LLM. Exactly one of `ToolCalls` or `FinalContent` must be set per turn.

```go
type Turn struct {
    // ToolCalls is set when this turn should simulate the LLM requesting tool use.
    // StopReason will be set to "tool_use" in the response.
    // Must be non-empty if FinalContent is empty.
    ToolCalls []provider.ToolCall

    // FinalContent is set when this turn should simulate the LLM producing its
    // final answer. StopReason will be set to "end_turn" in the response.
    // Must be non-empty if ToolCalls is nil.
    FinalContent string

    // Assertions contains optional checks to run against the CompletionRequest
    // received for this turn, before the scripted response is returned.
    Assertions TurnAssertions
}
```

### `TurnAssertions`

Checks that the agent constructed the request correctly before this turn fires. All fields are optional — a zero-value `TurnAssertions` performs no checks.

```go
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
    // Use this to confirm the agent is feeding tool results back correctly.
    ToolResultPresent bool

    // ToolsOffered asserts that the Tools slice in the request is non-nil and non-empty.
    ToolsOffered bool
}
```

### `MockProvider`

```go
type MockProvider struct {
    // script is the ordered sequence of turns to play through.
    script []Turn

    // turn is the index of the next turn to execute.
    turn int

    // Calls captures every CompletionRequest received, in order.
    // Available for post-hoc assertions after the agent loop completes.
    Calls []provider.CompletionRequest

    // t is used to report assertion failures and unexpected calls.
    t *testing.T
}
```

---

## Constructor

```go
// New creates a MockProvider that will play through the given script in order.
// Each call to Complete consumes the next Turn. The test will fail immediately
// if Complete is called more times than there are turns in the script.
func New(t *testing.T, script []Turn) *MockProvider
```

---

## `Complete` Behaviour

The `Complete` method must:

1. Record the incoming `CompletionRequest` by appending it to `Calls`.
2. Fail the test immediately (via `t.Fatalf`) if called when no turns remain in the script.
3. Advance `turn` and retrieve the current `Turn`.
4. Run all non-zero `TurnAssertions` against the received request, reporting failures via `t.Errorf` (non-fatal, so all assertion failures in a turn are visible).
5. Construct and return a `CompletionResponse`:
   - If `Turn.ToolCalls` is set: return `CompletionResponse{ToolCalls: turn.ToolCalls, StopReason: "tool_use"}`.
   - If `Turn.FinalContent` is set: return `CompletionResponse{Content: turn.FinalContent, StopReason: "end_turn"}`.
6. Token counts (`InputTokens`, `OutputTokens`) should be returned as zero. The mock does not simulate token counting.

`ValidateKey` must return `nil` unconditionally.

`Name` must return the string `"mock"`.

---

## `AssertExhausted`

```go
// AssertExhausted asserts that all scripted turns were consumed.
// Call this at the end of each test to catch cases where the agent
// returned early without completing the expected number of LLM turns.
func (m *MockProvider) AssertExhausted()
```

Fails the test via `t.Errorf` if `m.turn != len(m.script)`.

---

## Tool Call ID Generation

Real providers return a unique `ID` per `ToolCall`. When a test scripts a `Turn` with `ToolCalls`, the test author must supply IDs on the `ToolCall` structs — the mock does not generate them. This keeps the mock simple and makes test data explicit. IDs like `"call-1"`, `"call-2"` are sufficient.

---

## Validation at Construction

`New` should fail the test immediately (via `t.Fatal`) if any `Turn` in the script is invalid:

- A turn where both `ToolCalls` and `FinalContent` are set.
- A turn where neither `ToolCalls` nor `FinalContent` is set.
- A turn with `ToolCalls` where any `ToolCall` has an empty `ID`.

Catching these at construction surfaces misconfigured test scripts before the agent loop runs.

---

## Error Simulation

To test agent error handling, add an optional error field to `Turn`:

```go
type Turn struct {
    // ... fields above ...

    // Err, if non-nil, causes Complete to return this error for this turn
    // instead of a scripted response. ToolCalls and FinalContent are ignored.
    Err error
}
```

This allows testing that the agent surfaces provider errors correctly without needing a real failure condition.

---

## Usage Pattern

Each integration test constructs a script that represents one coherent scenario, then drives `agent.Complete` (or the full cobra path once the injection seam exists) directly:

```go
func TestAgentSingleToolCallThenAnswer(t *testing.T) {
    mock := mockprovider.New(t, []mockprovider.Turn{
        {
            // Turn 1: LLM requests a tool call
            ToolCalls: []provider.ToolCall{
                {ID: "call-1", Name: "run_query", Args: map[string]any{
                    "executable": "uname",
                    "args":       []any{"-a"},
                }},
            },
            Assertions: mockprovider.TurnAssertions{
                ToolsOffered:    true,
                ContainsRole:    "user",
                LastMessageRole: "user",
            },
        },
        {
            // Turn 2: LLM sees tool result and produces final answer
            FinalContent: "uname -a\n# print system information",
            Assertions: mockprovider.TurnAssertions{
                ToolResultPresent:   true,
                LastMessageRole:     "tool",
            },
        },
    })

    messages := []provider.Message{
        {Role: "system", Content: "you are shai"},
        {Role: "user", Content: "what OS am I running"},
    }

    resp, err := agent.Complete(context.Background(), mock, messages)

    require.NoError(t, err)
    assert.Equal(t, "end_turn", resp.StopReason)
    assert.Contains(t, resp.Content, "uname")
    mock.AssertExhausted()
}
```

---

## What the Mock Does Not Own

The following are explicitly out of scope for the mock provider:

- **Tool dispatch**: `tools.Dispatch` runs real whitelisted commands during the agent loop. Integration tests that want to avoid real shell execution should either use a query that doesn't trigger tool use, or introduce a separate tool runner interface — that is a separate design concern.
- **Prompt construction**: System prompt building is tested separately. The mock receives whatever messages the agent passes it; it does not validate prompt content beyond the `TurnAssertions` fields explicitly set by the test.
- **Parser correctness**: Parser edge cases are unit-tested directly against `parser.Parse`. The mock's `FinalContent` values should be valid output-contract strings unless a test is specifically exercising parser error handling.