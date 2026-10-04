# Gamaj Panel — install and upgrade guide

Version covered: **`is.0.0.1`** · Default port: **616**

Gamaj Panel installs as a native binary with a systemd service. There is no
container install: `install` never pulls an image, never needs docker, and never
asks which install mode to use.

- Panel repository: `TheGamaj/Panel`
- Node runtime: install on each node server (section 6) — node project: [TheGamaj/Node](https://github.com/TheGamaj/Node)
- Telegram bot: install from the panel's **Applications** page (recommended) — bot project: [TheGamaj/Bot](https://github.com/TheGamaj/Bot)

<p align="center">
  <a href="https://t.me/TheGamaj">Channel</a> ·
  <a href="https://t.me/GamajGP">Support group</a> ·
  <a href="https://t.me/AsliCode">Code channel</a>
</p>

## 1. Requirements

| Item | Value |
|---|---|
| Operating system | Linux with systemd: Debian/Ubuntu, CentOS/Rocky/Alma, Fedora, Arch, Alpine |
| Access | root (or a user with sudo) |
| Port | **616** for the dashboard and API |
| Database | SQLite (default), MySQL, or MariaDB |
| Architectures | `linux-386`, `linux-amd64`, `linux-arm64`, `linux-armv5`, `linux-armv6`, `linux-armv7`, `linux-s390x` |
| Source builds | Go 1.25+, Node.js 20+ and npm for the dashboard |

## 2. Install with the official script (recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/TheGamaj/Panel/Asli/scripts/gamaj/gamaj.sh | sudo bash -s -- install
```

The installer:

1. detects the OS and architecture,
2. downloads the matching `gamaj-server` and `gamaj-cli` release assets,
3. installs them under `/opt/gamaj/bin`,
4. writes `/opt/gamaj/.env` from `.env.example` when it does not exist yet,
5. creates the `gamaj.service` systemd unit and starts it.

Options:

```bash
# a specific release
... | sudo bash -s -- install --version is.0.0.1

# database choice
... | sudo bash -s -- install --database sqlite
... | sudo bash -s -- install --database mysql
... | sudo bash -s -- install --database mariadb
```

Notes:

- Never run the installer as `sudo bash -c "$(curl ...)"`. The script body can
  exceed Linux's single-argument limit; always pipe the download into bash.
- When stdin is not a terminal (automation, CI, `curl | bash`) the installer
  makes every choice automatically and installs in binary mode.
- If the release does not publish an asset for your architecture, the installer
  stops with the expected asset name, the assets that do exist, and the list of
  supported architectures. It never leaves a half-installed panel behind.

## 3. Manual install from source

Use this when you want to run your own build.

```bash
git clone https://github.com/TheGamaj/Panel.git Gamaj-Panel
cd Gamaj-Panel

# 1. Dashboard bundle (once)
cd dashboard
npm ci
VITE_BASE_API=/api/ npm run build
cp build/index.html build/404.html
cd ..

# 2. Panel binaries (the dashboard bundle is embedded into the gateway)
bash scripts/build_binary.sh        # produces dist/gamaj-server and dist/gamaj-cli

# 3. Install the files
sudo mkdir -p /opt/gamaj/bin /var/lib/gamaj
sudo install -m 0755 dist/gamaj-server /opt/gamaj/bin/gamaj-server
sudo install -m 0755 dist/gamaj-cli /opt/gamaj/bin/gamaj-cli
sudo cp .env.example /opt/gamaj/.env
sudo editor /opt/gamaj/.env         # set GAMAJ_DATABASE_URL at least
```

Systemd unit (`/etc/systemd/system/gamaj.service`):

```ini
[Unit]
Description=Gamaj Panel
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/gamaj
Environment=GAMAJ_APP_DIR=/opt/gamaj
Environment=GAMAJ_ENV_FILE=/opt/gamaj/.env
Environment=GAMAJ_INSTALL_MODE=binary
Environment=GAMAJ_DATA_DIR=/var/lib/gamaj
ExecStart=/opt/gamaj/bin/gamaj-server
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now gamaj
sudo systemctl status gamaj --no-pager
```

## 4. Offline / air-gapped install

Build the binaries on a machine with internet access (section 3, step 2), then
copy `dist/gamaj-server`, `dist/gamaj-cli` and the installer script to the
target server and point the installer at the local files:

```bash
sudo GAMAJ_BINARY_SERVER_OVERRIDE="$PWD/gamaj-server" \
     GAMAJ_BINARY_CLI_OVERRIDE="$PWD/gamaj-cli" \
     GAMAJ_BINARY_OVERRIDE_VERSION=is.0.0.1 \
     bash gamaj.sh install --database sqlite
```

The same variables work for custom builds: the installer skips every download
and installs exactly the files you provide.

## 5. First run

```bash
# First administrator (full access)
sudo gamaj cli admin create --role full_access

# Service status and logs
sudo gamaj status
sudo gamaj logs
```

Open the dashboard:

- with a domain and TLS: `https://YOUR_DOMAIN:616/dashboard/`
- locally through an SSH tunnel: `ssh -L 616:localhost:616 user@serverip`, then
  `http://localhost:616/dashboard/`

### TLS certificates

```bash
# Domain certificate (Let's Encrypt)
sudo gamaj ssl issue --email you@example.com --domains panel.example.com

# Short-lived public IP certificate
sudo gamaj ssl issue --email you@example.com --ip-address 203.0.113.10

# Self-signed certificate for an IP (browser warning)
sudo gamaj ssl issue --email you@example.com --domains 203.0.113.10 --provider self-signed

# Renew
sudo gamaj ssl renew
```

The interactive installer can also run this step for you (`prompt_ssl_setup`),
and you can point Gamaj at existing files instead:

```dotenv
GAMAJ_SSL_CERTFILE=/etc/letsencrypt/live/panel.example.com/fullchain.pem
GAMAJ_SSL_KEYFILE=/etc/letsencrypt/live/panel.example.com/privkey.pem
```

### Firewall

Only the gateway port has to be reachable from outside:

```bash
# ufw
sudo ufw allow 616/tcp

# firewalld
sudo firewall-cmd --permanent --add-port=616/tcp && sudo firewall-cmd --reload
```

## 6. Add a node

On the node server:

```bash
curl -fsSL https://raw.githubusercontent.com/TheGamaj/Panel/Asli/scripts/gamaj/gamaj-node.sh | sudo bash -s -- install
```

Then, in the panel, open **Nodes** and add the server address and control port
(default `62050`). The panel issues the node certificate and runtime config; the
node reports usage, online sessions, and executes host actions on request.

## 7. Upgrade

```bash
sudo gamaj update               # latest published release
sudo gamaj update --version is.0.0.1
```

Upgrades replace the binaries in place, keep `/opt/gamaj/.env` and
`/var/lib/gamaj`, run database migrations on start, and restart the service.

## 8. Uninstall

```bash
sudo gamaj uninstall            # asks before removing data as well
```

## 9. Troubleshooting

| Symptom | Check |
|---|---|
| Panel does not answer on 616 | `systemctl status gamaj`, `journalctl -u gamaj -n 100`, and `ss -ltn | grep 616` |
| Installer says no matching asset | The release must publish `gamaj-linux-<arch>.tar.gz` (or the split server/CLI assets) for your architecture |
| Dashboard shows a login loop | Check `GAMAJ_DATABASE_URL` in `/opt/gamaj/.env` and that the service can write to `GAMAJ_DATA_DIR` |
| TLS certificate errors | `sudo gamaj ssl renew`, and confirm `GAMAJ_SSL_CERTFILE` / `GAMAJ_SSL_KEYFILE` paths |
| Node stays offline in the panel | Verify the control port is open between the two servers and that the node certificate matches the record in the panel |

## 10. Verify the install path end to end

```bash
sudo bash scripts/tests/e2e-binary-install.sh
```

The script builds the dashboard and binaries, installs them into a sandbox
prefix, starts the service, and fails unless the panel is healthy on port 616,
serves the dashboard, and accepts an admin login through the API. It needs
Linux, systemd, root, Go, and Node.js; on other platforms it prints `SKIP` and
exits successfully.
