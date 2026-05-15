"""
Shared fixtures for install.sh tests.

Design principles:
- fake_env: hermetic HOME/PATH/SHELL replacement via tmp_path; never touches real FS
- fake_release_server: local HTTP server that serves a synthetic tarball,
  replacing both the GitHub API endpoint and the download URL
- run_install: defined in helpers.py; accepts a fake_env and optionally
  overrides SHELL to test bash vs zsh paths
"""

import os
import stat
import tarfile
import textwrap
import threading
from collections.abc import Generator
from http.server import BaseHTTPRequestHandler, HTTPServer
from pathlib import Path
from typing import Any

import pytest

# ── Fake home environment ────────────────────────────────────────────────────


@pytest.fixture()
def fake_env(tmp_path: Path) -> dict[str, str]:
    """
    Returns a dict of env vars that redirect all install.sh writes into tmp_path.

    Structure created:
        tmp_path/home/           <- $HOME
        tmp_path/home/.bashrc    <- pre-existing, contains "# existing bashrc"
        tmp_path/home/.zshrc     <- pre-existing, contains "# existing zshrc"
        tmp_path/bin/            <- prepended to $PATH so any tool lookups resolve

    The fixture does NOT set $SHELL; individual tests set it to control which
    rc file the script targets.
    """
    home = tmp_path / "home"
    home.mkdir()
    (home / ".bashrc").write_text("# existing bashrc\n")
    (home / ".zshrc").write_text("# existing zshrc\n")

    bin_dir = tmp_path / "bin"
    bin_dir.mkdir()

    env = os.environ.copy()
    env["HOME"] = str(home)
    env["PATH"] = f"{bin_dir}:{env.get('PATH', '')}"
    # Unset any caller-side overrides that could leak into the subprocess
    env.pop("XDG_CONFIG_HOME", None)
    return env


# ── Fake GitHub release HTTP server ─────────────────────────────────────────


def _build_fake_tarball(tmp_path: Path) -> bytes:
    """
    Builds an in-memory tarball that mirrors the structure install.sh expects
    after extracting shai_linux_amd64.tar.gz:

        _shai_bin   <- the Go binary (fake: a tiny shell script that exits 0)
        shai.sh     <- bash wrapper
        shai.zsh    <- zsh wrapper

    Returns the raw tarball bytes.
    """
    staging = tmp_path / "_tarball_staging"
    staging.mkdir()

    binary = staging / "_shai_bin"
    binary.write_text("#!/usr/bin/env bash\nexit 0\n")
    binary.chmod(binary.stat().st_mode | stat.S_IEXEC)

    (staging / "shai.sh").write_text(
        textwrap.dedent("""\
            #!/usr/bin/env bash
            # fake shai bash wrapper
            shai() { _shai_bin "$@"; }
        """)
    )
    (staging / "shai.zsh").write_text(
        textwrap.dedent("""\
            #!/usr/bin/env zsh
            # fake shai zsh wrapper
            shai() { _shai_bin "$@"; }
        """)
    )

    archive_path = tmp_path / "shai_linux_amd64.tar.gz"
    with tarfile.open(archive_path, "w:gz") as tar:
        for f in staging.iterdir():
            tar.add(f, arcname=f.name)

    return archive_path.read_bytes()


@pytest.fixture(scope="session")
def fake_release_server(
    tmp_path_factory: pytest.TempPathFactory,
) -> Generator[dict[str, Any], None, None]:
    """
    Starts a local HTTP server that:
      - Responds to GET /repos/ameb8/shai/releases/latest with a JSON body
        containing tag_name "v0.0.0-test"
      - Responds to GET /ameb8/shai/releases/download/v0.0.0-test/shai_linux_amd64.tar.gz
        with the fake tarball bytes

    Both the GitHub API host and the download host are intercepted by patching
    install.sh at runtime (see helpers.run_install).

    Returns a dict: {"api_url": str, "download_url_base": str, "host": str, "port": int}
    """
    tarball_bytes = _build_fake_tarball(tmp_path_factory.mktemp("tarball"))
    version_tag = "v0.0.0-test"

    class Handler(BaseHTTPRequestHandler):
        def log_message(
            self, *args: Any
        ) -> None:  # suppress access log noise in test output
            pass

        def do_GET(self) -> None:
            if self.path.endswith("/releases/latest"):
                body = f'{{"tag_name": "{version_tag}"}}'.encode()
                self.send_response(200)
                self.send_header("Content-Type", "application/json")
                self.send_header("Content-Length", str(len(body)))
                self.end_headers()
                self.wfile.write(body)
            elif self.path.endswith(".tar.gz"):
                self.send_response(200)
                self.send_header("Content-Type", "application/octet-stream")
                self.send_header("Content-Length", str(len(tarball_bytes)))
                self.end_headers()
                self.wfile.write(tarball_bytes)
            else:
                self.send_response(404)
                self.end_headers()

    server = HTTPServer(("127.0.0.1", 0), Handler)
    port = server.server_address[1]
    thread = threading.Thread(target=server.serve_forever, daemon=True)
    thread.start()

    yield {
        "host": "127.0.0.1",
        "port": port,
        "version": version_tag,
        "base_url": f"http://127.0.0.1:{port}",
    }

    server.shutdown()
