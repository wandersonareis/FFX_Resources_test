// Diagnóstico de árvore: inventário de arquivos (data/ vs mods/) e
// divergência de estrutura por entrada, com emissão de log em ordem.
//
// Paralelismo é a NÍVEL DE ARQUIVO: os workers coletam o relatório de cada
// arquivo (stat, leitura do original, merge, comparação de rows — todos os
// loops internos sincronos e dependentes de índice) e o CHAMADOR emite os
// logs em ordem canônica. Nada loga de dentro das goroutines.
package services

import (
	"fmt"
	"runtime"
	"sync"

	"ffxresources/backend/common"
	"ffxresources/backend/loggingService"
)

// diagWorkers limita o pool de I/O do inventário/normalização. I/O-bound:
// alguns workers escoam o disco sem estourar memória.
var diagWorkers = runtime.NumCPU()

// DiagRow é uma linha de divergência (sem contraparte) listada no arquivo
// de diagnóstico.
type DiagRow struct {
	Index   int    `json:"index"`
	Name    string `json:"name,omitempty"`
	Snippet string `json:"snippet,omitempty"`
}

func diagRowOf(index int, name, text string) DiagRow {
	return DiagRow{Index: index, Name: name, Snippet: snippet(text)}
}

// snippet corta o texto para reconhecimento visual (~24 runes).
func snippet(text string) string {
	runes := []rune(text)
	if len(runes) == 0 {
		return ""
	}
	max := 24
	if len(runes) < max {
		max = len(runes)
	}
	return string(runes[:max])
}

// entryRef identifica a entrada de origem de um relatório de diagnóstico.
type entryRef struct {
	kind    string
	id      string
	version common.GameVersion
}

func (r entryRef) key() string {
	return fmt.Sprintf("%s/%s (%s)", r.kind, r.id, r.version)
}

// divergeDiag é o relatório de divergência de estrutura entre data/ e o
// estado atual de uma entrada — coletado pelos workers e emitido em ordem.
type divergeDiag struct {
	ref        entryRef
	dataRows   int
	modsRows   int
	onlyInMods []DiagRow
	onlyInData []DiagRow
}

func (d divergeDiag) empty() bool {
	return len(d.onlyInMods) == 0 && len(d.onlyInData) == 0
}

// logDivergence emite o aviso de divergência: console curto colorido,
// arquivo com as listas completas (Index, Name, snippet).
func logDivergence(d divergeDiag) {
	if d.empty() {
		return
	}
	matched := d.modsRows - len(d.onlyInMods)
	loggingService.DiagWarn(d.ref.key(),
		fmt.Sprintf("original e tradução divergem — %d de %d rows casaram (%d rows só em mods, %d rows só em data)",
			matched, d.modsRows, len(d.onlyInMods), len(d.onlyInData)),
		map[string]any{
			"kind":              d.ref.kind,
			"id":                d.ref.id,
			"version":           d.ref.version.String(),
			"data_rows":         d.dataRows,
			"mods_rows":         d.modsRows,
			"rows_somente_mods": d.onlyInMods,
			"rows_somente_data": d.onlyInData,
		})
}

// parallelFor executa fn em paralelo sobre os índices 0..n-1 com pool
// limitado. fn(i) NUNCA loga: a emissão é do chamador, em ordem.
func parallelFor(n int, fn func(i int)) {
	if n <= 0 {
		return
	}
	workers := diagWorkers
	if workers > n {
		workers = n
	}
	if workers < 1 {
		workers = 1
	}
	jobs := make(chan int)
	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func() {
			defer wg.Done()
			for i := range jobs {
				fn(i)
			}
		}()
	}
	for i := 0; i < n; i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
}
