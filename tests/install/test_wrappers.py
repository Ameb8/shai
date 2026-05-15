"""
test_wrappers.py

Verifies shell wrapper script installation across supported shell environments.
Ensures that both shai.sh and shai.zsh are correctly deployed regardless of
the detected or active user shell.
"""

import subprocess
from pathlib import Path
from typing import Any

import pytest
from helpers import run_install


@pytest.fixture()
def installed_bash(
    fake_env: dict[str, str], fake_release_server: dict[str, Any]
) -> tuple[subprocess.CompletedProcess[str], Path]:
    """
    Fixturizes a mock bash installation.

    Args:
        fake_env: The mocked environment dictionary.
        fake_release_server: The mocked release server response data.

    Returns:
        tuple: A tuple containing the install process result and the home directory path.
    """
    result = run_install(fake_env, fake_release_server, shell="bash")
    return result, Path(fake_env["HOME"])


@pytest.fixture()
def installed_zsh(
    fake_env: dict[str, str], fake_release_server: dict[str, Any]
) -> tuple[subprocess.CompletedProcess[str], Path]:
    """
    Fixturizes a mock zsh installation.

    Args:
        fake_env: The mocked environment dictionary.
        fake_release_server: The mocked release server response data.

    Returns:
        tuple: A tuple containing the install process result and the home directory path.
    """
    result = run_install(fake_env, fake_release_server, shell="zsh")
    return result, Path(fake_env["HOME"])


def test_shai_sh_exists_after_bash_install(
    installed_bash: tuple[subprocess.CompletedProcess[str], Path],
) -> None:
    """Verifies that shai.sh is installed when the active shell is bash."""
    _, home = installed_bash
    assert (home / ".config" / "shai" / "shai.sh").exists()


def test_shai_zsh_exists_after_bash_install(
    installed_bash: tuple[subprocess.CompletedProcess[str], Path],
) -> None:
    """Verifies that shai.zsh is installed even when the active shell is bash."""
    _, home = installed_bash
    assert (home / ".config" / "shai" / "shai.zsh").exists()


def test_shai_sh_exists_after_zsh_install(
    installed_zsh: tuple[subprocess.CompletedProcess[str], Path],
) -> None:
    """Verifies that shai.sh is installed even when the active shell is zsh."""
    _, home = installed_zsh
    assert (home / ".config" / "shai" / "shai.sh").exists()


def test_shai_zsh_exists_after_zsh_install(
    installed_zsh: tuple[subprocess.CompletedProcess[str], Path],
) -> None:
    """Verifies that shai.zsh is installed when the active shell is zsh."""
    _, home = installed_zsh
    assert (home / ".config" / "shai" / "shai.zsh").exists()


def test_config_dir_created(
    installed_bash: tuple[subprocess.CompletedProcess[str], Path],
) -> None:
    """Verifies that the required base configuration directory is created during install."""
    _, home = installed_bash
    assert (home / ".config" / "shai").is_dir()
