#!/bin/sh
# Instala o lembrete (Linux e macOS):
#   curl -fsSL https://raw.githubusercontent.com/lucasvillatore/lembrete/main/install.sh | sh
set -eu

REPO=lucasvillatore/lembrete
DEST="${LEMBRETE_DIR:-$HOME/.local/bin}"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$os" in
  linux | darwin) ;;
  *) echo "Sistema não suportado: $os" >&2; exit 1 ;;
esac

arch=$(uname -m)
case "$arch" in
  x86_64 | amd64) arch=amd64 ;;
  aarch64 | arm64) arch=arm64 ;;
  *) echo "Arquitetura não suportada: $arch" >&2; exit 1 ;;
esac

url="https://github.com/$REPO/releases/latest/download/lembrete_${os}_${arch}.tar.gz"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

echo "Baixando $url"
curl -fsSL "$url" | tar -xz -C "$tmp" lembrete
mkdir -p "$DEST"
install -m 755 "$tmp/lembrete" "$DEST/lembrete"
echo "Instalado em $DEST/lembrete"

case ":$PATH:" in
  *":$DEST:"*) ;;
  *) echo "Atenção: $DEST não está no PATH. Adicione no seu shell: export PATH=\"$DEST:\$PATH\"" ;;
esac

# Mostra os lembretes vencidos ao abrir o terminal (só adiciona uma vez).
linha='command -v lembrete >/dev/null && lembrete check'
case "${SHELL:-}" in
  */zsh) rc="$HOME/.zshrc" ;;
  */bash) rc="$HOME/.bashrc" ;;
  *) rc="" ;;
esac
if [ -z "$rc" ]; then
  echo "Para ver os lembretes ao abrir o terminal, adicione no seu shell: $linha"
elif ! grep -qF "$linha" "$rc" 2>/dev/null; then
  printf '\n# lembrete: mostra os lembretes vencidos ao abrir o terminal\n%s\n' "$linha" >>"$rc"
  echo "Adicionado em $rc"
fi

echo "Pronto. Abra um terminal novo e rode: lembrete"
