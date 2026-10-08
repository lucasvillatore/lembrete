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

Precisa do [Go](https://go.dev/dl/) 1.22 ou mais novo.

```bash
git clone git@github.com:lucasvillatore/lembrete.git
cd lembrete
go build -o lembrete .

# deixa o comando disponível no PATH
mkdir -p ~/.local/bin
ln -sf "$PWD/lembrete" ~/.local/bin/lembrete
```

Confira se `~/.local/bin` está no seu `PATH` (`echo $PATH`). Se não estiver, adicione `export PATH="$HOME/.local/bin:$PATH"` no `~/.zshrc` ou `~/.bashrc`.

### Mostrar os lembretes ao abrir o terminal

Adicione no final do `~/.zshrc` (ou `~/.bashrc`):

```bash
command -v lembrete >/dev/null && lembrete check
```

Pronto: todo terminal novo mostra os lembretes vencidos. Se não houver nenhum, não aparece nada.

### Atualizar

```bash
cd lembrete && git pull && go build -o lembrete .
```

O link no PATH aponta para o binário, então a versão nova vale na hora.

## Onde ficam os dados

Em `~/.lembretes.json`, um JSON simples com `id`, `quando` (epoch), `status` (`pendente`, `feito`, `ignorado`) e `nota`. Dá para fazer backup ou editar na mão.

## Código

| Arquivo | Conteúdo |
|---|---|
| `main.go` | Comandos e formatação da listagem/aviso |
| `tui.go` | Tela de criação: calendário, hora e nota |
| `lista.go` | Checklist do `ok` / `rm` |
| `store.go` | Leitura e gravação do `~/.lembretes.json` |
