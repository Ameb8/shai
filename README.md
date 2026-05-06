# shai

`shai` turns a natural-language request into one shell command.

It prints:
- explanation lines prefixed with `#`
- the generated command prefixed with `>`

It does not execute the command for you.

## Basic Usage

Run with a query:

```bash
shai "show disk usage by folder in this directory"
```

Current top-level behavior:
- `shai [query]` runs a query
- `shai config ...` manages config

## Currently Supported Providers

Right now, implemented providers are:
- `gemini`
- `mistral`

## Query Flags (Current)

- `--provider, -p` provider override for this run
- `--model, -m` model override for this run
- `--think` enabled flag (wired in CLI)
- `--dry-run` print command output mode
- `--verbose, -v` enabled flag (wired in CLI)
- `--no-explain` hide explanation lines
- `--shell` override detected shell (`bash|zsh|fish`)

## Configuration

Config file location:

```text
~/.config/shai/config.toml
```

Directory/file permissions used by `shai`:
- `~/.config/shai` -> `0700`
- `config.toml` -> `0600`

### Manage Config from CLI

Set API key (interactive prompt):

```bash
shai config set-key --provider gemini
shai config set-key --provider mistral
```

Set active provider:

```bash
shai config set-provider gemini
```

Set default model for a provider:

```bash
shai config set-model --provider gemini gemini-2.5-flash
```

Show current config:

```bash
shai config get
```

### Minimal `config.toml` Example

```toml
[active]
provider = "gemini"

[providers.gemini]
api_key = "YOUR_API_KEY"
default_model = "gemini-2.5-flash"

[providers.mistral]
api_key = "YOUR_API_KEY"
default_model = "mistral-small-latest"
```

If no active provider is configured, query mode fails and asks you to set one.

## Build

```bash
go build ./...
```
