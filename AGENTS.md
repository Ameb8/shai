# SHAI Project Instructions

## Overview
*shai* is a terminal-native AI agent that converts natural language into a single shell command, staged in the user’s prompt buffer for manual execution. 
- **Design Spec:** `design.md`
- **User Docs:** `README.md`

## Tech Stack
- **Language:** Go 1.22+
- **CLI Framework:** [Cobra](https://github.com/spf13/cobra)
- **Configuration:** [Viper](https://github.com/spf13/viper)
- **Concurrency:** Context-aware completion loops with tool-use support.

## Key Components & Directories
- `cmd/`: CLI command definitions (root, query, config).
- `internal/agent/`: The core completion loop logic.
- `internal/provider/`: LLM provider interfaces and implementations (Gemini, Mistral, etc.).
- `internal/tools/`: Tool definitions, the whitelisted shell runner, and structured tools.
- `internal/prompt/`: System prompt templates and context building.
- `internal/parser/`: Logic for extracting commands and explanations from LLM responses.
- `internal/config/`: Configuration persistence and key management.

## Core Mandates (MUST FOLLOW)

### 1. Output Contract (STRICT)
The agent's final output MUST follow this exact format:
```
<command>
# <explanation>
# <explanation continued (optional)>
# WARNING: <optional warning for destructive/privileged actions>
```
- No markdown code fences.
- No multiple options.
- Exactly one command.
- Explanation must be less than 40 words.

### 2. Security & Safety
- **Never Execute:** `shai` must NEVER execute commands directly. It only proposes them to the user's buffer.
- **Whitelisting:** All shell commands used by the agent during reasoning (via `run_query`) MUST be whitelisted in `internal/tools/runner.go`.
- **Environment Scrubbing:** Always scrub sensitive environment variables (API keys, secrets) before passing the environment to subprocesses.
- **BYOK:** Never proxy or store user API keys remotely. Use `internal/config` for local storage.

### 3. Tool Use
- Prefer whitelisted shell commands over custom tools.
- Use structured tools only when shell output is unsafe or inconsistent across OSs.
- Max tool iterations: 10.

## Development Workflows

### Testing & Quality
- **Run all tests:** `make test`
- **Format code:** `make fmt`
- **Check dependencies:** `make tidy`
- **Validation:** Always verify that new flags or config options do not break the core `query` flow.

### Activating Skills
When performing specific tasks, activate the relevant skill for expert guidance:
- `add-provider`: For adding new LLM backends.
- `add-config-setting`: For adding new persistence options.
- `add-cobra-flag`: For extending the CLI interface.
- `go-comment-style`: For maintaining Go documentation standards.

## Coding Standards
- Follow standard Go idioms and naming conventions.
- Use `context.Context` for all IO-bound operations.
- Ensure all new features are accompanied by unit tests in the same package.
