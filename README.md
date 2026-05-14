# shai

`shai` is a terminal-native AI agent that turns a natural-language request into a single shell command.

It prints:
- Explanation lines prefixed with `#`
- The generated command prefixed with `>`

`shai` does not execute commands directly. Instead, it stages them for your review. In *Zsh* shells, they are injected into the shell's input buffer, allowing for execution by simply pressing the enter key. Bash shells print out the command normally and copies it to the keyboard.

`shai` is able to query system and environment data in order to construct correct commands. The final output command returned by `shai` is not restricted to read-only commands. `shai` will output a warning if LLM deems command to be risky or dangerous.

---

## Installation

---

### HomeBrew (Recommended for MacOS)

#### Install `shai` with the Official Tap:

```bash
brew install ameb8/tools/shai
```

or

```bash
brew tap ameb8/tools
brew install shai
```

#### Source `shai` Automatically

##### Add to `~/.zshrc`

```bash
source $(brew --prefix)/etc/profile.d/shai.zsh
```

##### Add to `~/.bashrc`

```bash
source $(brew --prefix)/etc/profile.d/shai.sh
```

This installs the `_shai_bin` binary and configures shell wrappers that provide the best experience for your shell.

---

### Install Script (Recommended for Linux)

The easiest way to install `shai` on Linux. Automatically detects your architecture and shell.

```bash
curl -fsSL https://raw.githubusercontent.com/ameb8/shai/master/install.sh | bash
```

This installs the `_shai_bin` binary and configures shell wrappers that provide the best experience for your shell.

#### Adding `shai` to Another Shell

If you switch shells later (e.g. from bash to zsh), run the included registration script:

```bash
bash ~/.config/shai/install-shell.sh
```

Then reload your shell config:

```bash
# zsh
source ~/.zshrc

# bash
source ~/.bashrc
```

---

### Other Installation Methods

<details>
<summary><b>Manual installation from GitHub Releases</b></summary>

You can also download the pre-compiled binaries and shell wrapper scripts directly from the *GitHub Releases* page.

#### Download the archive

##### AMD64

```bash
curl -LO https://github.com/ameb8/shai/releases/download/v0.1.0/shai_linux_amd64.tar.gz
```

##### ARM64

```bash
curl -LO https://github.com/ameb8/shai/releases/download/v0.1.0/shai_linux_arm64.tar.gz
```

#### Extract the binary

##### AMD64

```bash
tar -xzf shai_linux_amd64.tar.gz
```

##### ARM64

```bash
tar -xzf shai_linux_arm64.tar.gz
```

#### Create Directories for the Binary and the Scripts

```bash
mkdir -p ~/.local/bin ~/.config/shai
```

#### Move the Binary into PATH

```bash
mv _shai_bin ~/.local/bin/
```

#### Move the Wrapper Scripts to a Config Folder

```bash
mv shai.sh shai.zsh ~/.config/shai/
```

#### Source `shai` Automatically

##### Add to `~/.zshrc`

```bash
source ~/.config/shai/shai.zsh
```

##### Add to `~/.bashrc`

```bash
source ~/.config/shai/shai.sh
```
</details>

---

## Basic Usage

Run with a natural language query:

```bash
> shai "find all logs in /var/log modified in the last 24 hours"
# Lists all `.log` files in `/var/log` modified in the last 24 hours. 
# Use `-type f` to exclude directories
> find /var/log -type f -mtime -1 -name '*.log'
```

---

### Shell Integration

`shai` behaves differently depending on your shell to provide the most ergonomic experience:

- **Zsh:** The generated command is injected directly into your input buffer. You can edit it or just press `Enter` to execute.
- **Bash:** Since Bash doesn't support buffer injection, `shai` automatically copies the generated command to your clipboard and displays it.

---

## Currently Supported Providers

- `gemini` (Google)
- `mistral` (Mistral AI)
- `grok` (xAI)

---

## Configuration

Config file location: `~/.config/shai/config.toml`

Config file can be modified manually or through the *CLI*

### Setup via CLI

1. **Set your API key:**
   ```bash
   shai config set-key --provider gemini
   ```

2. **Set the active provider:**
   ```bash
   shai config set-provider gemini
   ```

3. **(Optional) Set a default model:**
   ```bash
   shai config set-model --provider gemini gemini-1.5-flash
   ```

4. **View current configuration:**
   ```bash
   shai config get
   ```

### Minimal `config.toml` Example

```toml
[active]
provider = "gemini"

[providers.gemini]
api_key = "YOUR_API_KEY"
default_model = "gemini-1.5-flash"
```

---

## CLI Flags

- `--copy, -c` Force copy generated command to clipboard
- `--dry-run` print command output mode
- `--model, -m` model override for this run
- `--no-explain` hide explanation lines
- `--provider, -p` provider override for this run
- `--shell` override detected shell (`bash|zsh`)
- `--think` enabled flag
- `--verbose, -v` enabled flag
- `--version` version for shai

---

## License

MIT

---