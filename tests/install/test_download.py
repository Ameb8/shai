"""
Module: test_download.py

Verify the download → extract → install pipeline.
All network calls are directed to the fake_release_server.
"""

import stat
import subprocess
from pathlib import Path
from typing import Any

import pytest
from helpers import run_install


@pytest.fixture()
def installed(
    fake_env: dict[str, str], fake_release_server: dict[str, Any]
) -> tuple[subprocess.CompletedProcess[str], Path]:
    """
    Run install.sh once using the fake environment and release server.

    Args:
        fake_env: Environment variables for the test, including HOME.
        fake_release_server: Mock server configuration and version information.

    Returns:
        tuple[subprocess.CompletedProcess[str], Path]: The completed installation
            process and the HOME directory path.
    """
    result = run_install(fake_env, fake_release_server, shell="bash")
    return result, Path(fake_env["HOME"])


def test_install_exits_zero(
    installed: tuple[subprocess.CompletedProcess[str], Path],
) -> None:
    """
    Verify that the installation script completes successfully.

    Args:
        installed: The installation result and home path fixture.
    """
    result, _ = installed
    assert result.returncode == 0, (
        f"install.sh failed:\nstdout: {result.stdout}\nstderr: {result.stderr}"
    )


def test_binary_placed_at_correct_path(
    installed: tuple[subprocess.CompletedProcess[str], Path],
) -> None:
    """
    Check if the binary is located in the expected directory.

    Args:
        installed: The installation result and home path fixture.
    """
    _, home = installed
    binary = home / ".local" / "bin" / "_shai_bin"
    assert binary.exists(), f"_shai_bin not found at {binary}"


def test_binary_is_executable(
    installed: tuple[subprocess.CompletedProcess[str], Path],
) -> None:
    """
    Confirm the installed binary has executable permissions.

    Args:
        installed: The installation result and home path fixture.
    """
    _, home = installed
    binary = home / ".local" / "bin" / "_shai_bin"
    assert binary.exists()
    assert binary.stat().st_mode & stat.S_IEXEC, "Binary is not executable"


def test_binary_is_not_a_directory(
    installed: tuple[subprocess.CompletedProcess[str], Path],
) -> None:
    """
    Ensure the installed path is a file and not a directory.

    Args:
        installed: The installation result and home path fixture.
    """
    _, home = installed
    binary = home / ".local" / "bin" / "_shai_bin"
    assert binary.is_file(), "_shai_bin exists but is a directory"


def test_version_printed_in_output(
    installed: tuple[subprocess.CompletedProcess[str], Path],
    fake_release_server: dict[str, Any],
) -> None:
    """
    Verify that the correct version is printed during installation.

    Args:
        installed: The installation result and home path fixture.
        fake_release_server: Mock server configuration containing the expected version.
    """
    result, _ = installed
    assert fake_release_server["version"] in result.stdout
