#!/usr/bin/env bash
# Gamaj — host-side runner for the Linux + systemd end-to-end install check.
#
# e2e-binary-install.sh needs Linux, systemd and root. On a Windows or macOS
# workstation this wrapper finds an already-available Linux runtime and runs the
# real check inside it, so the developer gets the same signal locally without
# setting up Linux by hand.
#
# Order of preference:
#   1. already on Linux  -> run the check directly
#   2. WSL2 distro       -> run inside the distro
#   3. Docker / Podman   -> run privileged with systemd as PID 1
#   4. Lima / Colima     -> run inside the Linux VM
#   5. Multipass         -> run inside a transient Ubuntu VM
#
# Nothing is installed automatically. When no runtime is present the script
# reports SKIP and explains exactly what to install.
#
# Usage:
#   bash scripts/tests/e2e-vm-runner.sh
#   GAMAJ_E2E_WSL_DISTRO=Ubuntu bash scripts/tests/e2e-vm-runner.sh
#
# Exit codes: 0 = check passed (or SKIP), non-zero = check failed.

set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$script_dir/../.." && pwd)"
e2e_script="$script_dir/e2e-binary-install.sh"

inner_cmd="cd '$(printf '%s' "$repo_root" | sed "s/'/'\\\\''/g")' && sudo bash scripts/tests/e2e-binary-install.sh"

skip() {
    echo "SKIP: $*"
    echo
    echo "The end-to-end install check needs a Linux kernel with systemd and root."
    echo "Install exactly one of the following, then re-run this script:"
    echo "  - WSL2   : wsl --install -d Ubuntu        (Windows, already present on many machines)"
    echo "  - Docker : https://docs.docker.com/desktop/install/windows-install/"
    echo "  - Podman : https://podman.io/docs/installation"
    echo "  - Lima   : brew install lima              (macOS)"
    echo "  - Multipass: snap install multipass       (Linux hosts)"
    exit 0
}

have() { command -v "$1" >/dev/null 2>&1; }

# ---------------------------------------------------------------- 1. native
if [ "$(uname -s)" = "Linux" ]; then
    if ! have systemctl; then
        skip "systemctl is not available"
    fi
    if [ "$(id -u)" -ne 0 ]; then
        skip "the check must run as root"
    fi
    exec bash "$e2e_script"
fi

# -------------------------------------------------------------------- 2. WSL
if [ "$(uname -s)" = "MINGW"* ] || [ "$(uname -s)" = "MSYS"* ] || [ "$(uname -s)" = "CYGWIN"* ]; then
    if have wsl.exe; then
        distro="${GAMAJ_E2E_WSL_DISTRO:-}"
        if [ -z "$distro" ]; then
            distro="$(wsl.exe -l -q 2>/dev/null | tr -d '\r' | grep -v '^$' | head -1 || true)"
        fi
        if [ -n "$distro" ]; then
            echo "WSL2: running the end-to-end install check inside '$distro'"
            # wsl.exe resolves Windows paths itself, so hand it the Windows form.
            win_repo="$(cygpath -w "$repo_root" 2>/dev/null || printf '%s' "$repo_root")"
            exec wsl.exe -d "$distro" -u root -- bash -lc \
                "cd '$(printf '%s' "$win_repo" | sed "s/'/'\\\\''/g")' && bash scripts/tests/e2e-binary-install.sh"
        fi
        echo "WSL2 is present but has no distro; skipping that route."
    fi
fi

# ------------------------------------------------------- 3. Docker / Podman
for runtime in docker podman; do
    if have "$runtime"; then
        if ! "$runtime" info >/dev/null 2>&1; then
            echo "$runtime is installed but not usable; skipping that route."
            continue
        fi
        echo "$runtime: running the end-to-end install check in a systemd container"
        exec "$runtime" run --rm -it \
            --privileged \
            --cgroupns=host \
            --volume /sys/fs/cgroup:/sys/fs/cgroup:rw \
            --volume "$repo_root:$repo_root" \
            --workdir "$repo_root" \
            docker.io/library/ubuntu:24.04 \
            bash -c "apt-get update -qq && apt-get install -y -qq systemd jq curl >/dev/null && $inner_cmd"
    fi
done

# ----------------------------------------------------------- 4. Lima/Colima
if have limactl; then
    echo "lima: running the end-to-end install check inside a Linux VM"
    exec limactl shell "$(limactl list --format '{{.Name}}' 2>/dev/null | head -1)" \
        bash -lc "$inner_cmd"
fi
if have colima; then
    echo "colima: starting the VM and running the end-to-end install check"
    colima start >/dev/null 2>&1 || true
    colima ssh -- bash -lc "sudo $inner_cmd"
    exit $?
fi

# ------------------------------------------------------------- 5. Multipass
if have multipass; then
    instance="${GAMAJ_E2E_VM:-gamaj-e2e}"
    if ! multipass list | awk '{print $1}' | grep -qx "$instance"; then
        echo "multipass: launching transient VM '$instance'"
        multipass launch --name "$instance" --cpus 2 --memory 2G 24.04 >/dev/null
        multipass exec "$instance" -- sudo apt-get update -qq
        multipass exec "$instance" -- sudo apt-get install -y -qq systemd jq curl >/dev/null
    fi
    echo "multipass: running the end-to-end install check inside '$instance'"
    multipass exec "$instance" -- bash -lc "$inner_cmd"
    exit $?
fi

skip "no Linux runtime (WSL, Docker, Podman, Lima, Colima, Multipass) was found"