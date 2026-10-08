package main

import (
	"cmp"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
)

const (
	statusPendente = "pendente"
	statusFeito    = "feito"
	statusIgnorado = "ignorado"
)

type Lembrete struct {
	ID     int    `json:"id"`
	Quando int64  `json:"quando"` // epoch
	Status string `json:"status"`
	Nota   string `json:"nota"`
}

func arquivo() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".lembretes.json")
}

func carregar() ([]Lembrete, error) {
	dados, err := os.ReadFile(arquivo())
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var lembretes []Lembrete
	err = json.Unmarshal(dados, &lembretes)
	slices.SortFunc(lembretes, func(a, b Lembrete) int { return cmp.Compare(a.Quando, b.Quando) })
	return lembretes, err
}

// salvar grava num arquivo temporário e renomeia, para não corromper o arquivo se algo falhar no meio.
func salvar(lembretes []Lembrete) error {
	dados, err := json.MarshalIndent(lembretes, "", "  ")
	if err != nil {
		return err
	}

	tmp := arquivo() + ".tmp"
	if err := os.WriteFile(tmp, dados, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, arquivo())
}

func proximoID(lembretes []Lembrete) int {
	maior := 0
	for _, l := range lembretes {
		maior = max(maior, l.ID)
	}
	return maior + 1
}
