//go:build darwin

package services

import (
	"os/exec"
)

// revealFileOnOS destaca o arquivo no Finder (macOS).
//
// `open -R <arquivo>` é o equivalente ao "revelar" do Explorer: abre a pasta
// que contém o item e o deixa selecionado. Comportamento herdado do código
// anterior, agora isolado atrás da build tag para que um GOOS errado não
// compile esta chamada por engano.
func revealFileOnOS(abs string) error {
	// Sem Wait(): o `open` pode devolver código de saída não zero depois de
	// abrir a janela, e o app não deve ficar preso junto com o Finder.
	return exec.Command("open", "-R", abs).Start()
}
