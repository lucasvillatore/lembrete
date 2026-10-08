#!/usr/bin/env bash
# Publica uma versão nova: cria a tag e envia; o GitHub Actions compila e publica no Releases.
#   ./scripts/release.sh            patch: v0.1.0 -> v0.1.1
#   ./scripts/release.sh minor      v0.1.0 -> v0.2.0
#   ./scripts/release.sh major      v0.1.0 -> v1.0.0
#   ./scripts/release.sh v0.3.0     versão exata
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

erro() { echo "$*" >&2; exit 1; }

[[ $(git branch --show-current) == main ]] || erro "Rode na main."
[[ -z $(git status --porcelain) ]] || erro "Há mudanças não commitadas."
git fetch -q origin main --tags
[[ $(git rev-parse HEAD) == $(git rev-parse origin/main) ]] || erro "A main local está diferente da origin/main. Rode git pull."

atual=$(git tag --list 'v*' --sort=-v:refname | head -1)
atual=${atual:-v0.0.0}
IFS=. read -r maj min pat <<<"${atual#v}"
case "${1:-patch}" in
  patch) nova="v$maj.$min.$((pat + 1))" ;;
  minor) nova="v$maj.$((min + 1)).0" ;;
  major) nova="v$((maj + 1)).0.0" ;;
  v[0-9]*.[0-9]*.[0-9]*) nova=$1 ;;
  *) erro "Uso: $0 [patch|minor|major|vX.Y.Z]" ;;
esac
git rev-parse -q --verify "refs/tags/$nova" >/dev/null && erro "A tag $nova já existe."

echo "Verificando se compila..."
go vet ./... && go build -o /dev/null .

echo
git log --oneline "$atual"..HEAD 2>/dev/null || git log --oneline
echo
read -rp "Publicar $nova (atual: $atual)? [s/N] " ok
[[ $ok == [sS] ]] || { echo "Cancelado."; exit 0; }

git tag -a "$nova" -m "$nova"
git push -q origin "$nova"
echo "Tag $nova enviada."

if command -v gh >/dev/null; then
  echo "Acompanhando o build..."
  sleep 5
  run=$(gh run list --workflow release.yml --branch "$nova" --limit 1 --json databaseId --jq '.[0].databaseId')
  gh run watch "$run" --exit-status >/dev/null && echo "Publicado: $(gh release view "$nova" --json url --jq .url)"
else
  echo "Acompanhe em: https://github.com/lucasvillatore/lembrete/actions"
fi
