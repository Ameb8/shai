# GoReleaser + Homebrew Distribution Setup

A step-by-step guide for releasing a Go CLI tool via GoReleaser with a Homebrew tap.

This guide covers the `shai` architecture specifically: a Go/Cobra binary (`_shai_bin`) that is invoked by shell wrapper functions (`shai` in `shai.sh` / `shai.zsh`). The wrapper functions must be auto-sourced on shell startup — no manual `source` required from the user. Homebrew handles this by installing the shell files into directories that are automatically loaded by bash and zsh.

---

## Prerequisites

- A Go CLI project in a GitHub repository
- GitHub CLI (`gh`) installed, or access to GitHub in the browser
- GoReleaser installed locally for testing (`brew install goreleaser` or see [goreleaser.com](https://goreleaser.com/install/))
- `shai.sh` and `shai.zsh` already written and present in your repository under a directory like `shell/`

---

## Step 1: Set up version injection in your Go code

Add version variables to your `main.go` that GoReleaser will populate at build time via ldflags.

```go
// main.go
package main

var (
    version = "dev"
    commit  = "none"
    date    = "unknown"
)

func main() {
    // optionally wire these into a --version flag
}
```

If you're using a library like [cobra](https://github.com/spf13/cobra), set these on your root command:

```go
rootCmd.Version = fmt.Sprintf("%s (%s, %s)", version, commit, date)
```

These variables are placeholders — GoReleaser overwrites them at build time using `-ldflags`. When running normally (e.g. `go run .`), they stay as `"dev"`, `"none"`, `"unknown"`.

---

## Step 1b: Name the Go binary `_shai_bin` and write the shell wrappers

The user-facing command is `shai`, which is a shell function — not the Go binary directly. The binary is named `_shai_bin` to make clear it is an internal implementation detail and to avoid colliding with the shell function name. Homebrew puts `_shai_bin` on `$PATH`, so the wrappers can call it by name with no hardcoded paths.

The two implementations differ because of what each shell can do:

**`shell/shai.sh` (bash)** — bash cannot stage a command in the user's line editor buffer, so it passes `--copy` to the binary, which copies the suggested command to the clipboard instead:

```bash
# shell/shai.sh
shai() {
    _shai_bin --copy "$@"
}
```

**`shell/shai.zsh` (zsh)** — zsh's `print -z` can push a command directly into the interactive shell's ZLE input buffer, letting the user review/edit it before pressing Enter. This requires running inside the user's shell process, which is why a function (not a script) is mandatory here:

```zsh
# shell/shai.zsh
_SHAI_CMD_FILE_PREFIX="shai_cmd"

shai() {
    local cmd_file
    mkdir -p "${TMPDIR:-/tmp}"
    cmd_file=$(mktemp "${TMPDIR:-/tmp}/${_SHAI_CMD_FILE_PREFIX}.XXXXXX") || {
        echo "shai: failed to create temp file" >&2
        return 1
    }

    _shai_bin --cmd-file="$cmd_file" "$@"
    local exit_code=$?

    if [[ $exit_code -ne 0 ]]; then
        rm -f "$cmd_file"
        return $exit_code
    fi

    local cmd
    cmd=$(< "$cmd_file")
    rm -f "$cmd_file"

    # Stage the command in the ZLE buffer for the user to review/edit/execute
    [[ -n "$cmd" ]] && print -z "$cmd"
}
```

Note that `_SHAI_BIN` from the development version has been removed — in the installed version `_shai_bin` is on `$PATH` and is called directly. If you want the same file to work for both local dev and installed use, you can keep a fallback at the top of the function:

```zsh
# Resolve binary: prefer local dev build, fall back to installed
local bin="${SHAI_DEV_BIN:-_shai_bin}"
```

Then replace `_shai_bin` calls in the function body with `"$bin"`.

---

## Step 1c: Understand how Homebrew auto-sources shell files

Homebrew installs packages under a prefix (e.g. `/opt/homebrew` on Apple Silicon, `/usr/local` on Intel Mac, `/home/linuxbrew/.linuxbrew` on Linux). Two directories under that prefix handle auto-loading shell code with no user action required:

**Bash** — `$(brew --prefix)/etc/profile.d/` is sourced by bash login shells automatically. macOS Terminal.app opens login shells by default, so this just works. On Linux, users are expected to have `eval "$(brew shellenv)"` in their `.bash_profile`, which triggers `profile.d` sourcing.

**Zsh** — `$(brew --prefix)/share/zsh/site-functions/` is on zsh's `$fpath`. Files here are available for zsh autoloading. The autoload mechanism works as follows: the file must be named exactly the function name with **no extension** (i.e. `shai`, not `shai.zsh`), and the file contains the function body. When `shai` is first invoked, zsh loads the file and executes it.

However, `shai.zsh` also defines `_SHAI_CMD_FILE_PREFIX` outside the function body. Pure autoload files only define a single function. The clean solution is to install `shai.zsh` to `etc/profile.d/` as well (zsh sources `profile.d` when `brew shellenv` is active), rather than using the autoload mechanism. This makes bash and zsh consistent: both are sourced at shell startup from `profile.d/`, both define the `shai` function, and neither requires any manual user setup.

The formula therefore installs:

| File | Installed to | Mechanism |
|---|---|---|
| `shell/shai.sh` | `etc/profile.d/shai.sh` | sourced by bash login shells |
| `shell/shai.zsh` | `etc/profile.d/shai.zsh` | sourced by zsh (via `brew shellenv`) |
| `_shai_bin` binary | `bin/_shai_bin` | on `$PATH` |

> **Why not `etc/bash_completion.d/`?** That directory is for bash completions only. Files there are sourced only when bash-completion is installed and configured. `etc/profile.d/` is the correct location for general-purpose shell functions.

---

## Step 2: Create a Homebrew tap repository

A Homebrew tap is a GitHub repository named `homebrew-<name>`. GoReleaser will push the generated formula to it automatically.

1. Create a new **public** GitHub repository named `homebrew-tap` (or any name you like — the `homebrew-` prefix is required by Homebrew).
2. Initialize it with a `README.md` so it's non-empty.
3. Create the `Formula/` directory by adding a placeholder file (GoReleaser will populate it on first release):

```
homebrew-tap/
  README.md
  Formula/
    .gitkeep
```

Users will install your tool with:

```bash
brew tap yourname/tap
brew install shai
# or combined:
brew install yourname/tap/shai
```

---

## Step 3: Create a GitHub Personal Access Token for the tap

GoReleaser needs permission to push to the tap repository (which is separate from your main CLI repo).

1. Go to **GitHub → Settings → Developer settings → Personal access tokens → Tokens (classic)**.
2. Click **Generate new token (classic)**.
3. Give it a descriptive name like `goreleaser-tap`.
4. Select the `repo` scope (full repository access).
5. Click **Generate token** and copy it immediately — you won't see it again.

Next, add it as a secret to your CLI repository:

1. Go to your CLI repo → **Settings → Secrets and variables → Actions**.
2. Click **New repository secret**.
3. Name it `TAP_TOKEN` and paste the token value.

The default `GITHUB_TOKEN` that Actions provides automatically is scoped only to the current repo, so GoReleaser can't use it to push to the tap repo. The `TAP_TOKEN` PAT is what enables cross-repo writes.

---

## Step 4: Create `.goreleaser.yaml`

Add this file at the root of your CLI repository. The key differences from a standard setup are the binary name (`_shai_bin`) and the `extra_files` block that bundles the shell wrapper files into the release archives.

```yaml
# .goreleaser.yaml
version: 2

before:
  hooks:
    - go mod tidy

builds:
  - binary: _shai_bin         # internal binary; users invoke via the `shai` shell function
    env:
      - CGO_ENABLED=0         # static binary, no C deps
    goos: [linux, darwin]     # no Windows — shell functions don't apply there
    goarch: [amd64, arm64]
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.date={{.Date}}

archives:
  - format: tar.gz
    name_template: "{{.ProjectName}}_{{.Os}}_{{.Arch}}"
    files:
      - shell/shai.sh         # include shell wrappers in the archive
      - shell/shai.zsh

checksum:
  name_template: "checksums.txt"

release:
  github:
    owner: yourname           # your GitHub username or org
    name: shai                # your repo name

brews:
  - name: shai
    repository:
      owner: yourname         # owner of the tap repo
      name: homebrew-tap      # tap repo name
      token: "{{ .Env.HOMEBREW_TAP_GITHUB_TOKEN }}"
    homepage: "https://github.com/yourname/shai"
    description: "Short description of what shai does"
    license: "MIT"
    # GoReleaser generates the url/sha256 blocks automatically.
    # The install block below is custom Ruby that we supply verbatim.
    install: |
      bin.install "_shai_bin"
      (etc/"profile.d").install "shai.sh"
      (etc/"profile.d").install "shai.zsh"
    test: |
      system "#{bin}/_shai_bin --version"

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - "^chore:"
```

Key fields to understand:

- **`binary: _shai_bin`** — the compiled binary is named `_shai_bin`. Homebrew puts it on `$PATH`, so the shell wrappers can call it by name without any hardcoded paths.
- **`archives.files`** — GoReleaser includes these extra files in every release archive alongside the binary. The formula's `install` block can then reference them by filename.
- **`install` block** — this is verbatim Ruby inserted into the generated formula. It installs three things: the binary into `bin/`, the bash wrapper into `etc/profile.d/`, and the zsh wrapper also into `etc/profile.d/`. Both shell files are sourced at login shell startup via the `profile.d` mechanism — consistent across bash and zsh.
- **`CGO_ENABLED=0`** produces fully static binaries, which is what you want for cross-platform distribution.
- **`ldflags: -s -w`** strips debug info and DWARF data, shrinking binary size significantly.
- **`brews.repository.token`** references the `HOMEBREW_TAP_GITHUB_TOKEN` environment variable, which your Actions workflow will supply from the `TAP_TOKEN` secret.
- **`changelog.filters.exclude`** keeps the generated changelog clean by dropping commit prefixes that aren't user-facing.

---

## Step 5: Test the build locally with `--snapshot`

Before pushing any tags, verify GoReleaser can build your project successfully. The `--snapshot` flag builds everything without requiring a git tag and without publishing anything.

```bash
goreleaser release --snapshot --clean
```

After running, inspect the `dist/` directory — you should see binaries for each OS/arch combination, archives, and a `checksums.txt`. If anything fails here (missing files, bad ldflags paths, import errors), fix it before continuing.

You can also check your config for syntax errors without building:

```bash
goreleaser check
```

---

## Step 6: Create the GitHub Actions release workflow

Add this file to your repository. It triggers automatically whenever you push a tag matching `v*`.

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write   # required to create GitHub Releases

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0    # goreleaser needs full git history for changelog

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: stable

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          HOMEBREW_TAP_GITHUB_TOKEN: ${{ secrets.TAP_TOKEN }}
```

The `fetch-depth: 0` is critical. Without it, Actions does a shallow clone, and GoReleaser can't walk the commit history to find the previous tag for changelog generation.

---

## Step 7: Tag a release

With everything in place, releasing is a one-liner:

```bash
# Make sure your working tree is clean and on the right commit
git tag -a v0.1.0 -m "release: initial release"
git push origin v0.1.0
```

This triggers the Actions workflow. You can monitor it under the **Actions** tab in your GitHub repo. A successful run will:

1. Build binaries for all configured OS/arch pairs.
2. Create archives and a `checksums.txt`.
3. Publish a GitHub Release with all assets attached.
4. Commit and push an updated `Formula/shai.rb` to your tap repo.

---

## Step 8: Verify the Homebrew formula

Once the release workflow completes, check your tap repo — GoReleaser should have pushed a `Formula/shai.rb` that looks roughly like this:

```ruby
class Shai < Formula
  desc "Short description of what shai does"
  homepage "https://github.com/yourname/shai"
  version "0.1.0"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/yourname/shai/releases/download/v0.1.0/shai_darwin_arm64.tar.gz"
      sha256 "abc123..."
    end
    on_intel do
      url "https://github.com/yourname/shai/releases/download/v0.1.0/shai_darwin_amd64.tar.gz"
      sha256 "def456..."
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/yourname/shai/releases/download/v0.1.0/shai_linux_arm64.tar.gz"
      sha256 "..."
    end
    on_intel do
      url "https://github.com/yourname/shai/releases/download/v0.1.0/shai_linux_amd64.tar.gz"
      sha256 "..."
    end
  end

  def install
    bin.install "_shai_bin"
    (etc/"profile.d").install "shai.sh"
    (etc/"profile.d").install "shai.zsh"
  end

  test do
    system "#{bin}/_shai_bin --version"
  end
end
```

The three install lines are the critical part:

- `bin.install "_shai_bin"` puts the Go binary on `$PATH`.
- `(etc/"profile.d").install "shai.sh"` places the bash wrapper where bash login shells auto-source it.
- `(etc/"profile.d").install "shai.zsh"` places the zsh wrapper in the same directory — zsh sources `profile.d` files when `brew shellenv` is active, which all Homebrew users have configured.

Both shell files define the `shai()` function. After a new login shell starts, `shai` is available immediately with no user action.

Test it yourself:

```bash
brew tap yourname/tap
brew install shai
# Open a new terminal (to get a fresh login shell), then:
shai --help
```

You can verify the files landed in the right places:

```bash
ls $(brew --prefix)/etc/profile.d/shai.sh
ls $(brew --prefix)/etc/profile.d/shai.zsh
which _shai_bin
```

---

## Step 9: Releasing subsequent versions

For every future release, just tag and push:

```bash
git tag -a v1.2.3 -m "release: add --watch flag"
git push origin v1.2.3
```

GoReleaser overwrites the formula in the tap repo on each release, so users running `brew upgrade shai` will get the latest binary and the latest shell wrappers automatically.

---

## Notes and gotchas

**Shell wrapper updates reach users automatically.** Because `shai.sh` and `shai.zsh` are bundled in the release archive and installed by the formula's `install` block, running `brew upgrade shai` replaces both the binary and the shell files. Users don't need to do anything extra.

**Why both shell files go to `etc/profile.d/`, not `share/zsh/site-functions/`.** Zsh's autoload mechanism (`share/zsh/site-functions/`) expects a file named exactly the function name with no extension, containing only the function body — no outer `function(){}` wrapper, no variable definitions outside the function. `shai.zsh` defines `_SHAI_CMD_FILE_PREFIX` at the top level, which makes it incompatible with pure autoloading. Installing to `etc/profile.d/` instead sources the entire file at shell startup, which handles top-level variable definitions correctly and is simpler to reason about. Zsh sources `profile.d` files when `brew shellenv` is active, which is a prerequisite for using any Homebrew-installed tool.

**Users need to open a new terminal after install.** The shell files are sourced at login shell startup. After `brew install shai`, the current shell session won't have the function yet — opening a new terminal window picks it up. This is standard behavior for any tool installed this way (same as `fzf`, `zoxide`, etc.) and does not require documentation for the user beyond the Homebrew install success message.

**Bash: `etc/profile.d/` is for login shells.** On macOS, Terminal.app opens login shells by default, so this works out of the box. On Linux, users running non-login interactive bash shells (some terminal emulators default to this) may not see the function. They can add `source $(brew --prefix)/etc/profile.d/shai.sh` to their `.bashrc` manually. This is a general limitation of the `profile.d` mechanism.

**The bash and zsh implementations use different flags intentionally.** Bash passes `--copy` so the binary writes the suggested command to the clipboard. Zsh passes `--cmd-file` so the binary writes the command to a temp file, which the function reads and stages in the ZLE buffer with `print -z`. This is the correct architecture: `print -z` only works in the user's interactive shell process, which a subprocess (script) cannot access.

**Pre-releases.** If your tag contains a pre-release identifier (`v1.0.0-beta.1`), GoReleaser marks the GitHub release as a pre-release and by default skips updating the Homebrew formula. You can override this with `prerelease: auto` in your `release:` config.

**First-time formula path.** The `Formula/` directory in your tap repo must exist before GoReleaser tries to push. The `.gitkeep` placeholder from Step 2 handles this.

**GoReleaser Pro.** The free tier covers everything in this guide. Pro adds features like signing binaries with cosign, Homebrew cask support, Scoop manifests for Windows, and Docker image publishing.

**Local builds with `goreleaser build`.** During development, `goreleaser build --single-target --snapshot` builds only the binary for your current OS/arch quickly, without producing archives or touching any remote services.

### Project-Specific Info

Homebrew Tap Repo is https://github.com/ameb8/homebrew-tools
This prooject's repo is https://github.com/ameb8/shai

## Additional Instructions

Only implement the files/logic/changes that happen inside this repo. Files in the homebrew tab repository will be managed in that repo, as well as github personall access tokens. But everything in this document internal to the shai repository should be implemented.

## Repo-Local Implementation Plan

1. Add build-time version metadata variables in `main.go` and wire them into the Cobra root command so `--version` works for GoReleaser and Homebrew formula tests.
2. Update the shell wrapper scripts in `scripts/` to call `_shai_bin` from `PATH` by default, while retaining `SHAI_DEV_BIN` as a local-development override.
3. Add a root `.goreleaser.yaml` that builds `_shai_bin` for Darwin/Linux amd64/arm64, bundles the `scripts/shai.sh` and `scripts/shai.zsh` wrappers, publishes GitHub release assets for `ameb8/shai`, and generates a Homebrew formula targeting `ameb8/homebrew-tools`.
4. Add `.github/workflows/release.yml` so pushed `v*` tags run GoReleaser with repository `GITHUB_TOKEN` and the externally-managed `TAP_TOKEN` secret.
5. Run formatting and tests to verify the existing query flow and new version path remain intact.
