package services

import (
	"fmt"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

// buildTestIndex monta um índice sintético com `files` arquivos, cada um com
// uma única row cujo texto é "match <id> " + padding + "fim". Serve para
// exercitar o corte de payload e o snippet sem fixture de jogo.
func buildTestIndex(files int) textSearchIndex {
	index := textSearchIndex{}
	for i := 0; i < files; i++ {
		id := fmt.Sprintf("file%03d", i)
		row := newIndexedTextRow(i, "name", "xx match "+id+" "+strings.Repeat("pad ", 40)+"fim", "")
		index.files = append(index.files, indexedTextFile{kind: KindEvents, id: id, rows: []indexedTextRow{row}})
	}
	return index
}

func TestSearchCountsTotalsAndTruncatesPayload(t *testing.T) {
	previous := MaxSearchFiles
	MaxSearchFiles = 3
	defer func() { MaxSearchFiles = previous }()

	response := buildTestIndex(10).search("match")

	if response.TotalFiles != 10 {
		t.Fatalf("TotalFiles = %d, want 10 (os totais cobrem a base inteira)", response.TotalFiles)
	}
	if response.TotalRows != 10 {
		t.Fatalf("TotalRows = %d, want 10", response.TotalRows)
	}
	if len(response.Results) != 3 {
		t.Fatalf("len(Results) = %d, want 3 (corte em MaxSearchFiles)", len(response.Results))
	}
	if !response.Truncated {
		t.Fatal("Truncated = false, want true")
	}
	for _, result := range response.Results {
		for _, match := range result.Rows {
			if match.SnippetHit == "" {
				t.Fatalf("row do payload sem SnippetHit: %+v", match)
			}
		}
	}
}

func TestSearchEmptyQueryReturnsNothing(t *testing.T) {
	response := buildTestIndex(2).search("   ")
	if response.Results != nil || response.TotalFiles != 0 || response.Truncated {
		t.Fatalf("consulta vazia deveria devolver resposta vazia: %+v", response)
	}
}

func TestSearchSnippetCutsAroundMatchWithEllipsis(t *testing.T) {
	long := strings.Repeat("a", 200) + " Needle " + strings.Repeat("b", 200)
	index := textSearchIndex{files: []indexedTextFile{{
		kind: KindEvents,
		id:   "long",
		rows: []indexedTextRow{newIndexedTextRow(1, "n", long, "")},
	}}}

	match := index.search("needle").Results[0].Rows[0]

	if !strings.EqualFold(match.SnippetHit, "needle") {
		t.Fatalf("SnippetHit = %q, want a ocorrência", match.SnippetHit)
	}
	if !strings.HasPrefix(match.SnippetBefore, "…") || !strings.HasSuffix(match.SnippetAfter, "…") {
		t.Fatalf("snippet longo deveria ter reticências nas pontas: before=%q after=%q",
			match.SnippetBefore, match.SnippetAfter)
	}
	if len(match.SnippetBefore) > searchSnippetContext+len("…")+utf8.UTFMax {
		t.Fatalf("contexto antes estourou o raio: %d bytes", len(match.SnippetBefore))
	}
	if got := match.SnippetBefore + match.SnippetHit + match.SnippetAfter; !strings.Contains(got, "Needle") {
		t.Fatalf("snippet perdeu a caixa original: %q", got)
	}
}

func TestSearchSnippetKeepsWholeShortRow(t *testing.T) {
	index := textSearchIndex{files: []indexedTextFile{{
		kind: KindObjects,
		id:   "short",
		rows: []indexedTextRow{newIndexedTextRow(2, "", "Sword +5", "Espada curta +5")},
	}}}

	response := index.search("espada")
	match := response.Results[0].Rows[0]
	if !match.Mods || match.Data {
		t.Fatalf("flags erradas: %+v", match)
	}
	if match.SnippetBefore != "" || match.SnippetHit != "Espada" || match.SnippetAfter != " curta +5" {
		t.Fatalf("snippet curto deveria vir inteiro e sem reticências: %+v", match)
	}
}

func TestSearchSnippetSlicesRunesOnBoundaries(t *testing.T) {
	// "日本語" muda de tamanho em bytes entre caixas — os cortes precisam
	// permanecer em fronteira de rune (nunca partir um byte de multibyte).
	text := strings.Repeat("あ", 100) + "Palavra" + strings.Repeat("い", 100)
	index := textSearchIndex{files: []indexedTextFile{{
		kind: KindEvents,
		id:   "utf8",
		rows: []indexedTextRow{newIndexedTextRow(0, "", text, "")},
	}}}

	for _, match := range index.search("palavra").Results[0].Rows {
		joined := match.SnippetBefore + match.SnippetHit + match.SnippetAfter
		if !utf8.ValidString(joined) {
			t.Fatalf("snippet com rune partido no meio: %q", joined)
		}
	}
}

func TestIndexFreshThrottleReusesVerdict(t *testing.T) {
	previous := TextSearchFreshCheckInterval
	TextSearchFreshCheckInterval = time.Hour
	defer func() { TextSearchFreshCheckInterval = previous }()

	// Base vazia → veredito "frescor" true, cacheado pela hora do intervalo.
	slot := &textSearchIndexSlot{index: textSearchIndex{stamps: map[string]indexedSourceStamps{}}}
	if !slot.indexFresh() {
		t.Fatal("índice sem stamps é sempre fresco — sanity")
	}
	// A base fica velha (um arquivo indexado sumiu do disco), mas DENTRO do
	// intervalo a consulta reaproveita o veredito cacheado — sem novo stat.
	slot.index.stamps["sumiu.bin"] = indexedSourceStamps{hasData: true}
	if !slot.indexFresh() {
		t.Fatal("veredito cacheado valia true dentro do intervalo")
	}
	// Com o intervalo zerado a revalidação roda de novo e vê a mudança.
	TextSearchFreshCheckInterval = 0
	if slot.indexFresh() {
		t.Fatal("revalidação deveria ver o stamp ausente e marcar o índice velho")
	}
}
