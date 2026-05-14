package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/ameb8/shai/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewMistralProvider_BaseURL verifies that the Mistral provider correctly uses the BaseURL from configuration.
func TestNewMistralProvider_BaseURL(t *testing.T) {
	ctx := context.Background()

	t.Run("default url", func(t *testing.T) {
		p, err := NewMistralProvider(ctx, config.ProviderConfig{APIKey: "test-key"}, "")
		require.NoError(t, err)
		assert.Equal(t, mistralDefaultURL, p.url)
	})

	t.Run("custom url", func(t *testing.T) {
		customURL := "https://custom.mistral.ai/v1"
		p, err := NewMistralProvider(ctx, config.ProviderConfig{APIKey: "test-key", BaseURL: customURL}, "")
		require.NoError(t, err)
		assert.Equal(t, customURL, p.url)
	})
}
// TestMistralComplete verifies that the Mistral provider correctly handles completion requests and tool calls.
func TestMistralComplete(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   string
		expectedStop   string
		expectedTools  int
		expectedInput  int
		expectedOutput int
	}{
		{
			name: "tool calls",
			responseBody: `{
				"choices": [{
					"finish_reason": "tool_calls",
					"message": {
						"role": "assistant",
						"tool_calls": [{
							"id": "call_1",
							"type": "function",
							"function": {
								"name": "run_query",
								"arguments": "{\"executable\":\"ls\",\"args\":[\"-la\"]}"
							}
						}]
					}
				}],
				"usage": {"prompt_tokens": 11, "completion_tokens": 4}
			}`,
			expectedStop:   "tool_use",
			expectedTools:  1,
			expectedInput:  11,
			expectedOutput: 4,
		},
		{
			name: "text response",
			responseBody: `{
				"choices": [{
					"finish_reason": "stop",
					"message": {
						"role": "assistant",
						"content": "Hello!"
					}
				}],
				"usage": {"prompt_tokens": 5, "completion_tokens": 2}
			}`,
			expectedStop:   "end_turn",
			expectedTools:  0,
			expectedInput:  5,
			expectedOutput: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotReq mistralChatRequest
			// Define a mock transport to capture and respond to API requests.
			transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
				assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
				err := json.NewDecoder(r.Body).Decode(&gotReq)
				require.NoError(t, err)

				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Body:       io.NopCloser(bytes.NewBufferString(tt.responseBody)),
				}, nil
			})

			ctx := context.Background()
			p, err := NewMistralProvider(ctx, config.ProviderConfig{APIKey: "test-key"}, "mistral-test")
			require.NoError(t, err)
			p.client = &http.Client{Transport: transport}

			// Execute the completion request and validate the response structure.
			resp, err := p.Complete(ctx, CompletionRequest{
				Messages: []Message{{Role: "user", Content: "hello"}},
				Tools: []ToolDefinition{{
					Name:        "run_query",
					Description: "Run a command",
					Parameters: map[string]ParameterDef{
						"executable": {Type: "string"},
						"args":       {Type: "array"},
					},
					Required: []string{"executable", "args"},
				}},
			})
			require.NoError(t, err)

			assert.Equal(t, "mistral-test", gotReq.Model)
			assert.Equal(t, tt.expectedStop, resp.StopReason)
			assert.Len(t, resp.ToolCalls, tt.expectedTools)
			assert.Equal(t, tt.expectedInput, resp.InputTokens)
			assert.Equal(t, tt.expectedOutput, resp.OutputTokens)

			if tt.expectedTools > 0 {
				assert.Equal(t, "call_1", resp.ToolCalls[0].ID)
				assert.Equal(t, "run_query", resp.ToolCalls[0].Name)
				assert.Equal(t, "ls", resp.ToolCalls[0].Args["executable"])
			}
		})
	}
}
