package tools

import (
	"fmt"
	"net/url"
	"strings"
)

// CommandPolicy defines the interface for validating whether a set of command
// arguments is allowed according to safety rules.
type CommandPolicy interface {
	Allow(args []string) error
}

// anyArgsPolicy is a policy that allows all arguments for a command.
type anyArgsPolicy struct{}

func (anyArgsPolicy) Allow([]string) error {
	return nil
}

// AnyArgs is a global policy instance that permits any arguments.
var AnyArgs CommandPolicy = anyArgsPolicy{}

// SubcommandList is a policy that restricts execution to a specific set
// of allowed subcommands.
type SubcommandList map[string]struct{}

// NewSubcommandList creates a new SubcommandList from a list of allowed subcommand names.
func NewSubcommandList(names ...string) SubcommandList {
	allowed := make(SubcommandList, len(names))
	for _, name := range names {
		allowed[name] = struct{}{}
	}
	return allowed
}

// Allow validates that the first argument is an allowed subcommand.
func (p SubcommandList) Allow(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing subcommand")
	}

	subcommand := args[0]
	if strings.HasPrefix(subcommand, "-") {
		return fmt.Errorf("subcommand must be the first argument")
	}
	if _, ok := p[subcommand]; !ok {
		return fmt.Errorf("subcommand %q is not allowed", subcommand)
	}

	return nil
}

// CommandTree routes argument validation to nested policies based on the
// first argument (the subcommand).
type CommandTree map[string]CommandPolicy

// Allow routes the validation request to the policy matching the first argument.
func (p CommandTree) Allow(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing subcommand")
	}

	subcommand := args[0]
	if strings.HasPrefix(subcommand, "-") {
		return fmt.Errorf("subcommand must be the first argument")
	}

	child, ok := p[subcommand]
	if !ok {
		return fmt.Errorf("subcommand %q is not allowed", subcommand)
	}

	return child.Allow(args[1:])
}

// denyArgsPolicy blocks specific exact arguments or argument prefixes.
type denyArgsPolicy struct {
	deniedExact    map[string]struct{}
	deniedPrefixes []string
}

// newDenyArgsPolicy creates a policy that rejects the specified exact arguments
// or any argument starting with the given prefixes.
func newDenyArgsPolicy(exact []string, prefixes []string) denyArgsPolicy {
	denied := make(map[string]struct{}, len(exact))
	for _, arg := range exact {
		denied[arg] = struct{}{}
	}
	return denyArgsPolicy{
		deniedExact:    denied,
		deniedPrefixes: prefixes,
	}
}

// Allow checks each argument against the set of denied strings and prefixes.
func (p denyArgsPolicy) Allow(args []string) error {
	for _, arg := range args {
		if _, ok := p.deniedExact[arg]; ok {
			return fmt.Errorf("argument %q is not allowed", arg)
		}
		for _, prefix := range p.deniedPrefixes {
			if strings.HasPrefix(arg, prefix) {
				return fmt.Errorf("argument %q is not allowed", arg)
			}
		}
	}
	return nil
}

// gitBranchPolicy restricts git branch commands to safe, read-only inspection flags.
type gitBranchPolicy struct{}

func (gitBranchPolicy) Allow(args []string) error {
	allowedFlags := map[string]struct{}{
		"-a":             {},
		"-r":             {},
		"-v":             {},
		"-vv":            {},
		"--all":          {},
		"--list":         {},
		"--remotes":      {},
		"--show-current": {},
		"--verbose":      {},
	}

	allowPatterns := false
	for _, arg := range args {
		if _, ok := allowedFlags[arg]; ok {
			if arg == "--list" {
				allowPatterns = true
			}
			continue
		}
		if strings.HasPrefix(arg, "-") {
			return fmt.Errorf("git branch flag %q is not allowed", arg)
		}
		if !allowPatterns {
			return fmt.Errorf("git branch positional args are not allowed")
		}
	}

	return nil
}

// gitRemotePolicy restricts git remote commands to read-only operations.
type gitRemotePolicy struct{}

func (gitRemotePolicy) Allow(args []string) error {
	if len(args) == 0 {
		return nil
	}
	if len(args) == 1 && args[0] == "-v" {
		return nil
	}
	if args[0] == "show" || args[0] == "get-url" {
		return nil
	}
	return fmt.Errorf("git remote command %q is not allowed", args[0])
}

// curlGetOnlyPolicy ensures curl is only used for GET requests and doesn't
// write to files or upload data.
type curlGetOnlyPolicy struct{}

func (curlGetOnlyPolicy) Allow(args []string) error {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		lower := strings.ToLower(arg)

		switch {
		case arg == "-X" || lower == "--request":
			if i+1 >= len(args) {
				return fmt.Errorf("%s requires a method", arg)
			}
			i++
			if !strings.EqualFold(args[i], "GET") {
				return fmt.Errorf("curl may only use GET requests")
			}
		case strings.HasPrefix(arg, "-X") && len(arg) > 2:
			if !strings.EqualFold(arg[2:], "GET") {
				return fmt.Errorf("curl may only use GET requests")
			}
		case strings.HasPrefix(lower, "--request="):
			method := strings.TrimPrefix(arg, "--request=")
			if !strings.EqualFold(method, "GET") {
				return fmt.Errorf("curl may only use GET requests")
			}
		case arg == "-d" || lower == "--data" || lower == "--data-raw" ||
			lower == "--data-binary" || lower == "--data-urlencode" ||
			arg == "-F" || lower == "--form" ||
			arg == "-T" || lower == "--upload-file":
			return fmt.Errorf("curl request body and upload flags are not allowed")
		case arg == "-o" || arg == "-O" || lower == "--output" || lower == "--remote-name":
			return fmt.Errorf("curl output file flags are not allowed")
		case strings.HasPrefix(arg, "-d") && len(arg) > 2:
			return fmt.Errorf("curl request body flags are not allowed")
		case strings.HasPrefix(lower, "--data=") ||
			strings.HasPrefix(lower, "--data-raw=") ||
			strings.HasPrefix(lower, "--data-binary=") ||
			strings.HasPrefix(lower, "--data-urlencode=") ||
			strings.HasPrefix(lower, "--form=") ||
			strings.HasPrefix(lower, "--upload-file="):
			return fmt.Errorf("curl request body and upload flags are not allowed")
		case strings.HasPrefix(lower, "--output="):
			return fmt.Errorf("curl output file flags are not allowed")
		case arg == "-I" || lower == "--head":
			return fmt.Errorf("curl HEAD requests are not allowed")
		}

		if looksLikeURL(arg) {
			if err := allowHTTPURL(arg); err != nil {
				return err
			}
		}
	}

	return nil
}

// looksLikeURL returns true if the string contains a scheme separator.
func looksLikeURL(value string) bool {
	return strings.Contains(value, "://")
}

// allowHTTPURL validates that a URL uses the http or https scheme.
func allowHTTPURL(value string) error {
	parsed, err := url.Parse(value)
	if err != nil {
		return fmt.Errorf("invalid URL %q", value)
	}
	switch parsed.Scheme {
	case "http", "https":
		return nil
	default:
		return fmt.Errorf("curl URL scheme %q is not allowed", parsed.Scheme)
	}
}

// AllowedCommands returns the global registry of whitelisted commands and
// their associated safety policies.
func AllowedCommands() map[string]CommandPolicy {
	findPolicy := newDenyArgsPolicy(
		[]string{"-delete", "-exec", "-execdir", "-ok", "-okdir"},
		[]string{"-fprint"},
	)
	gitReadOnlyOutputPolicy := newDenyArgsPolicy(
		[]string{"--output"},
		[]string{"--output="},
	)

	return map[string]CommandPolicy{
		"cat":      AnyArgs,
		"df":       AnyArgs,
		"dig":      AnyArgs,
		"du":       AnyArgs,
		"file":     AnyArgs,
		"find":     findPolicy,
		"free":     AnyArgs,
		"head":     AnyArgs,
		"id":       AnyArgs,
		"lsof":     AnyArgs,
		"ls":       AnyArgs,
		"netstat":  AnyArgs,
		"nslookup": AnyArgs,
		"ping":     AnyArgs,
		"ps":       AnyArgs,
		"ss":       AnyArgs,
		"stat":     AnyArgs,
		"tail":     AnyArgs,
		"uname":    AnyArgs,
		"uptime":   AnyArgs,
		"wc":       AnyArgs,
		"whereis":  AnyArgs,
		"which":    AnyArgs,
		"whoami":   AnyArgs,

		"cargo": NewSubcommandList("tree", "metadata"),
		"curl":  curlGetOnlyPolicy{},
		"git": CommandTree{
			"branch": gitBranchPolicy{},
			"diff":   gitReadOnlyOutputPolicy,
			"log":    AnyArgs,
			"remote": gitRemotePolicy{},
			"show":   gitReadOnlyOutputPolicy,
			"stash":  NewSubcommandList("list"),
			"status": AnyArgs,
		},
		"npm": NewSubcommandList("list", "ls", "info", "show", "outdated"),
		"pip": NewSubcommandList("list", "show", "freeze"),

		"docker": CommandTree{
			"compose": CommandTree{
				"ps": AnyArgs,
			},
			"images":  AnyArgs,
			"inspect": AnyArgs,
			"logs":    AnyArgs,
			"network": NewSubcommandList("ls"),
			"ps":      AnyArgs,
			"stats":   AnyArgs,
			"volume":  NewSubcommandList("ls"),
		},
		"docker-compose": NewSubcommandList("ps"),
	}
}
