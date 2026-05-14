package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ameb8/shai/internal/config"
)

// TestConfigSetURL verifies that the 'config set-url' command correctly updates a provider's base URL.
func TestConfigSetURL(t *testing.T) {
	// Isolate the test from the user's real configuration by using a temporary home directory.
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Setup a fresh configuration state for the test.
	testCfg := &config.Config{
		Providers: map[string]config.ProviderConfig{
			"test-provider": {
				DefaultModel: "test-model",
			},
		},
	}

	// Initialize the command hierarchy with the mock configuration.
	root := NewRootCmd(testCfg)

	// Execute the set-url command targeting the test provider.
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetArgs([]string{"config", "set-url", "http://test.url", "--provider", "test-provider"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	// Assert that the configuration object was updated with the new URL.
	gotURL := testCfg.Providers["test-provider"].BaseURL
	wantURL := "http://test.url"
	if gotURL != wantURL {
		t.Errorf("BaseURL = %q, want %q", gotURL, wantURL)
	}

	// Validate that the user receives a confirmation message.
	gotOut := out.String()
	wantOut := "Base URL for test-provider set to http://test.url."
	if !strings.Contains(gotOut, wantOut) {
		t.Errorf("output = %q, want it to contain %q", gotOut, wantOut)
	}
}

// TestConfigGetURL verifies that the 'config get' command correctly displays the configured base URL.
func TestConfigGetURL(t *testing.T) {
	// Prepare a configuration containing an existing base URL.
	testCfg := &config.Config{
		Providers: map[string]config.ProviderConfig{
			"test-provider": {
				DefaultModel: "test-model",
				BaseURL:      "http://test.url",
			},
		},
	}

	// Initialize the command hierarchy.
	root := NewRootCmd(testCfg)

	// Execute the config get command to retrieve settings.
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetArgs([]string{"config", "get"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	// Verify that the configured base URL is visible in the command output.
	gotOut := out.String()
	if !strings.Contains(gotOut, "Base URL: http://test.url") {
		t.Errorf("output does not contain expected Base URL, got: %q", gotOut)
	}
}
