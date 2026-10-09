package common

import (
	"os"
	"path/filepath"
	"strings"
)

// PruneEmptyDirs sobe a partir de `dir` removendo diretórios VAZIOS até o
// piso: `floor` é o último nível que nunca sai (a subida para ele encerra).
//
// É a contraparte de EnsurePathExists: o delete apagava só o arquivo e
// deixava o ramo que a gravação havia criado. A poda é a última etapa do
// delete daquele arquivo — NÃO é varredura de faxina: nada aqui caminha a
// árvore procurando o que limpar, só o ramo do arquivo que saiu.
//
// Best-effort por natureza: erro de leitura/permissão apenas interrompe a
// subida, para não transformar um delete concluído em falha.
//
// Só desce para dentro de floor (confere por filepath.Rel), então um floor
// errado derruba no máximo o ramo indicado, nunca árvores vizinhas.
func PruneEmptyDirs(dir, floor string) {
	f := filepath.Clean(floor)
	for d := filepath.Clean(dir); strictlyBelow(d, f); d = filepath.Dir(d) {
		entries, err := os.ReadDir(d)
		if err != nil || len(entries) > 0 {
			return // ilegível ou com conteúdo (arquivo oculto já basta)
		}
		if err := os.Remove(d); err != nil {
			return
		}
	}
}

// strictlyBelow relata se path é estritamente descendente de floor: igual ao
// piso ou fora dele (inclusive em volume/raiz diferente, onde Rel falha)
// devolve false, então a poda nunca alcança o piso nem o vizinho.
func strictlyBelow(path, floor string) bool {
	rel, err := filepath.Rel(floor, path)
	if err != nil {
		return false
	}
	return rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
