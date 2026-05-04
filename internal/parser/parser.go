package parser

import (
	"strings"
)

type ParsedResponse struct {
	Command     string
	Explanation []string // lines starting with "#", stripped of leading "# "
	HasWarning  bool
}

func Parse(raw string) (ParsedResponse, error) {
	raw = strings.TrimSpace(raw)
	// Strip markdown code fences if present
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
			// First non-comment line is the command
			// Subsequent ones (if any) are joined to handle multi-line commands
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
