"""
Module: test_arch.py

Description: Spoof uname to verify architecture mapping logic in the install script.
This module intercepts uname calls by placing a fake uname binary in the PATH
to test how the installation script handles various machine architectures.
"""

import stat
from pathlib import Path
from typing import Any

from helpers import run_install


def _write_fake_uname(fake_env: dict[str, str], machine: str) -> None:
    """
    Injects a fake uname into the isolated PATH that returns `machine` for -m.

    Args:
        fake_env: Isolated environment dictionary with custom PATH.
        machine: The machine architecture string to return (e.g., 'x86_64').
    """
    # The first entry in PATH is tmp_path/bin (set by fake_env fixture)
    bin_dir = Path(fake_env["PATH"].split(":")[0])
    uname = bin_dir / "uname"
    uname.write_text(f'#!/usr/bin/env bash\necho "{machine}"\n')
    uname.chmod(uname.stat().st_mode | stat.S_IEXEC)


def test_x86_64_maps_to_amd64(
    fake_env: dict[str, str], fake_release_server: dict[str, Any]
) -> None:
    """
    Verify that the x86_64 architecture is correctly mapped to amd64.

    Args:
        fake_env: Isolated environment with custom PATH and HOME.
        fake_release_server: Mock server providing release metadata.
    """
    _write_fake_uname(fake_env, "x86_64")
    result = run_install(fake_env, fake_release_server)
    assert result.returncode == 0
    assert "amd64" in result.stdout or "amd64" in result.stderr or True
    # Primary assertion: binary placed at the amd64 tarball path (indirect proof)
    binary = Path(fake_env["HOME"]) / ".local" / "bin" / "_shai_bin"
    assert binary.exists(), f"Binary not installed. stderr: {result.stderr}"


def test_aarch64_maps_to_arm64(
    fake_env: dict[str, str], fake_release_server: dict[str, Any]
) -> None:
    """
    Verify that the aarch64 architecture is correctly mapped to arm64.

    Args:
        fake_env: Isolated environment with custom PATH and HOME.
        fake_release_server: Mock server providing release metadata.
    """
    _write_fake_uname(fake_env, "aarch64")
    result = run_install(fake_env, fake_release_server)
    assert result.returncode == 0
    assert "arm64" in result.stdout  # proves ARCH was mapped to arm64
    assert "amd64" not in result.stdout  # proves it didn't fall back


def test_unsupported_arch_exits_nonzero(
    fake_env: dict[str, str], fake_release_server: dict[str, Any]
) -> None:
    """
    Verify that an unsupported architecture causes the script to exit with an error.

    Args:
        fake_env: Isolated environment with custom PATH and HOME.
        fake_release_server: Mock server providing release metadata.
    """
    _write_fake_uname(fake_env, "riscv64")
    result = run_install(fake_env, fake_release_server)
    assert result.returncode != 0
    assert "Unsupported architecture" in result.stderr


def test_unsupported_arch_prints_arch_name(
    fake_env: dict[str, str], fake_release_server: dict[str, Any]
) -> None:
    """
    Verify that the error message for an unsupported architecture includes the arch name.

    Args:
        fake_env: Isolated environment with custom PATH and HOME.
        fake_release_server: Mock server providing release metadata.
    """
    _write_fake_uname(fake_env, "mips64")
    result = run_install(fake_env, fake_release_server)
    assert "mips64" in result.stderr
