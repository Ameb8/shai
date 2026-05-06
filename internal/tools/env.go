package tools

import (
	"os"
	"strings"
)

// ScrubbedEnv returns the current process environment variables with sensitive
// information like API keys, secrets, and tokens filtered out.
func ScrubbedEnv() []string {
	raw := os.Environ()
	filtered := make([]string, 0, len(raw))

	for _, item := range raw {
		key, _, ok := strings.Cut(item, "=")
		if !ok || isSensitiveEnvKey(key) {
			continue
		}
		filtered = append(filtered, item)
	}

	return filtered
}

// isSensitiveEnvKey returns true if the environment variable key is likely
// to contain secrets or credentials based on common naming patterns.
func isSensitiveEnvKey(key string) bool {
	upper := strings.ToUpper(key)
	sensitiveContains := []string{
		"KEY",
		"SECRET",
		"TOKEN",
		"PASSWORD",
		"CREDENTIAL",
		"AUTH",
		"CERT",
	}
	for _, marker := range sensitiveContains {
		if strings.Contains(upper, marker) {
			return true
		}
	}

	sensitivePrefixes := []string{
		"AWS_",
		"GITHUB_",
		"OPENAI_",
		"ANTHROPIC_",
		"GEMINI_",
	}
	for _, prefix := range sensitivePrefixes {
		if strings.HasPrefix(upper, prefix) {
			return true
		}
	}

	return false
}
