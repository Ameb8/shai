shai() {
    local bin="${SHAI_DEV_BIN:-_shai_bin}"

    "$bin" --copy "$@"
}
