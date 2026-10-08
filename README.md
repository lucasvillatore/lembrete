# lembrete

CLI para lembrar das coisas **no próprio terminal**. Você cria um lembrete com dia e hora e, a partir desse momento, ele aparece **toda vez que você abre um terminal novo**, até você marcar como feito ou ignorado.

Feito em Go com [Bubble Tea](https://github.com/charmbracelet/bubbletea) e [Lip Gloss](https://github.com/charmbracelet/lipgloss).

## Como fica

**Ao abrir o terminal**, se houver lembretes vencidos:

```
╭────────────────────────────────────────────────────╮
│  🔔 Lembretes                                      │
│                                                    │
│    #1   sex 09/10 09:00  pagar a conta de luz      │
│                          vence dia 10              │
│                                                    │
│    #3   sex 09/10 14:00  revisar PR do time        │
│                                                    │
│  lembrete ok <id> · rm <id> · ls                   │
╰────────────────────────────────────────────────────╯
```

**Criando um lembrete** (`lembrete`): calendário, seletor de hora e editor de nota numa tela só.

```
╭───────────────────────────────────────╮ ╭──────────────────────────────────────────────╮
│                                       │ │                                              │
│          ‹  Outubro 2026  ›           │ │  nota                                        │
│                                       │ │                                              │
│   dom  seg  ter  qua  qui  sex  sáb   │ │  pagar a conta de luz                        │
│                        1    2    3    │ │  vence dia 10                                │
│    4    5    6    7    8  [ 9]  10    │ │                                              │
│   11   12   13   14   15   16   17    │ │                                              │
│   18   19   20   21   22   23   24    │ │                                              │
│   25   26   27   28   29   30   31    │ │                                              │
│                                       │ │                                              │
│         ╭────╮   ╭────╮               │ │                                              │
│  hora   │ 09 │ : │ 00 │               │ │                                              │
│         ╰────╯   ╰────╯               │ │                                              │
╰───────────────────────────────────────╯ ╰──────────────────────────────────────────────╯
   ←↑↓→ dia · PgUp/PgDn mês · t hoje · Enter → hora · Tab próximo · Ctrl+S salva · Esc sai
```

**Marcando como feito** (`lembrete ok`): checklist para marcar vários de uma vez.

```
╭───────────────────────────────────────────────────────────────────╮
│  Marcar como feito                                                │
│                                                                   │
│  [✔] #1   sex 09/10 09:00  pagar a conta de luz                   │
│  [ ] #2   sex 09/10 11:00  daily                                  │
│  [✔] #3   sex 09/10 14:00  revisar PR do time                     │
│                                                                   │
│  ↑↓ move · espaço marca · a todos · Enter confirma · Esc cancela  │
╰───────────────────────────────────────────────────────────────────╯
```

**Listando** (`lembrete ls`): pendentes, ✔ feitos e ~~ignorados~~ (riscados).

```
  #1   sex 09/10 09:00  pagar a conta de luz
✔ #2   sex 09/10 11:00  daily
  #3   sex 09/10 14:00  revisar PR do time          ← riscado quando ignorado
```

## Comandos

| Comando | O que faz |
|---|---|
| `lembrete` | Abre a tela para criar um lembrete |
| `lembrete add --note "texto" [--date DD/MM] [--time HH:MM]` | Cria direto, sem tela. Sem data/hora usa hoje/agora |
| `lembrete ls` | Lista todos, ordenados por data |
| `lembrete ok [id]` | Marca como feito. Sem id, abre o checklist |
| `lembrete rm [id]` | Marca como ignorado (fica riscado). Sem id, abre o checklist |
| `lembrete excluir [id]` | Apaga de vez. Sem id, pergunta qual |
| `lembrete check` | Mostra os pendentes vencidos (é o que roda ao abrir o terminal) |

`--date` aceita `09/10`, `09/10/2026`, `9/10` ou `2026-10-09`. Datas no passado são permitidas.

```bash
lembrete add --note "pagar a conta de luz" --date 09/10 --time 09:00
lembrete add --note $'linha 1\nlinha 2'          # nota com várias linhas
```

### Teclas da tela de criação

| Onde | Teclas |
|---|---|
| Calendário | ←↑↓→ (ou h/j/k/l) dia · PgUp/PgDn (ou `[` `]`) mês · `t` hoje · Enter vai para a hora |
| Hora | ↑↓ ajusta (minuto de 5 em 5) · ←→ hora/minuto · digite os números (`1` `4` = 14) · Enter vai para a nota |
| Nota | Editor normal: Enter quebra linha, Ctrl+V cola |
| Geral | Tab / Shift+Tab troca de campo · **Ctrl+S salva** · Esc cancela |

## Instalação

Linux e macOS, sem precisar de Go:

```bash
curl -fsSL https://raw.githubusercontent.com/lucasvillatore/lembrete/main/install.sh | sh
```

O script baixa o binário do último [release](https://github.com/lucasvillatore/lembrete/releases) para o seu sistema, instala em `~/.local/bin` e adiciona no `~/.zshrc` ou `~/.bashrc` a linha que mostra os lembretes ao abrir o terminal:

```bash
command -v lembrete >/dev/null && lembrete check
```

Para instalar em outra pasta: `curl -fsSL .../install.sh | LEMBRETE_DIR=/outra/pasta sh`. Para atualizar, rode o mesmo comando de novo.

Se `~/.local/bin` não estiver no seu `PATH`, o script avisa. Adicione `export PATH="$HOME/.local/bin:$PATH"` no `~/.zshrc` ou `~/.bashrc`.

### Com Go

```bash
go install github.com/lucasvillatore/lembrete@latest
```

### Compilar do código

```bash
git clone https://github.com/lucasvillatore/lembrete.git
cd lembrete
go build -o lembrete .
ln -sf "$PWD/lembrete" ~/.local/bin/lembrete
```

### Publicar uma versão nova

Na `main` atualizada, rode o script. Ele confere se a árvore está limpa e se compila, mostra os commits desde a última versão, pede confirmação, cria a tag e acompanha o build. O GitHub Actions compila para Linux e macOS (amd64 e arm64) com o [GoReleaser](https://goreleaser.com) e publica no Releases.

```bash
./scripts/release.sh            # patch: v0.1.0 -> v0.1.1
./scripts/release.sh minor      # v0.1.0 -> v0.2.0
./scripts/release.sh major      # v0.1.0 -> v1.0.0
./scripts/release.sh v0.3.0     # versão exata
```

Mudanças na `main` entram por pull request.

## Onde ficam os dados

Em `~/.lembretes.json`, um JSON simples com `id`, `quando` (epoch), `status` (`pendente`, `feito`, `ignorado`) e `nota`. Dá para fazer backup ou editar na mão.

## Código

| Arquivo | Conteúdo |
|---|---|
| `main.go` | Comandos e formatação da listagem/aviso |
| `tui.go` | Tela de criação: calendário, hora e nota |
| `lista.go` | Checklist do `ok` / `rm` |
| `store.go` | Leitura e gravação do `~/.lembretes.json` |
