# Tool Implementation Patterns

## Registry Registration (internal/tools/registry.go)

### Dispatch Switch Case
```go
case MyToolName:
    args, err := parseMyToolArgs(call.Args)
    if err != nil {
        return "", err
    }
    result, err := RunMyTool(ctx, args)
    if err != nil {
        // Return partial results if possible
        return marshalJSON(map[string]any{
            "error": err.Error(),
            "result": result,
        })
    }
    return marshalJSON(result)
```

### Argument Parser
```go
func parseMyToolArgs(raw map[string]any) (MyToolArgs, error) {
    param1, ok := raw["param1"].(string)
    if !ok {
        return MyToolArgs{}, fmt.Errorf("my_tool param1 must be a string")
    }
    return MyToolArgs{Param1: param1}, nil
}
```

## Tool Logic (internal/tools/my_tool.go)

### Struct Definitions
```go
type MyToolArgs struct {
    Param1 string `json:"param1"`
}

type MyToolResult struct {
    Data string `json:"data"`
}
```

### Core Execution logic
```go
func RunMyTool(ctx context.Context, args MyToolArgs) (MyToolResult, error) {
    // Implementation here...
    return MyToolResult{Data: "result"}, nil
}
```
