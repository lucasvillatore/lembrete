# lembrete

CLI em Go (Bubble Tea + Lip Gloss) para lembretes que aparecem ao abrir o terminal. Uso e instalação para humanos estão no `README.md`; aqui ficam as regras para evoluir o projeto.

## Estrutura

| Arquivo | Conteúdo |
|---|---|
| `main.go` | Roteamento dos comandos (`add`, `ls`, `ok`, `rm`, `excluir`, `check`), modo inline (`--note/--date/--time`) e formatação da listagem/aviso |
| `tui.go` | Tela de criação: calendário, seletor de hora, editor da nota. Também define cores e estilos compartilhados |
| `lista.go` | Checklist do `ok` / `rm` sem id |
| `store.go` | Leitura/gravação de `~/.lembretes.json` (gravação atômica: tmp + rename) |
| `install.sh` | Instalador `curl \| sh` (Linux e macOS) |
| `scripts/release.sh` | Publica uma versão (cria e envia a tag) |
| `.goreleaser.yaml`, `.github/workflows/release.yml` | Build e publicação no GitHub Releases |

Nomes de funções, variáveis, mensagens e textos em **português**, como no resto do código. Estilos de cor ficam em `var` no topo de `tui.go`/`main.go`/`lista.go`; reutilize os existentes (`corDestaque`, `estiloAtivo`, `estiloApagado`...) em vez de criar cores novas.

## Compatibilidade de dados

`~/.lembretes.json` é o dado real do usuário. Ao mudar a struct `Lembrete`:
- só **adicione** campos, com valor zero que faça sentido para registros antigos;
- nunca renomeie nem remova `id`, `quando` (epoch), `status` (`pendente` | `feito` | `ignorado`) e `nota`.

## Testar

```bash
go vet ./... && go build -o lembrete .
```

- **Nunca** rode o binário contra o `~/.lembretes.json` real para testar. Use um HOME temporário:
  `T=$(mktemp -d); HOME=$T ./lembrete add --note x --date 01/10; HOME=$T ./lembrete check; rm -rf $T`
- As telas (`lembrete` sem args, `ok`/`rm` sem id) precisam de terminal interativo e não rodam pelo Bash tool. Teste a lógica chamando `Update` do model com `tea.KeyMsg` num `_test.go` temporário, e a aparência imprimindo `View()`.
- Comandos com id (`ok 3`, `rm 3`, `excluir 3`) não pedem interação; sem id, travam esperando o usuário.

## Fluxo de mudança

- A `main` tem um ruleset: mudanças entram por **pull request**. Só o dono (`lucasvillatore`, admin) tem bypass para push direto e force-push. Use PR, a menos que ele peça explicitamente para commitar direto.
- Commits no padrão `tipo: descrição` (`feat`, `fix`, `docs`, `chore`, `ci`).
- O binário `lembrete` está no `.gitignore`; não commite.
- O repositório é **público**: não coloque exemplos do trabalho do usuário (ids, nomes de sistemas internos, tickets) no README, nos testes ou nos commits. Use exemplos genéricos.

## Release

Merge na `main` **não publica nada**. O release só sai quando uma tag `v*` é enviada:

```bash
git checkout main && git pull
./scripts/release.sh            # patch | minor | major | vX.Y.Z
```

O script confere se a `main` está limpa e igual à remota e se compila, mostra os commits desde a última tag, pede confirmação (`s/N`), cria e envia a tag e acompanha o build. O GitHub Actions roda o GoReleaser para linux/darwin × amd64/arm64 e publica `lembrete_<os>_<arch>.tar.gz` + `checksums.txt`.

- Windows não é suportado de propósito.
- O `install.sh` baixa sempre `releases/latest`, então os nomes dos arquivos no `.goreleaser.yaml` (`name_template`) e no `install.sh` precisam continuar batendo.
- O `install.sh` é lido direto da `main`: mudanças nele valem assim que entram na `main`, sem release.
- A confirmação do script é interativa; quando rodar pelo Bash tool, só publique se o usuário pedir, e confirme a versão com ele antes.

## Pendências conhecidas

- Os actions do workflow (`goreleaser/goreleaser-action@v6`, `actions/checkout@v4`, `actions/setup-go@v5`) estão fixados por tag, não por SHA de commit. Fixar por SHA é a recomendação de segurança.
- Não existe `lembrete reabrir <id>`; para desfazer um `ok`/`rm` errado, hoje é `excluir` e criar de novo.
