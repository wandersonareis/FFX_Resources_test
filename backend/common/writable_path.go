package common

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ErrVbfImmutable sinaliza tentativa de gravar DENTRO de um container .vbf.
var ErrVbfImmutable = fmt.Errorf("container .vbf é somente leitura")

// CheckWritablePath valida o destino de uma gravação do app.
//
// Regras do projeto:
//
//  1. O .vbf é a FONTE dos binários reais (arquivos de ~19GB) e é somente
//     leitura: nunca se grava dentro de um *.vbf — a leitura extrai, o
//     resultado vai para mods/.
//  2. data/ é a árvore extraída desses binários, também original: dentro da
//     árvore do jogo (GameFilesRoot) só mods/ pode ser gravado.
//  3. Fora da árvore do jogo o destino é do chamador (cópia de teste, export
//     escolhido pelo usuário) — a regra 1 continua valendo.
//
// É a rede de proteção dos escritores: cada formatador monta o caminho
// correto (mods/) e esta função garante que nenhum caminho de escrita vire a
// tocar no container ou no original.
func CheckWritablePath(path string) error {
	clean := filepath.Clean(path)
	if strings.EqualFold(filepath.Ext(clean), ".vbf") {
		return fmt.Errorf("%w: %s", ErrVbfImmutable, clean)
	}
	if GameFilesRoot == "" {
		return nil
	}
	// Lowercase dos dois lados: volume/raiz variam em maiúsculas no Windows e
	// o Rel falharia (ou marcaria "fora") por isso.
	root := filepath.Clean(strings.ToLower(GameFilesRoot))
	target := strings.ToLower(clean)
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return nil // volumes diferentes: fora da árvore do jogo
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return nil // fora da árvore do jogo
	}
	// ModsFolder é "mods/" (com barra): limpa antes de comparar componentes.
	mods := strings.ToLower(filepath.Clean(ModsFolder))
	if rel == mods || strings.HasPrefix(rel, mods+string(filepath.Separator)) {
		return nil
	}
	return fmt.Errorf(
		"gravação fora de mods/ bloqueada: %s (data/ e o .vbf são somente leitura; destino válido é mods/)",
		clean)
}
