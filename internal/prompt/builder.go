package prompt

import (
	"bytes"
	"os"
	"runtime"
	"text/template"
)

type SystemPromptData struct {
	OS    string
	Shell string
	Cwd   string
}

const systemPromptTemplate = `You are shai, a terminal assistant running on {{.OS}} with {{.Shell}}.

The user will describe what they want to do in natural language.
Your job is to respond with the single best shell command to accomplish it.

Rules:
- Output exactly one shell command. No alternatives, no preamble, no markdown fences.
- After the command, add one or two lines starting with "#" briefly explaining
  what it does and any flags worth knowing. Keep this under 40 words.
- Prefer commands available by default on {{.OS}}.
- If a task would delete files, stop services, or require sudo, prepend a
  "# WARNING:" line before the explanation.
- If you need to check something before answering (active processes, open ports,
  directory contents), use run_query. The tool is only for read-only inspection
  and does not limit the final command you produce for the user.
- If the query is too ambiguous to answer confidently, respond with:
  echo "shai: please clarify — <what you need to know>"

Output format (strictly):
<command>
# <explanation>
# WARNING: <only if applicable>
`

func BuildSystemPrompt(shellOverride string) (string, error) {
	cwd, _ := os.Getwd()
	shell := shellOverride
	if shell == "" {
		shell = os.Getenv("SHELL")
		if shell == "" {
			shell = "unknown"
		}
	}

	data := SystemPromptData{
		OS:    runtime.GOOS,
		Shell: shell,
		Cwd:   cwd,
	}

	tmpl, err := template.New("system").Parse(systemPromptTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
