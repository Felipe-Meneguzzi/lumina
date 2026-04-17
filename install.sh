#!/usr/bin/env bash
# Lumina installer — baixa o binário mais recente do GitHub Releases e coloca
# no PATH do usuário.
#
# Uso rápido (sem clonar o repo):
#   curl -fsSL https://raw.githubusercontent.com/menegas/lumina/main/install.sh | bash
#
# Variáveis de ambiente reconhecidas:
#   LUMINA_VERSION   — tag a instalar (ex: v0.3.1). Default: latest
#   INSTALL_DIR      — diretório destino. Default: ~/.local/bin (fallback /usr/local/bin)
#   LUMINA_REPO      — override do repo (owner/name). Default: menegas/lumina

set -euo pipefail

REPO="${LUMINA_REPO:-menegas/lumina}"
VERSION="${LUMINA_VERSION:-latest}"
INSTALL_DIR="${INSTALL_DIR:-}"

err()  { printf '\033[31merror:\033[0m %s\n' "$*" >&2; }
info() { printf '\033[36m==>\033[0m %s\n' "$*"; }
warn() { printf '\033[33mwarn:\033[0m %s\n' "$*" >&2; }

need() {
  command -v "$1" >/dev/null 2>&1 || { err "comando '$1' não encontrado — instale antes de rodar o installer"; exit 1; }
}

# ── 1. Dependências mínimas ──────────────────────────────────────────────────
need uname
if command -v curl >/dev/null 2>&1; then
  FETCH="curl -fsSL"
  FETCH_OUT="curl -fsSL -o"
elif command -v wget >/dev/null 2>&1; then
  FETCH="wget -qO-"
  FETCH_OUT="wget -qO"
else
  err "nem curl nem wget estão instalados"
  exit 1
fi

# ── 2. Detectar OS e arquitetura ─────────────────────────────────────────────
OS_RAW="$(uname -s)"
case "$OS_RAW" in
  Linux)   OS="linux"  ;;
  Darwin)  OS="darwin" ;;
  *)
    err "sistema não suportado: $OS_RAW (Lumina roda em Linux e macOS — no Windows use WSL)"
    exit 1
    ;;
esac

ARCH_RAW="$(uname -m)"
case "$ARCH_RAW" in
  x86_64|amd64)   ARCH="amd64" ;;
  aarch64|arm64)  ARCH="arm64" ;;
  *)
    err "arquitetura não suportada: $ARCH_RAW"
    exit 1
    ;;
esac

# ── 3. Resolver versão ───────────────────────────────────────────────────────
if [[ "$VERSION" == "latest" ]]; then
  info "consultando última release de $REPO"
  # A API retorna um JSON; extraímos tag_name sem depender de jq.
  API_URL="https://api.github.com/repos/$REPO/releases/latest"
  TAG="$($FETCH "$API_URL" | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -n1)"
  if [[ -z "$TAG" ]]; then
    err "não foi possível descobrir a última release de $REPO"
    err "verifique se o repo tem releases publicados ou defina LUMINA_VERSION"
    exit 1
  fi
else
  TAG="$VERSION"
fi
info "versão: $TAG"

# ── 3b. Já está instalado nessa versão? ──────────────────────────────────────
if command -v lumina >/dev/null 2>&1; then
  CURRENT="$(lumina --version 2>/dev/null || true)"
  if [[ -n "$CURRENT" && "$CURRENT" == "$TAG" ]]; then
    info "lumina $TAG já está instalado em $(command -v lumina) — nada a fazer"
    exit 0
  fi
  if [[ -n "$CURRENT" ]]; then
    info "atualizando de $CURRENT para $TAG"
  fi
fi

# ── 4. Decidir diretório de instalação ───────────────────────────────────────
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
info "destino: $INSTALL_DIR"

# ── 5. Baixar binário para tmp e mover atomicamente ──────────────────────────
ASSET="lumina-${OS}-${ARCH}"
URL="https://github.com/${REPO}/releases/download/${TAG}/${ASSET}"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

info "baixando $URL"
if ! $FETCH_OUT "$TMP_DIR/lumina" "$URL"; then
  err "falha ao baixar $ASSET da release $TAG"
  err "confira se o asset '$ASSET' existe em https://github.com/${REPO}/releases/tag/${TAG}"
  exit 1
fi

chmod +x "$TMP_DIR/lumina"

# ── 6. Checksum opcional (se a release publicar checksums.txt) ───────────────
CHECKSUM_URL="https://github.com/${REPO}/releases/download/${TAG}/checksums.txt"
if command -v sha256sum >/dev/null 2>&1; then
  if $FETCH "$CHECKSUM_URL" >"$TMP_DIR/checksums.txt" 2>/dev/null; then
    EXPECTED="$(grep " $ASSET\$" "$TMP_DIR/checksums.txt" | awk '{print $1}' || true)"
    if [[ -n "$EXPECTED" ]]; then
      ACTUAL="$(sha256sum "$TMP_DIR/lumina" | awk '{print $1}')"
      if [[ "$EXPECTED" != "$ACTUAL" ]]; then
        err "checksum inválido: esperado $EXPECTED, obtido $ACTUAL"
        exit 1
      fi
      info "checksum ok"
    fi
  fi
fi

# ── 7. Instalar ──────────────────────────────────────────────────────────────
DEST="${INSTALL_DIR}/lumina"
if [[ -w "$INSTALL_DIR" ]]; then
  mv -f "$TMP_DIR/lumina" "$DEST"
else
  warn "$INSTALL_DIR não é gravável; tentando com sudo"
  sudo mv -f "$TMP_DIR/lumina" "$DEST"
fi

info "instalado em $DEST"

# ── 8. Checar PATH ───────────────────────────────────────────────────────────
case ":$PATH:" in
  *":$INSTALL_DIR:"*)
    info "pronto — rode: lumina"
    ;;
  *)
    warn "$INSTALL_DIR não está no PATH"
    cat <<EOF

Adicione isto ao seu shell rc (~/.bashrc, ~/.zshrc, ~/.profile):

    export PATH="$INSTALL_DIR:\$PATH"

Depois recarregue: source ~/.bashrc (ou reabra o terminal)
EOF
    ;;
esac
