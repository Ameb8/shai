# Temporary file path for receiving the final command from the binary.
_SHAI_CMD_FILE_PREFIX="shai_cmd"

# Path to binary executable
_SHAI_BIN="./bin/_shai_bin"


shai() {
    # Create a temp file for the binary to write the final command into
    local cmd_file
    mkdir -p "${TMPDIR:-/tmp}"
    cmd_file=$(mktemp "${TMPDIR:-/tmp}/${_SHAI_CMD_FILE_PREFIX}.XXXXXX") || {
        # Error creating temp file
        echo "Failed to create temp file" >&2
        return 1
    }

    # Run the binary, passing all user arguments plus the temp file path.
    # Stdout streams directly to the terminal so explanation lines appear in real time.
    # Stderr is left untouched — the binary can use it freely for actual errors.
    "$_SHAI_BIN" --cmd-file="$cmd_file" "$@"
    local exit_code=$?

    # If the binary failed, clean up and propagate the exit code
    if [[ $exit_code -ne 0 ]]; then
    rm -f "$cmd_file"
    return $exit_code
    fi

    # Read the command the binary wrote to the temp file, then clean up
    local cmd
    cmd=$(< "$cmd_file")
    rm -f "$cmd_file"

    # Stage the command in the zsh line editor buffer without printing it.
    # The user can then edit, execute, or discard it as they see fit.
    [[ -n "$cmd" ]] && print -z "$cmd"
}
