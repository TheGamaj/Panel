# Gamaj Installer Scripts

Installer, lifecycle, and test scripts for Gamaj and Gamaj Node.

Gamaj ships **one install path**: the native binary service. There is no
container install, no compose file, and no install-mode switch to remember.

## Install Gamaj

```bash
curl -fsSL https://raw.githubusercontent.com/TheGamaj/Panel/Asli/scripts/gamaj/gamaj.sh | sudo bash -s -- install
```

Install the dev channel or a specific release:

```bash
curl -fsSL https://raw.githubusercontent.com/TheGamaj/Panel/Asli/scripts/gamaj/gamaj.sh | sudo bash -s -- install --dev
curl -fsSL https://raw.githubusercontent.com/TheGamaj/Panel/Asli/scripts/gamaj/gamaj.sh | sudo bash -s -- install --version is.0.0.1
```

Pick a database (SQLite, MySQL, or MariaDB):

```bash
curl -fsSL https://raw.githubusercontent.com/TheGamaj/Panel/Asli/scripts/gamaj/gamaj.sh | sudo bash -s -- install --database sqlite
curl -fsSL https://raw.githubusercontent.com/TheGamaj/Panel/Asli/scripts/gamaj/gamaj.sh | sudo bash -s -- install --database mysql
curl -fsSL https://raw.githubusercontent.com/TheGamaj/Panel/Asli/scripts/gamaj/gamaj.sh | sudo bash -s -- install --database mariadb
```

The panel serves on port **616**. `install` is non-interactive when stdin is not
a terminal, so it is safe to run from automation.

Do not run the installer as `sudo bash -c "$(curl ...)"`; the script body can
exceed Linux's single-argument limit. Always pipe the download into bash as
shown above.

## Manage Gamaj

```bash
sudo gamaj status
sudo gamaj restart
sudo gamaj logs
sudo gamaj backup
sudo gamaj update --dev
sudo gamaj update --version is.0.0.1
sudo gamaj core-update
sudo gamaj uninstall
```

## Install Gamaj Node

```bash
curl -fsSL https://raw.githubusercontent.com/TheGamaj/Panel/Asli/scripts/gamaj/gamaj-node.sh | sudo bash -s -- install
curl -fsSL https://raw.githubusercontent.com/TheGamaj/Panel/Asli/scripts/gamaj/gamaj-node.sh | sudo bash -s -- install --name gamaj-node2
```

A second node on the same host must use a different `--name` and a different
control port. Every node runs on the same native binary install, so OpenVPN,
WireGuard, L2TP, PPTP, Remote Access, and Extra VPN are available without any
extra mode switch.

## Tests

```bash
bash scripts/gamaj/test-node-installer.sh          # Node installer unit checks
bash scripts/gamaj/test-database-maintenance.sh    # Database lifecycle checks
sudo bash scripts/tests/e2e-binary-install.sh      # Full install on port 616
```

`e2e-binary-install.sh` builds the dashboard and the binaries, installs them into
a sandbox prefix, starts the service, and fails when the panel is not healthy on
port 616. It requires Linux, systemd, root, Go, and Node.js; elsewhere it reports
`SKIP` and exits successfully.
