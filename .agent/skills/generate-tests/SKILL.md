---
name: generate-tests
description: Generates unit tests for Go code using testify and table-driven testing patterns. Use when adding test coverage for new or existing functions, or when tests need to be updated to match the project's standards.
---

# Generate Tests

This skill guides the creation of unit tests in the `shai` project. We use `testify` for assertions and mocking, and prefer table-driven tests for clarity and maintainability.

## Workflow

1.  **Analyze the Target Code**: Understand the inputs, outputs, and edge cases of the function(s) you are testing.
2.  **Identify Dependencies**: Determine if any external dependencies (e.g., LLM providers, file system, network) need to be mocked.
3.  **Create/Update Test File**:
    -   Tests should be in the same package as the code they test.
    -   Filename should end in `_test.go`.
4.  **Implement Table-Driven Tests**:
    -   Use a slice of structs to define test cases.
    -   Include a `name` field for each case.
    -   Use `t.Run()` to execute each case.
5.  **Use Testify Assertions**: Prefer `assert.NoError`, `assert.Equal`, `assert.Contains`, etc., over manual `if err != nil` checks.
6.  **Validate**: Run the tests using `go test ./...` or `make test` to ensure they pass and provide adequate coverage.

## References & Assets

-   **Testing Patterns**: See [testing-patterns.md](references/testing-patterns.md) for concrete examples of table-driven tests and mocking.
-   **Test Template**: Use [test-template.go](assets/test-template.go) as a starting point for new test files.

## Mocking Guidelines

-   If testing code that interacts with the `provider.Provider` interface, use `MockProvider` (see `internal/agent/loop_test.go` for an example).
-   Always use `AssertExpectations(t)` to ensure all mocked calls were made.
