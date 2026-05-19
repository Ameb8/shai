---
name: add-tool
description: Add a new tool to the shai agent's reasoning loop. Use this skill when you need to extend the agent's capabilities with a new provider-agnostic tool, such as specialized file parsers, network utilities, or system monitors.
---

# Adding a New Tool

This guide walks through the process of adding a new tool to the `shai` agent.

## Workflow Overview

1.  **Define**: Add the tool definition to `internal/tools/registry.go`.
2.  **Implement**: Create the tool's core logic in a new file in `internal/tools/`.
3.  **Register**: Update the `Dispatch` function and add an argument parser in `internal/tools/registry.go`.
4.  **Test**: Verify the tool with unit tests.

## 1. Define the Tool

Open `internal/tools/registry.go` and add a new `provider.ToolDefinition` to the `Definitions()` function.

```go
{
    Name:        "my_tool_name",
    Description: "Clear description of what the tool does.",
    Parameters: map[string]provider.ParameterDef{
        "param1": {
            Type:        "string",
            Description: "Description of param1",
        },
    },
    Required: []string{"param1"},
}
```

## 2. Implement the Tool Logic

Create a new file `internal/tools/my_tool.go`. Use the template provided in [assets/tool_template.go](assets/tool_template.go).

- Define `MyToolArgs` and `MyToolResult` structs.
- Implement the core function (e.g., `RunMyTool(ctx context.Context, args MyToolArgs)`).
- **Handling System Context**: If your tool needs information about the OS, shell, or current working directory, use the `internal/sysenv` package.
  ```go
  import "github.com/ameb8/shai/internal/sysenv"

  func RunMyTool(ctx context.Context, args MyToolArgs) {
      runtime := sysenv.GetRuntime("") // Pass override if available, otherwise empty string
      // Use runtime.OS, runtime.Shell, runtime.Cwd
  }
  ```
- **Environment Scrubbing**: Use `sysenv.ScrubbedEnv()` if you are executing subprocesses to ensure secrets are not leaked.
- Ensure error handling is robust and returns JSON-serializable results.

## 3. Register and Parse Arguments

In `internal/tools/registry.go`:

1.  Add a case to the `Dispatch` function's switch statement.
2.  Implement a `parseMyToolArgs` function to validate and convert raw `map[string]any` arguments into your typed struct.

See [references/patterns.md](references/patterns.md) for implementation snippets.

## 4. Verify with Tests

Add unit tests in `internal/tools/my_tool_test.go` or update existing tests.
- Test successful execution.
- Test argument validation errors.
- Test tool-specific error conditions.
