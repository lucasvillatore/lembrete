package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

const uso = `lembrete — lembretes que aparecem ao abrir o terminal

  lembrete                 cria um lembrete (abre o calendário)
  lembrete add --note "texto" [--date DD/MM[/AAAA]] [--time HH:MM]
                           cria direto, sem tela (padrão: hoje, agora)
  lembrete ls              lista todos
  lembrete ok [id]         marca como feito (sem id: checklist, marca vários)
  lembrete rm [id]         marca como ignorado, fica riscado no ls (sem id: checklist)
  lembrete excluir [id]    apaga de vez (sem id: pergunta qual)
  lembrete check           mostra os pendentes vencidos (usado no .zshrc)`

var (
	estiloAviso = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(corDestaque).Padding(0, 2)
	estiloID    = lipgloss.NewStyle().Bold(true).Foreground(corDestaque)
	estiloData  = lipgloss.NewStyle().Foreground(corHoje)
	estiloFeito = lipgloss.NewStyle().Foreground(corApagado)
	estiloRisco = lipgloss.NewStyle().Foreground(corApagado).Strikethrough(true)
)

func main() {
	a := append(os.Args[1:], "", "")
	cmd, arg := a[0], a[1]

	var err error
	switch cmd {
	case "", "add":
		if len(os.Args) > 2 {
			err = criarInline(os.Args[2:])
		} else {
			err = criar()
		}
	case "ls":
		err = listar()
	case "check":
		err = checar()
	case "ok":
		err = mudarStatus(arg, statusFeito)
	case "rm":
		err = mudarStatus(arg, statusIgnorado)
	case "excluir":
		err = excluir(arg)
	case "-h", "--help", "help":
		fmt.Println(uso)
	default:
		err = fmt.Errorf("comando desconhecido: %s\n\n%s", cmd, uso)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func criar() error {
	quando, nota, ok, err := escolher()
	if err != nil {
		return err
	}
	if !ok {
		fmt.Println("Cancelado.")
		return nil
	}
	return gravar(quando, nota)
}

// criarInline: lembrete add --note "..." [--date ...] [--time HH:MM]. Sem data/hora, usa as atuais.
func criarInline(args []string) error {
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	nota := fs.String("note", "", "texto do lembrete (obrigatório)")
	data := fs.String("date", "", "DD/MM, DD/MM/AAAA ou AAAA-MM-DD (padrão: hoje)")
	hora := fs.String("time", "", "HH:MM (padrão: agora)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*nota) == "" {
		return fmt.Errorf("--note é obrigatório")
	}

	agora := time.Now()
	dia := time.Date(agora.Year(), agora.Month(), agora.Day(), 0, 0, 0, 0, time.Local)
	if *data != "" {
		d, err := parseData(*data, agora.Year())
		if err != nil {
			return err
		}
		dia = d
	}

	h, min := agora.Hour(), agora.Minute()
	if *hora != "" {
		t, err := time.Parse("15:04", *hora)
		if err != nil {
			return fmt.Errorf("hora inválida %q, use HH:MM", *hora)
		}
		h, min = t.Hour(), t.Minute()
	}

	return gravar(time.Date(dia.Year(), dia.Month(), dia.Day(), h, min, 0, 0, time.Local), strings.TrimSpace(*nota))
}

func parseData(s string, anoAtual int) (time.Time, error) {
	for _, layout := range []string{"2/1/2006", "2006-01-02", "2/1"} {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			if t.Year() == 0 { // "DD/MM" sem ano
				t = t.AddDate(anoAtual, 0, 0)
			}
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("data inválida %q, use DD/MM, DD/MM/AAAA ou AAAA-MM-DD", s)
}

func gravar(quando time.Time, nota string) error {
	lembretes, err := carregar()
	if err != nil {
		return err
	}

	l := Lembrete{ID: proximoID(lembretes), Quando: quando.Unix(), Status: statusPendente, Nota: nota}
	if err := salvar(append(lembretes, l)); err != nil {
		return err
	}

	fmt.Printf("Lembrete %s criado para %s\n", estiloID.Render(fmt.Sprintf("#%d", l.ID)), estiloData.Render(formatar(l.Quando)))
	return nil
}

func listar() error {
	lembretes, err := carregar()
	if err != nil {
		return err
	}
	if len(lembretes) == 0 {
		fmt.Println("Nenhum lembrete.")
		return nil
	}

	for _, l := range lembretes {
		fmt.Println(linha(l))
	}
	return nil
}

func checar() error {
	lembretes, err := carregar()
	if err != nil {
		return err
	}

	agora := time.Now().Unix()
	var vencidos []string
	for _, l := range lembretes {
		if l.Status == statusPendente && l.Quando <= agora {
			vencidos = append(vencidos, linha(l))
		}
	}
	if len(vencidos) == 0 {
		return nil
	}

	corpo := estiloAtivo.Render("🔔 Lembretes") + "\n\n" + strings.Join(vencidos, "\n\n") +
		"\n\n" + estiloApagado.Render("lembrete ok <id> · rm <id> · ls")
	fmt.Println(estiloAviso.Render(corpo))
	return nil
}

func mudarStatus(arg, status string) error {
	lembretes, err := carregar()
	if err != nil {
		return err
	}

	var ids []int
	if arg != "" {
		id, err := strconv.Atoi(strings.TrimPrefix(arg, "#"))
		if err != nil {
			return fmt.Errorf("id inválido: %s", arg)
		}
		ids = []int{id}
	} else {
		pendentes := slices.DeleteFunc(slices.Clone(lembretes), func(l Lembrete) bool { return l.Status != statusPendente })
		if len(pendentes) == 0 {
			fmt.Println("Nenhum lembrete pendente.")
			return nil
		}
		titulo := map[string]string{statusFeito: "Marcar como feito", statusIgnorado: "Ignorar"}[status]
		if ids, err = escolherVarios(titulo, pendentes); err != nil || len(ids) == 0 {
			return err
		}
	}

	for _, id := range ids {
		i := indice(lembretes, id)
		if i < 0 {
			return fmt.Errorf("lembrete #%d não existe", id)
		}
		lembretes[i].Status = status
	}
	if err := salvar(lembretes); err != nil {
		return err
	}
	for _, id := range ids {
		fmt.Printf("#%d marcado como %s.\n", id, status)
	}
	return nil
}

func excluir(arg string) error {
	lembretes, err := carregar()
	if err != nil {
		return err
	}

	id, err := resolverID(arg, lembretes, func(Lembrete) bool { return true })
	if err != nil || id == 0 {
		return err
	}

	i := indice(lembretes, id)
	if i < 0 {
		return fmt.Errorf("lembrete #%d não existe", id)
	}
	if err := salvar(slices.Delete(lembretes, i, i+1)); err != nil {
		return err
	}
	fmt.Printf("#%d excluído.\n", id)
	return nil
}

func indice(lembretes []Lembrete, id int) int {
	return slices.IndexFunc(lembretes, func(l Lembrete) bool { return l.ID == id })
}

// resolverID usa o id do argumento ou, sem argumento, lista os candidatos e pergunta. 0 = nada escolhido.
func resolverID(arg string, lembretes []Lembrete, candidato func(Lembrete) bool) (int, error) {
	if arg != "" {
		id, err := strconv.Atoi(strings.TrimPrefix(arg, "#"))
		if err != nil {
			return 0, fmt.Errorf("id inválido: %s", arg)
		}
		return id, nil
	}

	var opcoes []Lembrete
	for _, l := range lembretes {
		if candidato(l) {
			opcoes = append(opcoes, l)
		}
	}
	if len(opcoes) == 0 {
		fmt.Println("Nenhum lembrete para escolher.")
		return 0, nil
	}

	for _, l := range opcoes {
		fmt.Println(linha(l))
	}

	fmt.Print("\nid (Enter cancela): ")
	resp, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	resp = strings.TrimSpace(resp)
	if resp == "" {
		return 0, nil
	}
	return resolverID(resp, lembretes, candidato)
}

// linha formata um lembrete: "#3  sex 09/10 09:00  primeira linha da nota" (+ demais linhas indentadas).
func linha(l Lembrete) string {
	texto := strings.ReplaceAll(l.Nota, "\n", "\n"+strings.Repeat(" ", 24))
	id, data := fmt.Sprintf("%-4s", fmt.Sprintf("#%d", l.ID)), formatar(l.Quando)

	switch l.Status {
	case statusFeito:
		return estiloFeito.Render("✔ " + id + " " + data + "  " + texto)
	case statusIgnorado:
		return estiloRisco.Render("  " + id + " " + data + "  " + texto)
	}
	return "  " + estiloID.Render(id) + " " + estiloData.Render(data) + "  " + texto
}

func formatar(epoch int64) string {
	t := time.Unix(epoch, 0)
	layout := "02/01 15:04"
	if t.Year() != time.Now().Year() {
		layout = "02/01/2006 15:04"
	}
	return semana[t.Weekday()] + " " + t.Format(layout)
}
