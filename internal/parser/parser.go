package parser

import (
	"strings"
)

// ParsedResponse represents the structured output of an LLM completion.
// It separates the executable command from its human-readable explanation and metadata.
type ParsedResponse struct {
	Command     string
	Explanation []string // lines starting with "#", stripped of leading "# "
	HasWarning  bool
}

// Parse extracts the command and explanation from a raw LLM response string.
// It handles markdown code fences and strips shell prompts from the command lines.
func Parse(raw string) (ParsedResponse, error) {
	raw = strings.TrimSpace(raw)
	// Remove markdown code blocks to isolate the raw command and comments.
	if strings.HasPrefix(raw, "```") {
		lines := strings.Split(raw, "\n")
		if len(lines) > 2 {
			raw = strings.Join(lines[1:len(lines)-1], "\n")
		}
	}

	lines := strings.Split(raw, "\n")
	var response ParsedResponse
	var commandLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if strings.HasPrefix(trimmed, "#") {
			explanation := strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
			if strings.HasPrefix(explanation, "WARNING:") {
				response.HasWarning = true
			}
			response.Explanation = append(response.Explanation, explanation)
		} else {
			// Extract the command line, removing common shell prompt prefixes.
			// Subsequent lines are joined to handle multi-line commands.
			cmdLine := strings.TrimPrefix(trimmed, "$ ")
			cmdLine = strings.TrimPrefix(cmdLine, "> ")
			commandLines = append(commandLines, cmdLine)
		}
	}

	if len(commandLines) > 0 {
		response.Command = strings.Join(commandLines, "\n")
	}

	return response, nil
}
