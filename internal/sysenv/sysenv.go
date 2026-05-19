package sysenv

import (
	"os"
	"runtime"
	"strings"
	"sync"
)

// Runtime contains the environmental context of the current execution.
type Runtime struct {
	OS    string
	Shell string
	Cwd   string
}

var (
	instance *Runtime
	once     sync.Once
)

// GetRuntime returns the cached runtime environment. If it hasn't been
// initialized yet, it detects the environment using the provided shell override.
func GetRuntime(shellOverride string) *Runtime {
	once.Do(func() {
		cwd, _ := os.Getwd()
		shell := shellOverride
		if shell == "" {
			shell = os.Getenv("SHELL")
			if shell == "" {
				shell = "unknown"
			}
		}

		instance = &Runtime{
			OS:    runtime.GOOS,
			Shell: shell,
			Cwd:   cwd,
		}
	})
	return instance
}

// ResetRuntimeForTest clears the cached runtime environment, allowing it to be
// re-detected in subsequent calls to GetRuntime. This is intended for use in
// unit tests only.
func ResetRuntimeForTest() {
	instance = nil
	once = sync.Once{}
}

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
