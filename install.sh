#!/usr/bin/env bash
# Exit immediately on error (-e), treat unset variables as errors (-u),
# and propagate pipe failures (-o pipefail) so a failed curl|grep doesn't silently pass
set -euo pipefail

# ── Constants ────────────────────────────────────────────────────────────────
# GitHub repo in "owner/name" format — used to build API and download URLs
REPO="ameb8/shai"

# Where the Go binary will live; ~/.local/bin is the standard user-local bin
# directory that doesn't require sudo
INSTALL_DIR="${HOME}/.local/bin"

# Where the shell wrapper scripts will live; keeping them in their own config
# dir makes it easy to find, update, or uninstall them later
CONFIG_DIR="${HOME}/.config/shai"

# The source lines that get appended to the user's rc file — one per shell.
# Single-quoted so the ${HOME} inside is written literally into the rc file
# and evaluated at shell startup time, not at install time
WRAPPER_SOURCE_LINE='source "${HOME}/.config/shai/shai.sh"'
WRAPPER_SOURCE_LINE_ZSH='source "${HOME}/.config/shai/shai.zsh"'

# ── Detect latest version ────────────────────────────────────────────────────
# Query the GitHub releases API and parse the tag_name field from the JSON.
# Use grep + sed instead of jq to avoid requiring jq as a dependency
VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name"' | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')
echo "Installing shai ${VERSION}..."

# ── Detect architecture ──────────────────────────────────────────────────────
# uname -m returns the machine hardware name (e.g. x86_64, aarch64).
# Map hardware name to naming convention used in the GitHub release tarballs
ARCH=$(uname -m)
case "${ARCH}" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  *)
    # Fail fast if unsupported architecture
    echo "Unsupported architecture: ${ARCH}" >&2
    exit 1
    ;;
esac

# ── Download & extract ───────────────────────────────────────────────────────
# Create a temporary directory for download and extraction.
# Using mktemp ensures a unique, collision-free path even if multiple installs run concurrently
TMP=$(mktemp -d)

# Register a cleanup trap so the temp dir is always removed when the script
# exits — whether that's a clean exit, an error, or a Ctrl-C
trap 'rm -rf "${TMP}"' EXIT

# Build the tarball filename and full download URL from the detected values
TARBALL="shai_linux_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${TARBALL}"

echo "Downloading ${URL}..."
# -f: fail on HTTP errors, -s: silent, -S: show errors, -L: follow redirects
curl -fsSL "${URL}" -o "${TMP}/${TARBALL}"

# Extract into the temp dir; all three files (_shai_bin, shai.sh, shai.zsh)
# will land flat in $TMP
tar -xzf "${TMP}/${TARBALL}" -C "${TMP}"

# ── Install binary ───────────────────────────────────────────────────────────
# Create the bin directory if it doesn't exist yet (-p suppresses errors if it does)
mkdir -p "${INSTALL_DIR}"

# Move the Go binary into place
mv "${TMP}/_shai_bin" "${INSTALL_DIR}/_shai_bin"

# Ensure the binary is executable; the tarball may not have preserved permissions
chmod +x "${INSTALL_DIR}/_shai_bin"

# Prepend INSTALL_DIR to PATH for the remainder of this script session so any
# subsequent commands that invoke _shai_bin directly can find it
export PATH="${INSTALL_DIR}:${PATH}"

# ── Install shell wrappers ───────────────────────────────────────────────────
# Create the config dir if it doesn't already exist
mkdir -p "${CONFIG_DIR}"

# Move both wrapper scripts into the config dir regardless of the user's shell —
# both are cheap to store and having them ready avoids a re-install if the user
# switches shells later
mv "${TMP}/shai.sh"  "${CONFIG_DIR}/shai.sh"
mv "${TMP}/shai.zsh" "${CONFIG_DIR}/shai.zsh"

# ── Wire up rc files (idempotent) ────────────────────────────────────────────
# Appends a source line to an rc file only if it isn't already present.
# Takes two arguments: the rc file path and the line to append.
# This makes the installer safe to re-run (e.g. after an update) without
# accumulating duplicate source lines
add_source_line() {
  local rc_file="$1"
  local line="$2"

  # grep -qF: quiet, fixed-string (no regex) match — returns 0 if found
  if [[ -f "${rc_file}" ]] && grep -qF "${line}" "${rc_file}"; then
    echo "  ${rc_file} already configured, skipping."
    return
  fi

  # Blank line for readability in the rc file, then a comment and the source line
  echo "" >> "${rc_file}"
  echo "# shai shell integration" >> "${rc_file}"
  echo "${line}" >> "${rc_file}"
  echo "  Added source line to ${rc_file}"
}

# $SHELL holds the path to the user's login shell (e.g. /bin/zsh).
# basename strips the path, leaving just the shell name for the case match
CURRENT_SHELL=$(basename "${SHELL}")
case "${CURRENT_SHELL}" in
  zsh)
    add_source_line "${HOME}/.zshrc" "${WRAPPER_SOURCE_LINE_ZSH}"
    ;;
  bash)
    add_source_line "${HOME}/.bashrc" "${WRAPPER_SOURCE_LINE}"
    ;;
  *)
    # Shell isn't bash or zsh — write both rc files as a best-effort fallback.
    # The extra source line in the wrong file is harmless since the file it
    # points to will still exist; the user can clean it up if they want
    echo "Unknown shell '${CURRENT_SHELL}', writing to both rc files just in case."
    add_source_line "${HOME}/.bashrc" "${WRAPPER_SOURCE_LINE}"
    add_source_line "${HOME}/.zshrc"  "${WRAPPER_SOURCE_LINE_ZSH}"
    ;;
esac

# ── Done ─────────────────────────────────────────────────────────────────────
echo ""
echo "shai ${VERSION} installed successfully!"
echo ""
# Remind the user they need to reload their rc file for the source line to
# take effect in their current session; new sessions will pick it up automatically
echo "Restart your shell or run:"
case "${CURRENT_SHELL}" in
  zsh)  echo "  source ~/.zshrc" ;;
  bash) echo "  source ~/.bashrc" ;;
  *)    echo "  source ~/.bashrc  # or ~/.zshrc" ;;
esac