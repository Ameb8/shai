package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const (
	defaultRunTimeout = 8 * time.Second
	maxOutputBytes    = 4000
)

// RunQueryArgs contains the executable and arguments for a command to be executed.
type RunQueryArgs struct {
	Executable string   `json:"executable"`
	Args       []string `json:"args"`
}

// RunQueryResult contains the output and execution status of a command.
type RunQueryResult struct {
	Stdout    string `json:"stdout"`
	Stderr    string `json:"stderr"`
	ExitCode  int    `json:"exit_code"`
	Truncated bool   `json:"truncated"`
}

// RunQuery executes a whitelisted system command with a timeout, scrubbed
// environment, and output size limits.
func RunQuery(ctx context.Context, args RunQueryArgs) (RunQueryResult, error) {
	if err := validateRunQueryArgs(args); err != nil {
		return RunQueryResult{}, err
	}

	ctx, cancel := context.WithTimeout(ctx, defaultRunTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, args.Executable, args.Args...)
	cmd.Env = ScrubbedEnv()

	remaining := maxOutputBytes
	truncated := false
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &limitedWriter{buf: &stdout, remaining: &remaining, truncated: &truncated}
	cmd.Stderr = &limitedWriter{buf: &stderr, remaining: &remaining, truncated: &truncated}

	err := cmd.Run()
	result := RunQueryResult{
		Stdout:    stdout.String(),
		Stderr:    stderr.String(),
		ExitCode:  0,
		Truncated: truncated,
	}

	if ctx.Err() == context.DeadlineExceeded {
		result.ExitCode = -1
		return result, fmt.Errorf("command timed out after %s", defaultRunTimeout)
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			return result, nil
		}
		return result, err
	}

	return result, nil
}

// validateRunQueryArgs ensures the command is a bare executable name and
// complies with the global safety policy registry.
func validateRunQueryArgs(args RunQueryArgs) error {
	executable := strings.TrimSpace(args.Executable)
	if executable == "" {
		return fmt.Errorf("executable is required")
	}
	if strings.ContainsAny(executable, `/\`) {
		return fmt.Errorf("executable must be a bare command name")
	}

	policy, ok := AllowedCommands()[executable]
	if !ok {
		return fmt.Errorf("executable %q is not allowed", executable)
	}
	if err := policy.Allow(args.Args); err != nil {
		return fmt.Errorf("%s denied: %w", executable, err)
	}

	return nil
}

// limitedWriter captures output up to a maximum byte limit and tracks truncation.
type limitedWriter struct {
	buf       *bytes.Buffer
	remaining *int
	truncated *bool
}

// Write implements io.Writer for limitedWriter, enforcing the output budget.
func (w *limitedWriter) Write(p []byte) (int, error) {
	if *w.remaining <= 0 {
		*w.truncated = true
		return len(p), nil
	}

	toWrite := len(p)
	if toWrite > *w.remaining {
		toWrite = *w.remaining
		*w.truncated = true
	}

	if toWrite > 0 {
		if _, err := w.buf.Write(p[:toWrite]); err != nil {
			return 0, err
		}
		*w.remaining -= toWrite
	}

	return len(p), nil
}
