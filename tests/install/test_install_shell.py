"""
Module: test_install_shell.py

Validates that install-shell.sh correctly configures shell environments.
It checks file existence, permissions, and correct modification of rc files.
"""

import stat
import subprocess
from pathlib import Path
from typing import Any

import pytest
from helpers import assert_source_line_present, run_install

BASH_SOURCE_LINE: str = 'source "${HOME}/.config/shai/shai.sh"'
ZSH_SOURCE_LINE: str = 'source "${HOME}/.config/shai/shai.zsh"'


@pytest.fixture()
def post_install(
    fake_env: dict[str, str], fake_release_server: dict[str, Any]
) -> tuple[dict[str, str], Path]:
    """
    Perform a successful bash install to set up the test environment.

    Args:
        fake_env: Dictionary containing simulated environment variables.
        fake_release_server: Mock server configuration for releases.

    Returns:
        tuple[dict[str, str], Path]: A tuple containing the environment dictionary and the HOME Path.
    """
    result = run_install(fake_env, fake_release_server, shell="bash")
    assert result.returncode == 0, f"Setup install failed: {result.stderr}"
    return fake_env, Path(fake_env["HOME"])


def test_install_shell_script_exists(post_install: tuple[dict[str, str], Path]) -> None:
    """Verify that install-shell.sh is created in the expected location."""
    _, home = post_install
    assert (home / ".config" / "shai" / "install-shell.sh").exists()


def test_install_shell_script_is_executable(
    post_install: tuple[dict[str, str], Path],
) -> None:
    """Ensure the installed shell script has execution permissions."""
    _, home = post_install
    script = home / ".config" / "shai" / "install-shell.sh"
    assert script.stat().st_mode & stat.S_IEXEC


def _run_install_shell(
    env: dict[str, str], home: Path, shell: str
) -> subprocess.CompletedProcess[str]:
    """
    Execute the install-shell.sh script with a specific shell configuration.

    Args:
        env: Environment variables for the subprocess.
        home: Path to the simulated home directory.
        shell: The shell type to simulate (e.g., 'bash', 'zsh').

    Returns:
        subprocess.CompletedProcess[str]: The result of the script execution.
    """
    script = home / ".config" / "shai" / "install-shell.sh"
    run_env = env.copy()
    run_env["SHELL"] = f"/bin/{shell}"
    return subprocess.run(
        ["bash", str(script)],
        env=run_env,
        capture_output=True,
        text=True,
    )


def test_install_shell_adds_bash_source_line(
    post_install: tuple[dict[str, str], Path],
) -> None:
    """Verify that the script correctly adds the source line to .bashrc."""
    env, home = post_install
    # Clear .bashrc to ensure a clean slate for the test.
    bashrc = home / ".bashrc"
    bashrc.write_text("# clean bashrc\n")

    result = _run_install_shell(env, home, "bash")
    assert result.returncode == 0, f"install-shell.sh failed: {result.stderr}"
    assert_source_line_present(bashrc, BASH_SOURCE_LINE)


def test_install_shell_adds_zsh_source_line(
    post_install: tuple[dict[str, str], Path],
) -> None:
    """Verify that the script correctly adds the source line to .zshrc."""
    env, home = post_install
    zshrc = home / ".zshrc"
    zshrc.write_text("# clean zshrc\n")

    result = _run_install_shell(env, home, "zsh")
    assert result.returncode == 0, f"install-shell.sh failed: {result.stderr}"
    assert_source_line_present(zshrc, ZSH_SOURCE_LINE)


def test_install_shell_idempotent_bash(
    post_install: tuple[dict[str, str], Path],
) -> None:
    """Check that running the script multiple times does not duplicate lines in .bashrc."""
    env, home = post_install
    bashrc = home / ".bashrc"
    bashrc.write_text("# clean bashrc\n")

    _run_install_shell(env, home, "bash")
    _run_install_shell(env, home, "bash")
    assert_source_line_present(bashrc, BASH_SOURCE_LINE)


def test_install_shell_idempotent_zsh(
    post_install: tuple[dict[str, str], Path],
) -> None:
    """Check that running the script multiple times does not duplicate lines in .zshrc."""
    env, home = post_install
    zshrc = home / ".zshrc"
    zshrc.write_text("# clean zshrc\n")

    _run_install_shell(env, home, "zsh")
    _run_install_shell(env, home, "zsh")
    assert_source_line_present(zshrc, ZSH_SOURCE_LINE)


def test_install_shell_does_not_touch_wrong_rc_for_bash(
    post_install: tuple[dict[str, str], Path],
) -> None:
    """Ensure that installing for bash doesn't modify the zsh configuration."""
    env, home = post_install
    zshrc = home / ".zshrc"
    original_zshrc = zshrc.read_text()
    # Ensure only the targeted rc file is potentially modified.
    (home / ".bashrc").write_text("# clean\n")

    _run_install_shell(env, home, "bash")
    assert zshrc.read_text() == original_zshrc, (
        ".zshrc was modified when shell was bash"
    )
