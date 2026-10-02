# Gamaj Panel is.0.0.1

First release of Gamaj Panel. Single-purpose, self-hosted VPN management panel with its dashboard, CLI, and gateway binary.

## Highlights

- **Binary-only installation.** The install path is exclusively the native binary; no container runtime is ever required or attempted.
- **All VPN backends enabled by default.** OpenVPN, WireGuard, L2TP, PPTP, Remote Access, and Extra VPN work out of the box with no install-mode gate.
- **Panel service on port 616.** One port for the dashboard and the API, configurable through `GAMAJ_*` environment variables only.
- **Deterministic migrations.** The Go migrations runner with legacy schema revision detection.
- **Dashboard.** Full-featured web dashboard embedded in the gateway binary.
- **CLI.** `gamaj` command-line tool for server administration.
- **Environment naming guard.** CI fails on any environment variable without the `GAMAJ_` prefix, keeping configuration Gamaj-owned.

## Install

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/TheGamaj/Panel/Asli/scripts/gamaj/gamaj.sh)
```

The installer is non-interactive and sets up the `gamaj-panel` systemd service listening on port 616.

## Assets

Download the matching `gamaj-*` archive for your platform below. SHA256 checksums are published next to each archive.

## Requirements

- Linux (amd64, arm64, armv7) or Windows (amd64)
- MariaDB or MySQL database
- See [INSTALL.md](https://github.com/TheGamaj/Panel/blob/Asli/docs/INSTALL.md) for the full guide.
