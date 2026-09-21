// Package dto define o contrato neutro de texto entre os builders
// (binário/store → DTO) e os formatters (DTO → bytes/arquivo).
//
// Os formatters recebem apenas DTO pronto e devolvem DTO;
// quem monta o DTO é o intermediário backend/builders e quem
// aplica o DTO de volta no binário é o applier do domínio.
// Nenhum formatter importa pacotes de domínio concretos.
package dto

import "sort"

// TextRow é uma frase com seus textos por idioma e o hash xxHash64
// (hex 16 chars) de cada frase. Hash vazio/ausente significa texto vazio.
//
// Index posiciona a row na reconstrução do binário (events: índice da
// string; objects: índice do objeto; macro: índice da string no chunk).
// Name qualifica a row quando um índice carrega vários campos
// (objects: chave do segmento — "name", "ability1", ...; macro:
// "name"/"simplifiedName"). Vazio para events e omitido no JSON.
type TextRow struct {
	Index int               `json:"index"`
	Name  string            `json:"name,omitempty"`
	Hash  map[string]string `json:"hash,omitempty"`
	Text  map[string]string `json:"text"`
}

// FileEntry é um arquivo sem extensão como chave da Collection:
// metadata (mesmos campos dos geradores legados) + rows (ex-strings do JSON).
type FileEntry struct {
	Metadata Metadata  `json:"metadata"`
	Rows     []TextRow `json:"rows"`
}

// Collection é o documento textual completo: chave = nome do arquivo
// sem extensão (events: eventID; objects: basename; macro: chunk_N).
type Collection map[string]FileEntry

// SortedKeys devolve as chaves ordenadas para saída determinística.
func (c Collection) SortedKeys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// SortRows ordena as rows por Index (desempate por Name), necessário para
// reconstrução do binário na mesma ordem posicional e saída determinística.
func SortRows(rows []TextRow) {
	sort.SliceStable(rows, func(a, b int) bool {
		if rows[a].Index != rows[b].Index {
			return rows[a].Index < rows[b].Index
		}
		return rows[a].Name < rows[b].Name
	})
}
