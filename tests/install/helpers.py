"""
Module: helpers.py

Shared test utilities for the shai installation script tests.

run_install(script_path, env, server_info, shell="bash", extra_env=None)
    Patches install.sh in memory to redirect curl to the fake server,
    writes the patched script to a temp file, and executes it via bash.

    Returns subprocess.CompletedProcess. Never raises on non-zero exit;
    callers assert on .returncode themselves.

assert_source_line_present(rc_path, expected_line)
    Asserts that expected_line appears exactly once in rc_path.

assert_source_line_absent(rc_path, line)
    Asserts that line does not appear in rc_path.
"""

import os
import subprocess
import tempfile
from pathlib import Path
from typing import Any

INSTALL_SCRIPT = Path(__file__).parent.parent.parent / "install.sh"


def _patch_script(source: str, server: dict[str, Any]) -> str:
    """
    Rewrites GitHub URL patterns inside install.sh to point to a local fake server.

    It rewrites two URL patterns:
    1. The GitHub API endpoint used to fetch the latest release version:
       https://api.github.com/repos/${REPO}/releases/latest
       → http://127.0.0.1:<port>/repos/${REPO}/releases/latest

    2. The download URL:
       https://github.com/${REPO}/releases/download/...
       → http://127.0.0.1:<port>/${REPO}/releases/download/...

    Uses string replacement on constant URL prefixes so the patch is robust
    to variable changes inside the script.

    Args:
        source: The raw content of the install.sh script.
        server: A dictionary containing 'host' and 'port' of the fake server.

    Returns:
        str: The patched script content.
    """
    base = f"http://{server['host']}:{server['port']}"
    patched = source.replace(
        '"https://api.github.com/',
        f'"{base}/',
    ).replace(
        '"https://github.com/',
        f'"{base}/',
    )
    return patched


def run_install(
    fake_env: dict[str, str],
    server: dict[str, Any],
    shell: str = "bash",
    extra_env: dict[str, str] | None = None,
) -> subprocess.CompletedProcess[str]:
    """
    Executes install.sh with all external URLs redirected to the fake server.

    Args:
        fake_env: Environment dictionary from the fake_env fixture (contains HOME, PATH).
        server: Dictionary from fake_release_server fixture with 'host' and 'port'.
        shell: The shell to simulate ($SHELL), either "bash" or "zsh".
        extra_env: Optional additional environment variable overrides.

    Returns:
        subprocess.CompletedProcess[str]: The result of the script execution.
    """
    source = INSTALL_SCRIPT.read_text()
    patched = _patch_script(source, server)

    env = fake_env.copy()
    env["SHELL"] = f"/bin/{shell}"
    if extra_env:
        env.update(extra_env)

    with tempfile.NamedTemporaryFile(mode="w", suffix=".sh", delete=False) as f:
        f.write(patched)
        tmp_path = f.name

    os.chmod(tmp_path, 0o755)

    try:
        result = subprocess.run(
            ["bash", tmp_path],
            env=env,
            capture_output=True,
            text=True,
        )
    finally:
        os.unlink(tmp_path)

    return result


def assert_source_line_present(rc_path: Path, line: str) -> None:
    """
    Asserts that a specific line appears exactly once in the given file.

    Args:
        rc_path: Path to the resource file (e.g., .bashrc or .zshrc).
        line: The exact string that should be present.

    Raises:
        AssertionError: If the line is not found or appears multiple times.
    """
    content = rc_path.read_text()
    occurrences = content.count(line)
    assert occurrences == 1, (
        f"Expected source line to appear exactly once in {rc_path}, "
        f"but found {occurrences} occurrences.\n"
        f"Line: {line!r}\n"
        f"File content:\n{content}"
    )


def assert_source_line_absent(rc_path: Path, line: str) -> None:
    """
    Asserts that a specific line does not appear in the given file.

    Args:
        rc_path: Path to the resource file.
        line: The exact string that should be absent.

    Raises:
        AssertionError: If the line is found in the file.
    """
    content = rc_path.read_text()
    assert line not in content, (
        f"Source line should be absent from {rc_path} but was found.\n"
        f"Line: {line!r}\n"
        f"File content:\n{content}"
    )
