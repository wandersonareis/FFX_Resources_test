package services

import (
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"ffxresources/backend/common"
	"ffxresources/backend/dto"

	"golang.org/x/text/unicode/norm"
)

// MaxSearchFiles limita quantos ARQUIVOS por consulta entram no resultado
// (as rows de cada arquivo são ilimitadas). Variável de propósito: os
// testes ajustam para exercitar o corte sem fixture grande.
var MaxSearchFiles = 100

// searchSnippetContext é o raio (em RUNES) de contexto antes/depois do
// match no snippet exibido no modal de busca.
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

// indexedTextRow guarda o texto us em DUAS formas: a cópia DOBRADA para o
// match (minúscula, sem diacríticos — "difícil" casa "dificil") e o texto
// ORIGINAL para o snippet — não guarda DTOs completos, hashes nem os
// demais idiomas.
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

// FoldSearchText devolve a forma de MATCHING: minúsculas sem diacríticos
// ("difícil" -> "dificil", "coração" -> "coracao"). Minúsculas primeiro,
// depois NFD e remoção das marcas combinantes (Mn) — NFD, não NFKD: não
// dobra ligaduras nem largura de fonte. É a MESMA implementação usada para
// montar o índice, a consulta e o mapa de offsets do snippet.
func FoldSearchText(s string) string {
	folded, _ := foldOffsets(s)
	return folded
}

// foldOffsets devolve o texto dobrado (ver FoldSearchText) e, para cada
// rune DOBRADO, o índice do rune ORIGINAL de onde veio — o mapa que permite
// cortar o snippet no texto original usando um match feito no espaço
// dobrado (onde é "coracao") e ainda mostrar o acento de "coração".
func foldOffsets(s string) (string, []int32) {
	runes := []rune(s)
	var folded strings.Builder
	origins := make([]int32, 0, len(runes))
	for origIndex, r := range runes {
		for _, d := range norm.NFD.String(string(r)) {
			if unicode.Is(unicode.Mn, d) {
				continue
			}
			folded.WriteRune(unicode.ToLower(d))
			origins = append(origins, int32(origIndex))
		}
	}
	return folded.String(), origins
}

func newIndexedTextRow(index int, name, data, mods string) indexedTextRow {
	return indexedTextRow{
		index:    index,
		name:     name,
		data:     FoldSearchText(data),
		mods:     FoldSearchText(mods),
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
				indexed.data = FoldSearchText(text)
				indexed.dataOrig = text
			} else {
				indexed.mods = FoldSearchText(text)
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
	// Matching SEM acentos nos dois sentidos: "dificil" acha "difícil" e
	// vice-versa (a dobra é a mesma do índice).
	query = FoldSearchText(strings.TrimSpace(query))
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
// em antes/hit/depois — o hit mantém a caixa e os acentos ORIGINAIS (a
// consulta "coracao" acha e mostra "coração"). O match foi feito no espaço
// DOBRADO; o mapa de origens devolve os runes originais correspondentes,
// então todos os cortes caem em fronteira de rune por construção.
func (m *TextSearchMatch) withSnippet(row indexedTextRow, query string) {
	folded, original := row.data, row.dataOrig
	if m.Mods {
		folded, original = row.mods, row.modsOrig
	}
	pos := strings.Index(folded, query)
	if pos < 0 {
		return
	}
	// A cópia dobrada do original é byte-idêntica à indexada (mesma
	// função), então o mapa de origens indexa o MESMO espaço do match.
	_, origins := foldOffsets(original)
	foldedRuneStart := utf8.RuneCountInString(folded[:pos])
	queryRunes := utf8.RuneCountInString(query)
	if foldedRuneStart >= len(origins) || foldedRuneStart+queryRunes > len(origins) {
		return
	}
	origRunes := []rune(original)
	hitStart := int(origins[foldedRuneStart])
	hitEnd := hitStart + 1
	if last := foldedRuneStart + queryRunes - 1; last > foldedRuneStart {
		hitEnd = int(origins[last]) + 1
	}

	start := hitStart - searchSnippetContext
	if start < 0 {
		start = 0
	}
	tail := hitEnd + searchSnippetContext
	if tail > len(origRunes) {
		tail = len(origRunes)
	}

	m.SnippetBefore = string(origRunes[start:hitStart])
	if start > 0 {
		m.SnippetBefore = "…" + m.SnippetBefore
	}
	m.SnippetHit = string(origRunes[hitStart:hitEnd])
	m.SnippetAfter = string(origRunes[hitEnd:tail])
	if tail < len(origRunes) {
		m.SnippetAfter += "…"
	}
}
