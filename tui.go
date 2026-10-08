package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	meses  = [...]string{"Janeiro", "Fevereiro", "Março", "Abril", "Maio", "Junho", "Julho", "Agosto", "Setembro", "Outubro", "Novembro", "Dezembro"}
	semana = [...]string{"dom", "seg", "ter", "qua", "qui", "sex", "sáb"}
	ajuda  = [...]string{
		"←↑↓→ dia · PgUp/PgDn mês · t hoje · Enter → hora · Tab próximo · Ctrl+S salva · Esc sai",
		"↑↓ ajusta · ←→ hora/minuto · digite os números · Enter → nota · Tab próximo · Ctrl+S salva · Esc sai",
		"editor: Enter quebra linha · Ctrl+V cola · Tab próximo · Ctrl+S salva · Esc sai",
	}
)

const (
	focoCal = iota
	focoHora
	focoNota
)

const passoMin = 5 // ↑↓ no minuto anda de 5 em 5

var (
	corDestaque = lipgloss.Color("#7D56F4")
	corHoje     = lipgloss.Color("#04B5D6")
	corApagado  = lipgloss.Color("#585858")
	corFim      = lipgloss.Color("#8A8A8A")
	corTexto    = lipgloss.Color("#E4E4E4")

	estiloPainel  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2)
	estiloTitulo  = lipgloss.NewStyle().Bold(true).Foreground(corTexto)
	estiloAtivo   = lipgloss.NewStyle().Bold(true).Foreground(corDestaque)
	estiloApagado = lipgloss.NewStyle().Foreground(corApagado)

	estiloCelula     = lipgloss.NewStyle().Width(5).Align(lipgloss.Center)
	estiloSemana     = estiloCelula.Foreground(corApagado)
	estiloSel        = estiloCelula.Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Background(corDestaque)
	estiloSelInativo = estiloCelula.Bold(true).Foreground(corTexto).Background(lipgloss.Color("#3A3A3A"))
	estiloHoje       = estiloCelula.Bold(true).Foreground(corHoje)
	estiloFim        = estiloCelula.Foreground(corFim)
	estiloPassado    = estiloCelula.Foreground(corApagado)
	estiloCampo      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1).Bold(true).Foreground(corTexto)
)

type model struct {
	hoje, dia         time.Time
	hora, minuto      int
	campoMinuto       bool // no seletor de hora: false = hora, true = minuto
	foco              int
	nota              textarea.Model
	largura, altura   int
	pronto, cancelado bool
}

func novoModel() model {
	agora := time.Now()
	hoje := time.Date(agora.Year(), agora.Month(), agora.Day(), 0, 0, 0, 0, time.Local)

	ta := textarea.New()
	ta.Placeholder = "o que lembrar?"
	ta.ShowLineNumbers = false
	ta.CharLimit = 0
	ta.SetWidth(40)
	ta.SetHeight(9)
	ta.Prompt = ""
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	return model{hoje: hoje, dia: hoje, hora: 9, nota: ta}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.largura, m.altura = msg.Width, msg.Height
		// O painel da nota ocupa 2/3 da largura (descontando borda + padding = 6 colunas).
		m.nota.SetWidth(max(30, m.largura*2/3-6))
		// Altura: a tela toda, menos borda/padding/título do painel (6), ajuda (2) e margem (2).
		m.nota.SetHeight(max(9, m.altura-10))
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelado = true
			return m, tea.Quit
		case "ctrl+s":
			if strings.TrimSpace(m.nota.Value()) == "" {
				return m.focar(focoNota)
			}
			m.pronto = true
			return m, tea.Quit
		case "tab":
			return m.focar((m.foco + 1) % 3)
		case "shift+tab":
			return m.focar((m.foco + 2) % 3)
		}

		if m.foco == focoNota {
			var cmd tea.Cmd
			m.nota, cmd = m.nota.Update(msg)
			return m, cmd
		}

		if m.foco == focoHora {
			return m.teclaHora(msg.String())
		}

		switch msg.String() {
		case "enter":
			return m.focar(focoHora)
		case "left", "h":
			m.mover(0, -1)
		case "right", "l":
			m.mover(0, 1)
		case "up", "k":
			m.mover(0, -7)
		case "down", "j":
			m.mover(0, 7)
		case "pgup", "[":
			m.mover(-1, 0)
		case "pgdown", "]":
			m.mover(1, 0)
		case "t":
			m.dia = m.hoje
		}
		return m, nil
	}

	// Demais mensagens (ex.: piscar do cursor) vão para o editor.
	var cmd tea.Cmd
	m.nota, cmd = m.nota.Update(msg)
	return m, cmd
}

func (m model) teclaHora(k string) (tea.Model, tea.Cmd) {
	switch k {
	case "enter":
		return m.focar(focoNota)
	case "left", "right", "h", "l", ":":
		m.campoMinuto = !m.campoMinuto
	case "up", "k":
		m.ajustar(1)
	case "down", "j":
		m.ajustar(-1)
	default:
		// Digitar números: "1" depois "4" vira 14; se passar do limite, recomeça com o último dígito.
		if len(k) == 1 && k[0] >= '0' && k[0] <= '9' {
			v, limite := m.campo()
			if *v = (*v%10)*10 + int(k[0]-'0'); *v >= limite {
				*v = int(k[0] - '0')
			}
		}
	}
	return m, nil
}

// campo devolve o campo ativo do seletor de hora e o limite dele.
func (m *model) campo() (*int, int) {
	if m.campoMinuto {
		return &m.minuto, 60
	}
	return &m.hora, 24
}

func (m *model) ajustar(d int) {
	v, limite := m.campo()
	if m.campoMinuto {
		d *= passoMin
	}
	*v = (*v + d + limite) % limite
}

func (m model) focar(f int) (tea.Model, tea.Cmd) {
	m.foco = f
	if f == focoNota {
		return m, m.nota.Focus()
	}
	m.nota.Blur()
	return m, nil
}

// mover anda N meses ou N dias, mantendo o dia dentro do mês. Datas passadas são permitidas.
func (m *model) mover(nMeses, nDias int) {
	novo := m.dia.AddDate(0, 0, nDias)
	if nMeses != 0 {
		primeiro := time.Date(m.dia.Year(), m.dia.Month()+time.Month(nMeses), 1, 0, 0, 0, 0, time.Local)
		ultimo := primeiro.AddDate(0, 1, -1).Day()
		novo = time.Date(primeiro.Year(), primeiro.Month(), min(m.dia.Day(), ultimo), 0, 0, 0, 0, time.Local)
	}
	m.dia = novo
}

func (m model) celula(d time.Time) string {
	txt := fmt.Sprintf("%2d", d.Day())
	switch {
	case d.Equal(m.dia) && m.foco == focoCal:
		return estiloSel.Render(txt)
	case d.Equal(m.dia):
		return estiloSelInativo.Render(txt)
	case d.Before(m.hoje):
		return estiloPassado.Render(txt)
	case d.Equal(m.hoje):
		return estiloHoje.Render(txt)
	case d.Weekday() == time.Saturday || d.Weekday() == time.Sunday:
		return estiloFim.Render(txt)
	}
	return estiloCelula.Render(txt)
}

func (m model) painel(ativo bool, conteudo string) string {
	cor := corApagado
	if ativo {
		cor = corDestaque
	}
	return estiloPainel.BorderForeground(cor).Render(conteudo)
}

func (m model) calendario() string {
	var b strings.Builder

	titulo := fmt.Sprintf("‹  %s %d  ›", meses[m.dia.Month()-1], m.dia.Year())
	b.WriteString(lipgloss.PlaceHorizontal(35, lipgloss.Center, estiloRotulo(m.foco == focoCal).Render(titulo)) + "\n\n")

	for _, s := range semana {
		b.WriteString(estiloSemana.Render(s))
	}
	b.WriteString("\n")

	primeiro := time.Date(m.dia.Year(), m.dia.Month(), 1, 0, 0, 0, 0, time.Local)
	ultimo := primeiro.AddDate(0, 1, -1).Day()
	pad := int(primeiro.Weekday())
	b.WriteString(strings.Repeat(estiloCelula.Render(""), pad))
	semanas := 1
	for d := 1; d <= ultimo; d++ {
		b.WriteString(m.celula(time.Date(primeiro.Year(), primeiro.Month(), d, 0, 0, 0, 0, time.Local)))
		if (pad+d)%7 == 0 && d != ultimo {
			b.WriteString("\n")
			semanas++
		}
	}
	// Mantém a altura fixa (6 semanas) para o layout não pular entre meses.
	b.WriteString(strings.Repeat("\n", 6-semanas+1) + "\n")

	b.WriteString(m.seletorHora())

	return m.painel(m.foco != focoNota, b.String())
}

func (m model) editor() string {
	return m.painel(m.foco == focoNota, estiloRotulo(m.foco == focoNota).Render("nota")+"\n\n"+m.nota.View())
}

func estiloRotulo(ativo bool) lipgloss.Style {
	if ativo {
		return estiloAtivo
	}
	return estiloTitulo
}

// seletorHora desenha [HH] : [MM]; o campo ativo fica com a borda e o fundo destacados.
func (m model) seletorHora() string {
	campo := func(valor int, ativo bool) string {
		e := estiloCampo.BorderForeground(corApagado)
		if m.foco == focoHora {
			e = e.BorderForeground(corTexto)
			if ativo {
				e = e.BorderForeground(corDestaque).Background(corDestaque).Foreground(lipgloss.Color("#FFFFFF"))
			}
		}
		return e.Render(fmt.Sprintf("%02d", valor))
	}

	return lipgloss.JoinHorizontal(lipgloss.Center,
		estiloRotulo(m.foco == focoHora).Render("hora   "),
		campo(m.hora, !m.campoMinuto),
		estiloTitulo.Render(" : "),
		campo(m.minuto, m.campoMinuto),
	)
}

func (m model) View() string {
	if m.pronto || m.cancelado {
		return ""
	}

	tela := lipgloss.JoinHorizontal(lipgloss.Top, m.calendario(), " ", m.editor())
	tela = lipgloss.JoinVertical(lipgloss.Center, tela, "", estiloApagado.Render(ajuda[m.foco]))

	if m.largura == 0 {
		return tela
	}
	return lipgloss.Place(m.largura, m.altura, lipgloss.Center, lipgloss.Center, tela)
}

// escolher abre a tela de calendário + nota. ok = false se o usuário cancelou.
func escolher() (quando time.Time, nota string, ok bool, err error) {
	// Tela alternativa: limpa o terminal enquanto roda e restaura ao sair.
	final, err := tea.NewProgram(novoModel(), tea.WithAltScreen(), tea.WithOutput(os.Stderr)).Run()
	if err != nil {
		return time.Time{}, "", false, err
	}

	m := final.(model)
	if !m.pronto {
		return time.Time{}, "", false, nil
	}

	quando = time.Date(m.dia.Year(), m.dia.Month(), m.dia.Day(), m.hora, m.minuto, 0, 0, time.Local)
	return quando, strings.TrimSpace(m.nota.Value()), true, nil
}
