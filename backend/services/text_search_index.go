package services

import (
	"sort"
	"strings"
	"unicode/utf8"

	"ffxresources/backend/common"
	"ffxresources/backend/dto"
)

// MaxSearchFiles limita quantos ARQUIVOS por consulta entram no resultado
// (as rows de cada arquivo são ilimitadas). Variável de propósito: os
// testes ajustam para exercitar o corte sem fixture grande.
var MaxSearchFiles = 100

// searchSnippetContext é o bytes de contexto antes/depois do match no
// snippet exibido no modal de busca.
const searchSnippetContext = 80

// TextSearchMatch identifica uma row encontrada e em qual fonte o texto
// apareceu. Index + Name é a identidade estável da row entre formatos.
//
// O snippet vem JÁ PARTIDO em três campos (antes/hit/depois) na caixa
// original do arquivo: o frontend só envolve o hit em <mark>. Partir aqui
// evita repassar offsets de bytes que o JS (UTF-16) não alinharia.
type TextSearchMatch struct {
	Index         int    `json:"index"`
	Name          string `json:"name,omitempty"`
	Data          bool   `json:"data"`
	Mods          bool   `json:"mods"`
	SnippetBefore string `json:"snippetBefore,omitempty"`
	SnippetHit    string `json:"snippetHit,omitempty"`
	SnippetAfter  string `json:"snippetAfter,omitempty"`
}

// TextSearchResult é um arquivo com uma ou mais rows que deram match.
type TextSearchResult struct {
	Kind string            `json:"kind"`
	ID   string            `json:"id"`
	Rows []TextSearchMatch `json:"rows"`
}

// TextSearchResponse é a resposta completa de uma consulta: os arquivos
// (cortados em MaxSearchFiles, com Truncated) mais os TOTAIS contados sobre
// a base inteira — é o "N rows em M arquivos" exibido no modal.
type TextSearchResponse struct {
	Results    []TextSearchResult `json:"results"`
	TotalRows  int                `json:"totalRows"`
	TotalFiles int                `json:"totalFiles"`
	Truncated  bool               `json:"truncated"`
}

// indexedTextRow guarda o texto us em DUAS formas: a cópia minúscula para o
// match (busca substring barata) e o texto ORIGINAL para o snippet — não
// guarda DTOs completos, hashes nem os demais idiomas.
type indexedTextRow struct {
	index    int
	name     string
	data     string
	mods     string
	dataOrig string
	modsOrig string
}

type indexedTextFile struct {
	kind string
	id   string
	rows []indexedTextRow
}

type textSearchIndex struct {
	files  []indexedTextFile
	stamps map[string]indexedSourceStamps
}

type indexedRowID struct {
	index int
	name  string
}

func newIndexedTextRow(index int, name, data, mods string) indexedTextRow {
	return indexedTextRow{
		index:    index,
		name:     name,
		data:     strings.ToLower(data),
		mods:     strings.ToLower(mods),
		dataOrig: data,
		modsOrig: mods,
	}
}

// indexedFileFromEntries junta as duas fontes pela identidade da row, sem
// confundir campos diferentes que compartilham o mesmo índice.
func indexedFileFromEntries(kind, id string, data, mods *dto.FileEntry) indexedTextFile {
	rowsByID := make(map[indexedRowID]indexedTextRow)
	add := func(entry *dto.FileEntry, source common.FileSource) {
		if entry == nil {
			return
		}
		for _, row := range entry.Rows {
			text := row.Text[common.DefaultLocalization]
			if text == "" {
				continue
			}
			key := indexedRowID{index: row.Index, name: row.Name}
			indexed := rowsByID[key]
			indexed.index = row.Index
			indexed.name = row.Name
			if source == common.SourceData {
				indexed.data = strings.ToLower(text)
				indexed.dataOrig = text
			} else {
				indexed.mods = strings.ToLower(text)
				indexed.modsOrig = text
			}
			rowsByID[key] = indexed
		}
	}
	add(data, common.SourceData)
	add(mods, common.SourceMods)

	file := indexedTextFile{kind: kind, id: id, rows: make([]indexedTextRow, 0, len(rowsByID))}
	for _, row := range rowsByID {
		file.rows = append(file.rows, row)
	}
	return file
}

// search percorre apenas strings já decodificadas/normalizadas. A ordenação
// explícita mantém a resposta estável independentemente da ordem dos mapas.
// O corte em MaxSearchFiles limita só o payload: os totais cobrem a base
// inteira para o modal exibir "mostrando N de M".
func (index textSearchIndex) search(query string) TextSearchResponse {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return TextSearchResponse{}
	}

	response := TextSearchResponse{Results: make([]TextSearchResult, 0)}
	for _, file := range index.files {
		result := TextSearchResult{Kind: file.kind, ID: file.id}
		for _, row := range file.rows {
			match := TextSearchMatch{Index: row.index, Name: row.name}
			match.Data = strings.Contains(row.data, query)
			match.Mods = strings.Contains(row.mods, query)
			if !match.Data && !match.Mods {
				continue
			}
			// Snippet só para as rows que entram no payload.
			if len(response.Results) < MaxSearchFiles {
				match.withSnippet(row, query)
			}
			result.Rows = append(result.Rows, match)
		}
		if len(result.Rows) == 0 {
			continue
		}
		response.TotalFiles++
		response.TotalRows += len(result.Rows)
		if len(response.Results) >= MaxSearchFiles {
			response.Truncated = true
			continue
		}
		sort.Slice(result.Rows, func(a, b int) bool {
			if result.Rows[a].Index != result.Rows[b].Index {
				return result.Rows[a].Index < result.Rows[b].Index
			}
			return result.Rows[a].Name < result.Rows[b].Name
		})
		response.Results = append(response.Results, result)
	}
	sort.Slice(response.Results, func(a, b int) bool {
		if response.Results[a].Kind != response.Results[b].Kind {
			return response.Results[a].Kind < response.Results[b].Kind
		}
		return response.Results[a].ID < response.Results[b].ID
	})
	return response
}

// withSnippet corta o texto ORIGINAL em volta do match, partindo o resultado
// em antes/hit/depois. A posição vem da cópia minúscula (é onde o match foi
// feito); os cortes são aparados em fronteira de rune para nunca partir um
// caractere multibyte no meio.
func (m *TextSearchMatch) withSnippet(row indexedTextRow, query string) {
	lower := row.data
	original := row.dataOrig
	if m.Mods {
		lower = row.mods
		original = row.modsOrig
	}
	pos := strings.Index(lower, query)
	if pos < 0 || pos > len(original) {
		return
	}
	// A caixa pode mudar o tamanho em bytes em Unicode raro: se os offsets
	// não caem em fronteira de rune, aparar para o início do caractere.
	for pos > 0 && !utf8.RuneStart(original[pos]) {
		pos--
	}
	end := pos + len(query)
	if end > len(original) {
		end = len(original)
	}
	for end < len(original) && !utf8.RuneStart(original[end]) {
		end++
	}

	start := pos - searchSnippetContext
	if start < 0 {
		start = 0
	}
	for start > 0 && !utf8.RuneStart(original[start]) {
		start--
	}
	tail := end + searchSnippetContext
	if tail > len(original) {
		tail = len(original)
	}
	for tail < len(original) && !utf8.RuneStart(original[tail]) {
		tail++
	}

	m.SnippetBefore = original[start:pos]
	if start > 0 {
		m.SnippetBefore = "…" + m.SnippetBefore
	}
	m.SnippetHit = original[pos:end]
	m.SnippetAfter = original[end:tail]
	if tail < len(original) {
		m.SnippetAfter += "…"
	}
}
