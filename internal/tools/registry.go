package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ameb8/shai/internal/provider"
)

// RunQueryToolName is the unique identifier for the system inspection tool.
const (
	RunQueryToolName = "run_query"
	ProcInfoToolName = "proc_info"
)

// Definitions returns the list of provider-agnostic tool definitions available to the agent.
// These definitions inform the LLM about the tool's purpose and expected parameters.
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
		{
			Name:        ProcInfoToolName,
			Description: "Get structured information about running processes, optionally filtered by port, name, or user. Replaces ps, lsof, and ss for process/port inspection.",
			Parameters: map[string]provider.ParameterDef{
				"port": {
					Type:        "integer",
					Description: "Filter processes by the port they are listening on.",
				},
				"name": {
					Type:        "string",
					Description: "Filter processes by name (case-insensitive substring match).",
				},
				"user": {
					Type:        "string",
					Description: "Filter processes by user name.",
				},
			},
		},
	}
}

// Dispatch executes the requested tool call by routing it to its internal implementation.
// It handles argument parsing, execution, and result encoding for the agent's tool loop.
func Dispatch(ctx context.Context, call provider.ToolCall) (string, error) {
	switch call.Name {
	case RunQueryToolName:
		// Parse and validate arguments before executing the shell query.
		args, err := parseRunQueryArgs(call.Args)
		if err != nil {
			return "", err
		}

		// Execute the command and capture its output.
		result, err := RunQuery(ctx, args)
		if err != nil {
			// Encode the error and any partial result for the LLM to interpret.
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
	case ProcInfoToolName:
		args, err := parseProcInfoArgs(call.Args)
		if err != nil {
			return "", err
		}
		result, err := RunProcInfo(ctx, args)
		if err != nil {
			return marshalJSON(map[string]any{
				"error":  err.Error(),
				"result": result,
			})
		}
		return marshalJSON(result)
	default:
		return "", fmt.Errorf("unknown tool %q", call.Name)
	}
}

// parseRunQueryArgs converts raw map arguments into a typed RunQueryArgs structure.
// It ensures that all required parameters are present and correctly typed.
func parseRunQueryArgs(raw map[string]any) (RunQueryArgs, error) {
	// Validate that the executable is a plain string.
	executable, ok := raw["executable"].(string)
	if !ok {
		return RunQueryArgs{}, fmt.Errorf("run_query executable must be a string")
	}

	// Ensure arguments are provided as a string slice.
	args, err := stringSlice(raw["args"])
	if err != nil {
		return RunQueryArgs{}, fmt.Errorf("run_query args: %w", err)
	}

	return RunQueryArgs{
		Executable: executable,
		Args:       args,
	}, nil
}

// parseProcInfoArgs converts raw map arguments into a typed ProcInfoArgs structure.
func parseProcInfoArgs(raw map[string]any) (ProcInfoArgs, error) {
	args := ProcInfoArgs{}

	if port, ok := raw["port"]; ok {
		switch p := port.(type) {
		case float64:
			args.Port = int(p)
		case int:
			args.Port = p
		case string:
			if val, err := strconv.Atoi(p); err == nil {
				args.Port = val
			} else {
				return ProcInfoArgs{}, fmt.Errorf("proc_info port must be an integer")
			}
		default:
			return ProcInfoArgs{}, fmt.Errorf("proc_info port must be an integer")
		}
	}

	if name, ok := raw["name"].(string); ok {
		args.Name = name
	}

	if user, ok := raw["user"].(string); ok {
		args.User = user
	}

	return args, nil
}

// stringSlice converts an interface value into a string slice to normalize LLM outputs.
// It returns an error if the input contains non-string elements.
func stringSlice(value any) ([]string, error) {
	if value == nil {
		return nil, nil
	}

	switch typed := value.(type) {
	case []string:
		return typed, nil
	case []any:
		// Validate each element to ensure type safety.
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

// marshalJSON serializes any value into its JSON string representation for LLM consumption.
func marshalJSON(value any) (string, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}
