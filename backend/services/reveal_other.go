//go:build !windows && !darwin

package services

import (
	"os/exec"
	"path/filepath"
)

// revealFileOnOS abre o gerenciador de arquivos (Linux e demais sistemas).
//
// `xdg-open` não tem um equivalente confiável de "destacar o item" — ele abre
// a pasta que contém o arquivo. O caminho do ITEM não é útil aqui, só o do
// diretório. Comportamento herdado do código anterior, isolado atrás da build
// tag: qualquer GOOS novo que ainda não tenha revelação própria cai aqui de
// forma explícita, e não por commutação silenciosa em runtime.
func revealFileOnOS(abs string) error {
	// Sem Wait(): o xdg-open pode devolver código de saída não zero depois de
	// abrir a janela, e o app não deve ficar preso junto com o gerenciador.
	return exec.Command("xdg-open", filepath.Dir(abs)).Start()
}
