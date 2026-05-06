package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ameb8/shai/internal/provider"
)

// RunQueryToolName is the unique identifier for the system inspection tool.
const RunQueryToolName = "run_query"

// Definitions returns the list of provider-agnostic tool definitions available to the agent.
func Definitions() []provider.ToolDefinition {
	return []provider.ToolDefinition{
		{
			Name:        RunQueryToolName,
			Description: "Run one safe, read-only whitelisted terminal command to inspect system, filesystem, process, network, git, package, or container state before producing the final shell command. This tool does not restrict the final command you output to the user.",
			Parameters: map[string]provider.ParameterDef{
				"executable": {
					Type:        "string",
					Description: "Bare executable name, for example ls, ps, git, docker, docker-compose, curl, npm, or pip. Paths and shell invocations are rejected.",
				},
				"args": {
					Type:        "array",
					Description: "Command arguments as separate strings. Do not include shell syntax such as pipes, redirection, semicolons, &&, or command substitution.",
				},
			},
			Required: []string{"executable", "args"},
		},
	}
}

// Dispatch routes a tool call to the appropriate internal tool implementation
// and handles argument parsing and result encoding.
func Dispatch(ctx context.Context, call provider.ToolCall) (string, error) {
	switch call.Name {
	case RunQueryToolName:
		args, err := parseRunQueryArgs(call.Args)
		if err != nil {
			return "", err
		}
		result, err := RunQuery(ctx, args)
		if err != nil {
			resultJSON, marshalErr := json.Marshal(map[string]any{
				"error":  err.Error(),
				"result": result,
			})
			if marshalErr != nil {
				return "", err
			}
			return string(resultJSON), err
		}
		return marshalJSON(result)
	default:
		return "", fmt.Errorf("unknown tool %q", call.Name)
	}
}

// parseRunQueryArgs unmarshals raw tool arguments into a RunQueryArgs struct.
func parseRunQueryArgs(raw map[string]any) (RunQueryArgs, error) {
	executable, ok := raw["executable"].(string)
	if !ok {
		return RunQueryArgs{}, fmt.Errorf("run_query executable must be a string")
	}

	args, err := stringSlice(raw["args"])
	if err != nil {
		return RunQueryArgs{}, fmt.Errorf("run_query args: %w", err)
	}

	return RunQueryArgs{
		Executable: executable,
		Args:       args,
	}, nil
}

// stringSlice safely converts an interface value into a slice of strings.
func stringSlice(value any) ([]string, error) {
	if value == nil {
		return nil, nil
	}

	switch typed := value.(type) {
	case []string:
		return typed, nil
	case []any:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			str, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("all values must be strings")
			}
			result = append(result, str)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("must be an array of strings")
	}
}

// marshalJSON encodes a value as a JSON string.
func marshalJSON(value any) (string, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}
