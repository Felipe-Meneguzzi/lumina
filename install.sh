#!/usr/bin/env bash
# Lumina installer — downloads the latest binary from GitHub Releases and places
# it in the user's PATH.
#
# Quick install (no need to clone the repo):
#   curl -fsSL https://raw.githubusercontent.com/Felipe-Meneguzzi/lumina/main/install.sh | bash
#
# Recognised environment variables:
#   LUMINA_VERSION   — tag to install (e.g. v0.3.1). Default: latest
#   INSTALL_DIR      — destination directory. Default: ~/.local/bin (fallback /usr/local/bin)
#   LUMINA_REPO      — repo override (owner/name). Default: Felipe-Meneguzzi/lumina

set -euo pipefail

REPO="${LUMINA_REPO:-Felipe-Meneguzzi/lumina}"
VERSION="${LUMINA_VERSION:-latest}"
INSTALL_DIR="${INSTALL_DIR:-}"

err()  { printf '\033[31merror:\033[0m %s\n' "$*" >&2; }
info() { printf '\033[36m==>\033[0m %s\n' "$*"; }
warn() { printf '\033[33mwarn:\033[0m %s\n' "$*" >&2; }

need() {
  command -v "$1" >/dev/null 2>&1 || { err "command '$1' not found — install it before running the installer"; exit 1; }
}

# ── 1. Minimum dependencies ──────────────────────────────────────────────────
need uname
if command -v curl >/dev/null 2>&1; then
  FETCH="curl -fsSL"
  FETCH_OUT="curl -fsSL -o"
elif command -v wget >/dev/null 2>&1; then
  FETCH="wget -qO-"
  FETCH_OUT="wget -qO"
else
  err "neither curl nor wget is installed"
  exit 1
fi

# ── 2. Detect OS and architecture ────────────────────────────────────────────
OS_RAW="$(uname -s)"
case "$OS_RAW" in
  Linux)   OS="linux"  ;;
  Darwin)  OS="darwin" ;;
  *)
    err "unsupported OS: $OS_RAW (Lumina runs on Linux and macOS — on Windows use WSL)"
    exit 1
    ;;
esac

ARCH_RAW="$(uname -m)"
case "$ARCH_RAW" in
  x86_64|amd64)   ARCH="amd64" ;;
  aarch64|arm64)  ARCH="arm64" ;;
  *)
    err "unsupported architecture: $ARCH_RAW"
    exit 1
    ;;
esac

# ── 3. Resolve version ───────────────────────────────────────────────────────
if [[ "$VERSION" == "latest" ]]; then
  info "fetching latest release from $REPO"
  # The API returns JSON; extract tag_name without depending on jq.
  API_URL="https://api.github.com/repos/$REPO/releases/latest"
  TAG="$($FETCH "$API_URL" | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -n1)"
  if [[ -z "$TAG" ]]; then
    err "could not determine the latest release of $REPO"
    err "check that the repo has published releases or set LUMINA_VERSION"
    exit 1
  fi
else
  TAG="$VERSION"
fi
info "version: $TAG"

# ── 3b. Already installed at this version? ───────────────────────────────────
if command -v lumina >/dev/null 2>&1; then
  CURRENT="$(lumina --version 2>/dev/null || true)"
  if [[ -n "$CURRENT" && "$CURRENT" == "$TAG" ]]; then
    info "lumina $TAG is already installed at $(command -v lumina) — nothing to do"
    exit 0
  fi
  if [[ -n "$CURRENT" ]]; then
    info "upgrading from $CURRENT to $TAG"
  fi
fi

# ── 4. Choose installation directory ─────────────────────────────────────────
if [[ -z "$INSTALL_DIR" ]]; then
  if [[ -w "${HOME}/.local/bin" ]] || mkdir -p "${HOME}/.local/bin" 2>/dev/null; then
    INSTALL_DIR="${HOME}/.local/bin"
  elif [[ -w "/usr/local/bin" ]]; then
    INSTALL_DIR="/usr/local/bin"
  else
    INSTALL_DIR="${HOME}/.local/bin"
    mkdir -p "$INSTALL_DIR"
  fi
fi
info "destination: $INSTALL_DIR"

# ── 5. Download binary to a tmp dir and move atomically ──────────────────────
ASSET="lumina-${OS}-${ARCH}"
URL="https://github.com/${REPO}/releases/download/${TAG}/${ASSET}"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

info "downloading $URL"
if ! $FETCH_OUT "$TMP_DIR/lumina" "$URL"; then
  err "failed to download $ASSET from release $TAG"
  err "check that the asset '$ASSET' exists at https://github.com/${REPO}/releases/tag/${TAG}"
  exit 1
fi

chmod +x "$TMP_DIR/lumina"

# ── 6. Optional checksum (if the release publishes checksums.txt) ─────────────
CHECKSUM_URL="https://github.com/${REPO}/releases/download/${TAG}/checksums.txt"
if command -v sha256sum >/dev/null 2>&1; then
  if $FETCH "$CHECKSUM_URL" >"$TMP_DIR/checksums.txt" 2>/dev/null; then
    EXPECTED="$(grep " $ASSET\$" "$TMP_DIR/checksums.txt" | awk '{print $1}' || true)"
    if [[ -n "$EXPECTED" ]]; then
      ACTUAL="$(sha256sum "$TMP_DIR/lumina" | awk '{print $1}')"
      if [[ "$EXPECTED" != "$ACTUAL" ]]; then
        err "invalid checksum: expected $EXPECTED, got $ACTUAL"
        exit 1
      fi
      info "checksum ok"
    fi
  fi
fi

# ── 7. Install ───────────────────────────────────────────────────────────────
DEST="${INSTALL_DIR}/lumina"
if [[ -w "$INSTALL_DIR" ]]; then
  mv -f "$TMP_DIR/lumina" "$DEST"
else
  warn "$INSTALL_DIR is not writable; trying with sudo"
  sudo mv -f "$TMP_DIR/lumina" "$DEST"
fi

info "installed at $DEST"

# ── 8. Check PATH ────────────────────────────────────────────────────────────
case ":$PATH:" in
  *":$INSTALL_DIR:"*)
    info "all done — run: lumina"
    ;;
  *)
    warn "$INSTALL_DIR is not in PATH"
    cat <<EOF

Add this to your shell rc (~/.bashrc, ~/.zshrc, ~/.profile):

    export PATH="$INSTALL_DIR:\$PATH"

Then reload: source ~/.bashrc (or open a new terminal)
EOF
    ;;
esac
