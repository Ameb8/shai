# Integration Test Design: shai In-Process Go Tests

## Overview

This document specifies the integration test setup for shai. Tests are written in Go and run in-process — no subprocess spawning, no real LLM calls. The mock provider (`internal/testutil/mockprovider`) is already implemented and is the foundation for all tests here.

Two test paths exist:

- **Path 1 — `agent.Complete` directly:** Calls the agent loop with a mock provider and hand-constructed messages. Used for the majority of scenarios.
- **Path 2 — Full cobra path:** Executes the root cobra command in-process with captured stdout/stderr. Used for CLI-level concerns only.

All tests live in `internal/integration/`. Real tool dispatch (`tools.Dispatch`) executes during tool-use tests — this is intentional, as whitelisted commands are safe and read-only.

---

## Structural Changes Required

### 1. `cmd` package: `rootCmd` must be constructable without `init()` side effects

**Problem:** `rootCmd` is a package-level var initialized via `init()`. `initConfig()` is registered via `cobra.OnInitialize`, which calls `config.Load()` — reading from `~/.config/shai/config.toml` on disk. Tests cannot control this path, and it will `os.Exit(1)` if config loading fails in certain environments.

`cfg` is also a package-level var set by `initConfig()`. Path 2 tests need `cfg` to be non-nil and valid without reading real disk config.

**Fix:** Add a `NewRootCmd(cfgOverride *config.Config) *cobra.Command` constructor that:
- Builds and returns a fully configured cobra command tree (same flags, same subcommands)
- Accepts an optional `cfgOverride *config.Config` — when non-nil, skips `initConfig()` and uses the provided config directly
- When `cfgOverride` is nil, behaves exactly as today (registers `cobra.OnInitialize(initConfig)`)

The existing `rootCmd` package-level var and `Execute()` function remain untouched for production use. `NewRootCmd` is additive.

The command tree wiring (subcommands, flags) that currently lives in `init()` functions across `cmd/root.go` and `cmd/config.go` should be extracted into the constructor so both the package-level var and `NewRootCmd` share the same setup logic without duplication.

**Production code change surface:** `cmd/root.go` only. `cmd/config.go` init wiring may need minor extraction.

---

### 2. `cmd` package: `ProviderOverride` already exists — no changes needed

`ProviderOverride` is already implemented and the `runQuery` nil-check is correct. Path 2 tests set this before executing the command and clear it via `t.Cleanup`. No parallel safety concern as long as integration tests do not call `t.Parallel()` (they should not — see below).

---

## Test Package Structure

```
internal/integration/
    helpers_test.go         // shared helpers: RunCLI, buildMessages, minimalConfig
    agent_loop_test.go      // Path 1: agent.Complete scenarios
    cli_test.go             // Path 2: full cobra path scenarios
```

All files use `package integration` (not `package integration_test`) so they can access test helpers without export gymnastics. Do not call `t.Parallel()` anywhere in this package.

---

## Helpers (`helpers_test.go`)

### `buildMessages`

Constructs a standard message slice for Path 1 tests. Always passes `"/bin/bash"` as the shell override to make the system prompt deterministic across environments.

```go
func buildMessages(t *testing.T, query string) []provider.Message
```

Calls `prompt.BuildSystemPrompt("/bin/bash")` internally. Fails the test immediately on error.

### `minimalConfig`

Returns a `*config.Config` with empty-but-initialized maps, suitable for Path 2 tests where `ProviderOverride` is set and the real provider registry is never reached.

```go
func minimalConfig() *config.Config
```

Returns a config with `Active.Provider` set to `"mock"`, and initialized (non-nil) `Providers` and `Models` maps. No real API keys or provider entries needed.

### `RunCLI`

Executes the cobra command tree in-process and captures stdout/stderr as strings. Uses `NewRootCmd(minimalConfig())` so no disk config is read.

```go
func RunCLI(t *testing.T, args ...string) (stdout, stderr string, err error)
```

Implementation:
1. Create `outBuf, errBuf := &bytes.Buffer{}, &bytes.Buffer{}`
2. Call `cmd.NewRootCmd(minimalConfig())`
3. Call `root.SetOut(outBuf)` and `root.SetErr(errBuf)`
4. Call `root.SetArgs(args)`
5. Call `err = root.Execute()`
6. Return `outBuf.String(), errBuf.String(), err`

**Important:** `cmd.ProviderOverride` must be set before calling `RunCLI` in Path 2 tests, and cleared via `t.Cleanup`.

---

## Path 1 Test Scenarios (`agent_loop_test.go`)

Each test calls `agent.Complete(context.Background(), mock, buildMessages(t, query))` directly.

### Scenario 1: Direct answer, no tool use

Mock script: one turn, `FinalContent` set to a valid output-contract string.

Assert:
- `err` is nil
- `resp.StopReason == "end_turn"`
- `resp.Content` matches the scripted final content
- `mock.AssertExhausted()`

### Scenario 2: Single tool call, then final answer

Mock script:
- Turn 1: `ToolCalls` with one `run_query` call using a whitelisted command (e.g. `{"executable": "uname", "args": ["-a"]}`)
- Turn 2: `FinalContent` with valid output-contract string; `Assertions.ToolResultPresent = true`, `Assertions.LastMessageRole = "tool"`

Assert:
- `err` is nil
- `resp.StopReason == "end_turn"`
- `mock.AssertExhausted()`
- `len(mock.Calls) == 2`

This test exercises real tool dispatch — `uname -a` will actually execute. No assertion is made on its output.

### Scenario 3: Multi-turn tool calls, then final answer

Mock script:
- Turn 1: `ToolCalls` with `{"executable": "uname", "args": ["-a"]}`
- Turn 2: `ToolCalls` with `{"executable": "echo", "args": ["hello"]}`; `Assertions.ToolResultPresent = true`
- Turn 3: `FinalContent`; `Assertions.ToolResultPresent = true`

Assert:
- `err` is nil
- `len(mock.Calls) == 3`
- `mock.AssertExhausted()`

### Scenario 4: Tool iteration limit enforced

Mock script: exactly `agent.MaxToolIterations` turns, all with `ToolCalls`. The agent must exhaust the limit and return an error — the mock should never be called beyond the limit, so script exactly 10 turns (not 11).

Assert:
- `err != nil`
- Error message contains `"tool loop exceeded"`
- `len(mock.Calls) == agent.MaxToolIterations`

Note: `mock.AssertExhausted()` is correct here too since all 10 scripted turns will be consumed.

### Scenario 5: Provider error propagated

Mock script: one turn with `Err` set to a sentinel error value.

Assert:
- `err != nil`
- `errors.Is(err, sentinelErr)` — confirm the error is not wrapped in a way that loses identity

### Scenario 6: Malformed LLM response reaches parser

Mock script: one turn with `FinalContent` set to a string that violates the output contract (e.g. empty string, or markdown-fenced content).

Call `parser.Parse(resp.Content)` explicitly after `agent.Complete` returns — this mirrors what `runQuery` does and tests that the parser correctly errors or handles the case.

Assert based on what `parser.Parse` is specified to do with invalid input — check `internal/parser` for its error behaviour and assert accordingly.

---

## Path 2 Test Scenarios (`cli_test.go`)

Each test sets `cmd.ProviderOverride` and calls `RunCLI`.

```go
cmd.ProviderOverride = mockprovider.New(t, script)
t.Cleanup(func() { cmd.ProviderOverride = nil })
stdout, stderr, err := RunCLI(t, "query", "--shell", "/bin/bash", "find all go files")
```

### Scenario 7: Output contract on stdout

Mock script: one turn, `FinalContent = "find . -name '*.go'\n# find all Go source files"`.

Assert:
- `err` is nil
- `stdout` contains `"> find . -name '*.go'"` (the command line with `>` prefix)
- `stdout` contains `"# find all Go source files"` (the explanation)

### Scenario 8: `--no-explain` suppresses explanation

Same mock as Scenario 7. Add `"--no-explain"` to args.

Assert:
- `stdout` contains `"> find . -name '*.go'"`
- `stdout` does not contain `"# find all Go source files"`

### Scenario 9: `--dry-run` prints command, does not write file

Mock script: one turn with valid `FinalContent`.

Call with `"--dry-run"` and `"--cmd-file", "/tmp/shai-test-output"`.

Assert:
- `stdout` contains the command with `>` prefix
- `/tmp/shai-test-output` does not exist (use `t.Cleanup` to remove if it does, and assert before cleanup)

### Scenario 10: `--cmd-file` writes command to file

Mock script: one turn, `FinalContent = "echo hello\n# print hello"`.

Use `t.TempDir()` for the file path. Call with `"--cmd-file", filepath.Join(tmpDir, "cmd.txt")`.

Assert:
- `err` is nil
- File exists and contains exactly `"echo hello"` (command only, no explanation)
- `stdout` does not contain `"> echo hello"` (not printed to stdout when cmd-file is set)

---

## What Is Explicitly Out Of Scope

- Testing `config` subcommands — those are unit-testable independently and don't require the integration harness
- Asserting on exact tool output content — tool results are environment-dependent
- Asserting on system prompt content beyond what `buildMessages` guarantees
- Any test that requires a real provider API key

---

## Environment and CI Notes

- Always pass `--shell /bin/bash` (or `"/bin/bash"` to `BuildSystemPrompt`) in all tests to make the system prompt deterministic
- Do not call `t.Parallel()` anywhere in `internal/integration/`
- Tests that invoke real tool dispatch (Scenarios 2, 3, 4) require whitelisted executables (`uname`, `echo`) to be present — these are available on all Linux and macOS environments including GitHub Actions standard runners
- No API keys or network access required