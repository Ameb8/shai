package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ameb8/shai/internal/provider"
	"github.com/ameb8/shai/internal/tools"
)

// MaxToolIterations is the maximum number of times the agent can call tools
// in a single completion request to prevent infinite loops.
const MaxToolIterations = 10

// Complete handles the LLM completion loop, including automated tool use.
// It continues to call the provider as long as tool calls are requested,
// up to MaxToolIterations.
func Complete(ctx context.Context, p provider.Provider, messages []provider.Message) (provider.CompletionResponse, error) {
	for i := 0; i < MaxToolIterations; i++ {
		resp, err := p.Complete(ctx, provider.CompletionRequest{
			Messages: messages,
			Tools:    tools.Definitions(),
		})
		if err != nil {
			return provider.CompletionResponse{}, err
		}

		if resp.StopReason != "tool_use" || len(resp.ToolCalls) == 0 {
			return resp, nil
		}

		messages = append(messages, provider.Message{
			Role:      "assistant",
			Content:   resp.Content,
			ToolCalls: resp.ToolCalls,
		})

		for _, call := range resp.ToolCalls {
			printToolInvocation(call)
			output, dispatchErr := tools.Dispatch(ctx, call)
			toolErr := false
			if dispatchErr != nil {
				toolErr = true
				if output == "" {
					output = toolErrorJSON(dispatchErr)
				}
			}

			messages = append(messages, provider.Message{
				Role:       "tool",
				Content:    output,
				ToolCallID: call.ID,
				ToolName:   call.Name,
				ToolError:  toolErr,
			})
		}
	}

	return provider.CompletionResponse{}, fmt.Errorf("tool loop exceeded %d iterations", MaxToolIterations)
}

// toolErrorJSON formats a tool execution error into a JSON string.
func toolErrorJSON(err error) string {
	encoded, marshalErr := json.Marshal(map[string]any{
		"error": err.Error(),
	})
	if marshalErr != nil {
		return `{"error":"tool dispatch failed"}`
	}
	return string(encoded)
}

// printToolInvocation logs a user-friendly description of the tool being
// called to stderr.
func printToolInvocation(call provider.ToolCall) {
	if call.Name != tools.RunQueryToolName {
		fmt.Fprintf(os.Stderr, "# tool: invoking %s\n", call.Name)
		return
	}

	executable, _ := call.Args["executable"].(string)
	args, _ := stringSlice(call.Args["args"])
	command := strings.TrimSpace(strings.Join(append([]string{executable}, args...), " "))
	if command == "" {
		command = tools.RunQueryToolName
	}
	fmt.Fprintf(os.Stderr, "# tool: invoking %s\n", command)
}

// stringSlice safely extracts a string slice from an interface value.
func stringSlice(value any) ([]string, bool) {
	switch typed := value.(type) {
	case []string:
		return typed, true
	case []any:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			str, ok := item.(string)
			if !ok {
				return nil, false
			}
			result = append(result, str)
		}
		return result, true
	default:
		return nil, false
	}
}
