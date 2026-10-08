package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var estiloCursor = lipgloss.NewStyle().Bold(true).Foreground(corTexto).Background(lipgloss.Color("#3A3A3A"))

// lista é um checklist: ↑↓ move, espaço marca, Enter confirma, Esc cancela.
type lista struct {
	titulo  string
	itens   []Lembrete
	marcado map[int]bool // índice em itens
	cursor  int
	pronto  bool
}

func (l lista) Init() tea.Cmd { return nil }

func (l lista) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	k, ok := msg.(tea.KeyMsg)
	if !ok {
		return l, nil
	}

	switch k.String() {
	case "ctrl+c", "esc", "q":
		return l, tea.Quit
	case "up", "k":
		l.cursor = (l.cursor - 1 + len(l.itens)) % len(l.itens)
	case "down", "j":
		l.cursor = (l.cursor + 1) % len(l.itens)
	case " ", "x":
		l.marcado[l.cursor] = !l.marcado[l.cursor]
	case "a":
		todos := len(l.selecionados()) < len(l.itens)
		for i := range l.itens {
			l.marcado[i] = todos
		}
	case "enter":
		// Sem nada marcado, Enter confirma o item sob o cursor.
		if len(l.selecionados()) == 0 {
			l.marcado[l.cursor] = true
		}
		l.pronto = true
		return l, tea.Quit
	}
	return l, nil
}

func (l lista) selecionados() []int {
	var ids []int
	for i, it := range l.itens {
		if l.marcado[i] {
			ids = append(ids, it.ID)
		}
	}
	return ids
}

func (l lista) View() string {
	if l.pronto {
		return ""
	}

	var b strings.Builder
	b.WriteString(estiloAtivo.Render(l.titulo) + "\n\n")
	for i, it := range l.itens {
		caixa := "[ ]"
		if l.marcado[i] {
			caixa = estiloAtivo.Render("[✔]")
		}
		texto := fmt.Sprintf("%-4s %s  %s", fmt.Sprintf("#%d", it.ID), formatar(it.Quando), strings.SplitN(it.Nota, "\n", 2)[0])
		if i == l.cursor {
			texto = estiloCursor.Render(texto)
		}
		b.WriteString(caixa + " " + texto + "\n")
	}
	b.WriteString("\n" + estiloApagado.Render("↑↓ move · espaço marca · a todos · Enter confirma · Esc cancela"))

	return estiloAviso.Render(b.String()) + "\n"
}

// escolherVarios abre o checklist e devolve os ids marcados (nil se cancelou).
func escolherVarios(titulo string, itens []Lembrete) ([]int, error) {
	final, err := tea.NewProgram(lista{titulo: titulo, itens: itens, marcado: map[int]bool{}}, tea.WithOutput(os.Stderr)).Run()
	if err != nil {
		return nil, err
	}
	if l := final.(lista); l.pronto {
		return l.selecionados(), nil
	}
	return nil, nil
}
