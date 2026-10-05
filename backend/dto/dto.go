// Package dto define o contrato neutro de texto entre os builders
// (binário/store → DTO) e os formatters (DTO → bytes/arquivo).
//
// Os formatters recebem apenas DTO pronto e devolvem DTO;
// quem monta o DTO é o intermediário backend/builders e quem
// aplica o DTO de volta no binário é o applier do domínio.
// Nenhum formatter importa pacotes de domínio concretos.
package dto

import (
	"fmt"
	"sort"
)

// TextRow é uma frase com seus textos por idioma e o PONTEIRO de dupe
// (xxHash64 hex 16 chars) por idioma. O hash NUNCA valida texto: é o
// endereço de conteúdo usado para agrupar repetições.
//
// Origem do ponteiro (dois domínios, nunca misturados no mesmo lote):
//   - RAW (GetCollection → export/import/apply): hash do TEXTO do arquivo
//     (mods-first) — os bytes dos artefatos dependem dele;
//   - DISPLAY (GetEntry): hash do ORIGINAL de data/, criado uma única vez
//     na leitura pristine; coluna Traduzido (Text) reusa o MESMO ponteiro.
//
// Na view, hash(Text) == hash(Original) ⇔ a célula AINDA NÃO FOI traduzida
// (o binário de mods pode ter tradução parcial/total; células intactas
// mantêm o dupe possível). Hash vazio/ausente significa texto vazio.
//
// Index posiciona a row na reconstrução do binário (events: índice da
// string; objects: índice do objeto; macro: índice da string no chunk;
// lockit: posição no grupo do tipo, game 0..G-1 e utf8 G..G+U-1).
// Name qualifica a row quando um índice carrega vários campos
// (objects: chave do segmento — "name", "desc"/"help", "ability1", ...; macro:
// "name"/"name_simplified"; lockit: codificação da linha — "game" ou
// "utf8"). Vazio para events e omitido no JSON.
//
// Original é EXCLUSIVO de exibição (GetEntry): o texto do binário em
// data/, que é a fonte da verdade. Text continua sendo o estado atual
// (mods-first = último save). O campo nunca entra em apply/export — o
// payload de apply leva só index/name/hash/text.
//
// A tabela é a UNIÃO das rows dos dois lados; os dois flags dizem em qual
// lado a row falta e é ele que o frontend pinta em vermelho:
//
//	MissingInOriginal  → só existe na tradução (a coluna Original é vazia);
//	MissingInTranslated → só existe no original (a coluna Traduzido é vazia,
//	                       Text fica nil — o texto nunca vaza de um lado
//	                       para o outro).
//
// Como Original, os flags são de exibição e nunca entram em apply/export.
type TextRow struct {
	Index    int               `json:"index"`
	Name     string            `json:"name,omitempty"`
	Hash     map[string]string `json:"hash,omitempty"`
	Text     map[string]string `json:"text"`
	Original map[string]string `json:"original,omitempty"`
	// MissingInOriginal/MissingInTranslated marcam a row órfã de um dos
	// lados da comparação (união da tabela).
	MissingInOriginal   bool `json:"missingInOriginal,omitempty"`
	MissingInTranslated bool `json:"missingInTranslated,omitempty"`
}

// RefLink é a anotação de EXIBIÇÃO de uma row de referência dedupada
// (Text[us] == "$hash"): o texto atual da def, o pristine dela e onde ela
// vive (arquivo + Index/Name da row). É o que alimenta o texto linkado na
// UI (cor própria), o tooltip "repetição de" e o editor aberto na def
// através do link. Como Original, é exclusivo de exibição e nunca entra em
// apply/export (payload de apply leva só index/name/hash/text).
type RefLink struct {
	Text        string `json:"text"`
	Original    string `json:"original,omitempty"`
	SourceID    string `json:"sourceId,omitempty"`
	SourceIndex int    `json:"sourceIndex"`
	SourceName  string `json:"sourceName,omitempty"`
}

// RowKey devolve a chave estável da row (index:name) — a mesma que o
// frontend usa para rascunho e que o map Refs usa para achar a anotação.
func RowKey(r TextRow) string {
	return fmt.Sprintf("%d:%s", r.Index, r.Name)
}

// FileEntry é um arquivo sem extensão como chave da Collection:
// metadata (mesmos campos dos geradores legados) + rows (ex-strings do JSON).
// Refs é a anotação de EXIBIÇÃO das rows de ref dedupada (rowKey → def);
// só existe na entrega GetEntry/GetVbfTextEntry — o payload de apply
// nunca a carrega.
type FileEntry struct {
	Metadata Metadata           `json:"metadata"`
	Rows     []TextRow          `json:"rows"`
	Refs     map[string]RefLink `json:"refs,omitempty"`
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

// SortRows ordena as rows por Index de forma estável, necessário para a
// reconstrução posicional do binário.
func SortRows(rows []TextRow) {
	sort.SliceStable(rows, func(a, b int) bool {
		return rows[a].Index < rows[b].Index
	})
}
