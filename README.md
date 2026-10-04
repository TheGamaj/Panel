<p align="center">
 <a href="./README.md">English</a> /
 <a href="./docs/README-fa.md">فارسی</a> /
 <a href="./docs/README-ru.md">Русский</a> /
 <a href="./docs/README-zh-cn.md">简体中文</a>
</p>

<p align="center">
  <img src="./docs/assets/gamaj-logo.svg" alt="Gamaj" width="104" height="104">
</p>

<h1>GAMAJ</h1>

### Panel

<p align="center">
  <a href="https://github.com/TheGamaj/Panel/releases"><img alt="Release" src="https://img.shields.io/badge/release-is.0.0.1-2ea043?style=flat-square" /></a>
  <img alt="Go" src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go&logoColor=white" />
  <img alt="Port" src="https://img.shields.io/badge/default%20port-616-blue?style=flat-square" />
  <img alt="License" src="https://img.shields.io/badge/license-AGPL--3.0-lightgrey?style=flat-square" />
  <a href="https://github.com/TheGamaj/Panel/actions/workflows/binary-build.yml"><img alt="Build" src="https://img.shields.io/github/actions/workflow/status/TheGamaj/Panel/binary-build.yml?style=flat-square&label=build" /></a>
  <a href="https://github.com/TheGamaj/Panel/actions/workflows/e2e-install.yml"><img alt="E2E" src="https://img.shields.io/github/actions/workflow/status/TheGamaj/Panel/e2e-install.yml?style=flat-square&label=e2e%20install" /></a>
</p>

**گمج (Gamaj)** is a self-hosted management platform for Xray-based proxy
accounts. It gives you a web dashboard, a REST API, a CLI, and multi-node
distribution — all on one native binary plus a systemd service.

Current version: **`is.0.0.1`** · Default port: **616**

<p align="center">
  <a href="https://t.me/TheGamaj">Telegram channel</a> ·
  <a href="https://t.me/GamajGP">Support group</a> ·
  <a href="https://t.me/AsliCode">Code channel</a>
</p>

## Features

- **Web dashboard** (React + Chakra UI) with light/dark theme and four languages (English, فارسی, Русский, 简体中文)
- **REST API** for every dashboard action, plus **API keys** for bots and integrations
- **Multiple nodes** with mutual-TLS gRPC control and automatic certificate enrollment
- Protocols: **VLESS**, **VMess**, **Trojan**, **Shadowsocks**, plus VPN inbound types on nodes (OpenVPN, WireGuard, L2TP, PPTP, SSTP, Remote Access)
- **Multi-user per inbound**, multiple inbounds per user, fallbacks on a single port
- **Traffic, expiry, and concurrent-connection limits**, optional periodic (daily/weekly/…) traffic reset
- **Subscription links** for v2ray-style clients, Clash, and sing-box, with automatic share links and QR codes
- **Built-in Telegram integration**: reports, backups, and account notifications
- **Backup and restore** (database and full archive), export/import, and scheduled Telegram delivery
- **System monitoring**: CPU, memory, disk, online users, per-node usage, Xray logs
- **Multi-admin** with roles (`full_access`, `sudo`, `standard`, `seller`), 2FA, per-admin limits, and API keys
- **External app hosting**: install and manage third-party panel-style apps on the same server
- **HAProxy and host management**: host templates, addresses, certificates, and per-node host actions
- **Fully rebranded and self-contained**: one owner name, one port, no container layer

## Requirements

- Linux server with systemd (Debian/Ubuntu, CentOS/Rocky/Alma, Fedora, Arch, Alpine), root access
- Architectures: `linux-386`, `linux-amd64`, `linux-arm64`, `linux-armv5`, `linux-armv6`, `linux-armv7`, `linux-s390x`
- A database: SQLite (default), MySQL, or MariaDB

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/TheGamaj/Panel/Asli/scripts/gamaj/gamaj.sh | sudo bash -s -- install
```

Install a specific release:

```bash
curl -fsSL https://raw.githubusercontent.com/TheGamaj/Panel/Asli/scripts/gamaj/gamaj.sh | sudo bash -s -- install --version is.0.0.1
```

Choose the database:

```bash
... | sudo bash -s -- install --database sqlite
... | sudo bash -s -- install --database mysql
... | sudo bash -s -- install --database mariadb
```

Do not run the installer as `sudo bash -c "$(curl ...)"`; the script body can
exceed Linux's single-argument limit. Always pipe the download into bash.

Gamaj installs as a native binary with a systemd service. There is no container
install and no install-mode switch: one install path, all features available.

## After install

| Item | Path |
|---|---|
| Panel files | `/opt/gamaj` |
| Configuration | `/opt/gamaj/.env` |
| Data and database | `/var/lib/gamaj` |
| Service | `gamaj.service` |

Create the first administrator:

```bash
sudo gamaj cli admin create --role full_access
```

Open the dashboard:

- Domain with TLS: `https://YOUR_DOMAIN:616/dashboard/`
- SSH tunnel for local testing: `ssh -L 616:localhost:616 user@serverip`, then `http://localhost:616/dashboard/`

Full step-by-step instructions, including TLS and firewall notes, are in
[docs/INSTALL.md](docs/INSTALL.md).

## Manage

```bash
sudo gamaj status          # service status
sudo gamaj restart         # restart panel
sudo gamaj logs            # follow logs
sudo gamaj backup          # create a backup
sudo gamaj update          # update to the latest release
sudo gamaj core-update     # update Xray-core
sudo gamaj edit-env        # edit /opt/gamaj/.env
sudo gamaj ssl             # manage certificates
sudo gamaj uninstall       # remove Gamaj
```

## Configuration

Settings live in `/opt/gamaj/.env`. Every key is `GAMAJ_`-prefixed; the complete
list with comments is in [`.env.example`](.env.example). The most used ones:

| Key | Meaning | Default |
|---|---|---|
| `GAMAJ_DATABASE_URL` | Database connection (required) | — |
| `GAMAJ_HOST` | Gateway bind address | `0.0.0.0` |
| `GAMAJ_PORT` | Gateway port | `616` |
| `GAMAJ_DATA_DIR` | Data directory | `/var/lib/gamaj` |
| `GAMAJ_XRAY_JSON` | Xray configuration file | `/var/lib/gamaj/xray_config.json` |
| `GAMAJ_USER_AUTODELETE_DAYS` | Delete expired users after N days | `0` (disabled) |
| `GAMAJ_JWT_ACCESS_TOKEN_EXPIRE_MINUTES` | Dashboard session lifetime | `1440` |
| `GAMAJ_WEBHOOK_ADDRESS` | Notification webhook endpoints | — |

## Add a node

On another server:

```bash
curl -fsSL https://raw.githubusercontent.com/TheGamaj/Panel/Asli/scripts/gamaj/gamaj-node.sh | sudo bash -s -- install
```

Then in the dashboard, open **Nodes → add node**. The panel issues a certificate
and configuration; the node reports usage, online users, and host actions back
automatically. The node project lives at [TheGamaj/Node](https://github.com/TheGamaj/Node).

## Install the bot

The Telegram bot ships as a separate repository, [TheGamaj/Bot](https://github.com/TheGamaj/Bot),
and speaks only to the panel API. All payment settings live in the panel —
the bot only uses them through the API.

Recommended: open **Applications** in the panel and install **Gamaj Bot**; the
panel downloads the binary, creates a dedicated `gm_...` API key, writes the
bot's config, and registers the webhook.

Manual install:

```bash
git clone https://github.com/TheGamaj/Bot.git Gamaj-Bot
cd Gamaj-Bot
sudo ./scripts/install.sh    # installs Go and builds the binary automatically
sudo nano /opt/gamaj-bot/.gamaj-bot.json   # panel_url, api_key, bot_token, admin_id
sudo systemctl restart gamaj-bot
```

## Command line interface

`gamaj cli` manages users, admins, inbounds, and the database from the shell —
for example `gamaj cli user list`, `gamaj cli admin set-password`,
`gamaj cli migrate up`. See [docs/cli/README.md](docs/cli/README.md).

## Development

```bash
# Build the panel binary with the embedded dashboard
cd dashboard && npm ci && VITE_BASE_API=/api/ npm run build && cd ..
bash scripts/build_binary.sh          # dist/gamaj-server and dist/gamaj-cli

# Go checks
go build ./... && go vet ./internal/... ./cmd/...
go test ./internal/app/system/ ./internal/app/api/ ./internal/gateway/ \
        ./internal/app/backup/ ./internal/app/migrations/ ./cmd/... \
        ./internal/platform/envguard/

# End-to-end install check (Linux, systemd, root)
sudo bash scripts/tests/e2e-binary-install.sh
```

`go test ./internal/platform/envguard/` fails the build when an environment
variable without the `GAMAJ_` prefix (or a legacy name from before Gamaj) is
added anywhere in the tree.

## Documentation

- [docs/INSTALL.md](docs/INSTALL.md) — install and upgrade guide (also in Persian)
- [docs/cli/README.md](docs/cli/README.md) — CLI reference
- [docs/dashboard/README.md](docs/dashboard/README.md) — dashboard notes
- [docs/scripts/gamaj/README.md](docs/scripts/gamaj/README.md) — installer scripts
- [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md) — contribution rules
- [tutorials/](tutorials/) — step-by-step guides shipped inside the panel

## License

GNU Affero General Public License v3.0. See [LICENSE](LICENSE).
