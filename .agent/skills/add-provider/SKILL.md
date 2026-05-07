---
name: add-provider
description: Add support for a new LLM provider to the shai project. Use when you need to implement the Provider interface, register a new backend in the provider registry, and add corresponding unit tests.
---

# Add Provider

This skill guides you through adding a new LLM provider to `shai`.

## Workflow

1.  **Analyze the Provider API**: Understand the request/response format of the new LLM provider (e.g., Anthropic, OpenAI).
2.  **Implement the Interface**: Create a new file in `internal/provider/` and implement the `Provider` interface. See [implementation-guide.md](references/implementation-guide.md) for details.
3.  **Register the Provider**: Add the provider to the `NewProvider` factory in `internal/provider/registry.go`.
4.  **Verify**: Add unit tests in a new `_test.go` file in `internal/provider/`.
