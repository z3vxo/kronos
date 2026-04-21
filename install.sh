#!/usr/bin/env bash
set -e

GO_MIN="1.24.0"
GO_INSTALL="1.24.2"

check_go() {
    if command -v go &>/dev/null; then
        current=$(go version | grep -oP '\d+\.\d+(\.\d+)?' | head -1)
        # compare major.minor.patch
        if printf '%s\n%s\n' "$GO_MIN" "$current" | sort -V -C; then
            echo "[*] Go $current already installed"
            return 0
        fi
    fi
    return 1
}

install_go() {
    echo "[*] Installing Go $GO_INSTALL..."
    ARCH=$(uname -m)
    case $ARCH in
        x86_64)  ARCH="amd64" ;;
        aarch64) ARCH="arm64" ;;
        *)        echo "[!] Unsupported arch: $ARCH"; exit 1 ;;
    esac

    URL="https://go.dev/dl/go${GO_INSTALL}.linux-${ARCH}.tar.gz"
    TMP=$(mktemp -d)
    curl -fsSL "$URL" -o "$TMP/go.tar.gz"
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf "$TMP/go.tar.gz"
    rm -rf "$TMP"

    export PATH=$PATH:/usr/local/go/bin
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    echo "[+] Go $(go version) installed"
}

install_deps() {
    echo "[*] Installing system dependencies..."
    if command -v apt-get &>/dev/null; then
        sudo apt-get install -y gcc libsqlite3-dev
    elif command -v dnf &>/dev/null; then
        sudo dnf install -y gcc sqlite-devel
    elif command -v pacman &>/dev/null; then
        sudo pacman -S --noconfirm gcc sqlite
    else
        echo "[!] Could not detect package manager — install gcc and sqlite3 dev headers manually"
    fi
}

build() {
    local dir=$1
    local out=$2
    echo "[*] Building $out..."
    cd "$dir"
    go mod download
    go build -o "../bin/$out" .
    cd - &>/dev/null
    echo "[+] Built bin/$out"
}

ROOT=$(cd "$(dirname "$0")" && pwd)
mkdir -p "$ROOT/bin"

check_go || install_go
install_deps

build "$ROOT/server" "kronos-server"
build "$ROOT/client" "kronos-client"

echo ""
echo "[+] Done. Binaries are in $ROOT/bin/"
