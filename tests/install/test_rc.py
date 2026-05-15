"""
Module: test_rc.py

Validates RC file modification behavior during the shai installation process.
Ensures that the correct shell configuration files are updated with the appropriate
source lines and that these operations are idempotent and shell-specific.
"""

from pathlib import Path
from typing import Any

from helpers import assert_source_line_absent, assert_source_line_present, run_install

# These must match the constants in install.sh exactly to ensure the integration
# logic correctly identifies and manages the shai shell scripts.
BASH_SOURCE_LINE = 'source "${HOME}/.config/shai/shai.sh"'
ZSH_SOURCE_LINE = 'source "${HOME}/.config/shai/shai.zsh"'


class TestBashRcConfiguration:
    """
    Validation logic for Bash shell configuration (.bashrc).
    """

    def test_source_line_added_to_bashrc(
        self, fake_env: dict[str, str], fake_release_server: dict[str, Any]
    ) -> None:
        """Verify that the bash source line is added to .bashrc when shell is bash."""
        run_install(fake_env, fake_release_server, shell="bash")
        bashrc = Path(fake_env["HOME"]) / ".bashrc"
        assert_source_line_present(bashrc, BASH_SOURCE_LINE)

    def test_zshrc_not_modified_when_shell_is_bash(
        self, fake_env: dict[str, str], fake_release_server: dict[str, Any]
    ) -> None:
        """Ensure .zshrc remains untouched when installing for bash."""
        run_install(fake_env, fake_release_server, shell="bash")
        zshrc = Path(fake_env["HOME"]) / ".zshrc"
        assert_source_line_absent(zshrc, ZSH_SOURCE_LINE)
        assert_source_line_absent(zshrc, BASH_SOURCE_LINE)

    def test_integration_comment_written_to_bashrc(
        self, fake_env: dict[str, str], fake_release_server: dict[str, Any]
    ) -> None:
        """Verify that the shell integration header comment is present in .bashrc."""
        run_install(fake_env, fake_release_server, shell="bash")
        content = (Path(fake_env["HOME"]) / ".bashrc").read_text()
        assert "# shai shell integration" in content

    def test_idempotent_bash_rc(
        self, fake_env: dict[str, str], fake_release_server: dict[str, Any]
    ) -> None:
        """Running install.sh twice must not produce duplicate source lines in .bashrc."""
        run_install(fake_env, fake_release_server, shell="bash")
        run_install(fake_env, fake_release_server, shell="bash")
        bashrc = Path(fake_env["HOME"]) / ".bashrc"
        # assert_source_line_present checks for exactly 1 occurrence
        assert_source_line_present(bashrc, BASH_SOURCE_LINE)


class TestZshRcConfiguration:
    """
    Validation logic for Zsh shell configuration (.zshrc).
    """

    def test_source_line_added_to_zshrc(
        self, fake_env: dict[str, str], fake_release_server: dict[str, Any]
    ) -> None:
        """Verify that the zsh source line is added to .zshrc when shell is zsh."""
        run_install(fake_env, fake_release_server, shell="zsh")
        zshrc = Path(fake_env["HOME"]) / ".zshrc"
        assert_source_line_present(zshrc, ZSH_SOURCE_LINE)

    def test_bashrc_not_modified_when_shell_is_zsh(
        self, fake_env: dict[str, str], fake_release_server: dict[str, Any]
    ) -> None:
        """Ensure .bashrc remains untouched when installing for zsh."""
        run_install(fake_env, fake_release_server, shell="zsh")
        bashrc = Path(fake_env["HOME"]) / ".bashrc"
        assert_source_line_absent(bashrc, BASH_SOURCE_LINE)
        assert_source_line_absent(bashrc, ZSH_SOURCE_LINE)

    def test_integration_comment_written_to_zshrc(
        self, fake_env: dict[str, str], fake_release_server: dict[str, Any]
    ) -> None:
        """Verify that the shell integration header comment is present in .zshrc."""
        run_install(fake_env, fake_release_server, shell="zsh")
        content = (Path(fake_env["HOME"]) / ".zshrc").read_text()
        assert "# shai shell integration" in content

    def test_idempotent_zsh_rc(
        self, fake_env: dict[str, str], fake_release_server: dict[str, Any]
    ) -> None:
        """Running install.sh twice must not produce duplicate source lines in .zshrc."""
        run_install(fake_env, fake_release_server, shell="zsh")
        run_install(fake_env, fake_release_server, shell="zsh")
        zshrc = Path(fake_env["HOME"]) / ".zshrc"
        assert_source_line_present(zshrc, ZSH_SOURCE_LINE)


class TestUnknownShellFallback:
    """
    Validation logic for cases where the current shell is not explicitly supported.
    """

    def test_unknown_shell_writes_both_rc_files(
        self, fake_env: dict[str, str], fake_release_server: dict[str, Any]
    ) -> None:
        """
        When the shell is unknown, the installer should attempt to update both
        .bashrc and .zshrc as a safety measure.
        """
        run_install(fake_env, fake_release_server, shell="fish")
        home = Path(fake_env["HOME"])
        assert_source_line_present(home / ".bashrc", BASH_SOURCE_LINE)
        assert_source_line_present(home / ".zshrc", ZSH_SOURCE_LINE)
