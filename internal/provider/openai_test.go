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

// TestNewOpenAIProvider_BaseURL verifies that the OpenAI provider correctly handles
// both default and custom base URLs during initialization.
func TestNewOpenAIProvider_BaseURL(t *testing.T) {
	ctx := context.Background()

	t.Run("default url", func(t *testing.T) {
		// Ensure the default OpenAI API URL is used when no custom URL is provided.
		p, err := NewOpenAIProvider(ctx, config.ProviderConfig{APIKey: "test-key"}, "")
		require.NoError(t, err)
		assert.Equal(t, openaiDefaultURL, p.url)
	})

	t.Run("custom url", func(t *testing.T) {
		// Ensure a custom base URL can be injected via configuration.
		customURL := "https://custom.openai.com/v1/chat/completions"
		p, err := NewOpenAIProvider(ctx, config.ProviderConfig{APIKey: "test-key", BaseURL: customURL}, "")
		require.NoError(t, err)
		assert.Equal(t, customURL, p.url)
	})
}

// TestOpenAIComplete validates the OpenAI provider's ability to parse various
// API responses, including text completions and tool calls.
func TestOpenAIComplete(t *testing.T) {
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
				"usage": {"prompt_tokens": 12, "completion_tokens": 3}
			}`,
			expectedStop:   "tool_use",
			expectedTools:  1,
			expectedInput:  12,
			expectedOutput: 3,
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
				"usage": {"prompt_tokens": 6, "completion_tokens": 2}
			}`,
			expectedStop:   "end_turn",
			expectedTools:  0,
			expectedInput:  6,
			expectedOutput: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotReq openAIChatRequest
			// Mock the HTTP transport to intercept and validate requests.
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
			p, err := NewOpenAIProvider(ctx, config.ProviderConfig{APIKey: "test-key"}, "gpt-test")
			require.NoError(t, err)
			p.client = &http.Client{Transport: transport}

			// Execute the completion request and verify the parsed response.
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

			assert.Equal(t, "gpt-test", gotReq.Model)
			assert.Equal(t, tt.expectedStop, resp.StopReason)
			assert.Len(t, resp.ToolCalls, tt.expectedTools)
			assert.Equal(t, tt.expectedInput, resp.InputTokens)
			assert.Equal(t, tt.expectedOutput, resp.OutputTokens)

			// Verify tool call details if expected.
			if tt.expectedTools > 0 {
				assert.Equal(t, "call_1", resp.ToolCalls[0].ID)
				assert.Equal(t, "run_query", resp.ToolCalls[0].Name)
				assert.Equal(t, "ls", resp.ToolCalls[0].Args["executable"])
			}
		})
	}
}
