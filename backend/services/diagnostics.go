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
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/event"
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

// treePresence é o resultado do stat por arquivo do inventário.
type treePresence struct {
	inData bool
	inMods bool
}

// diagPresenceResolve resolve a presença (data/ vs mods/) das keys em
// paralelo — stat por arquivo, um worker por arquivo, loops internos nenhum.
//
// É DIAGNÓSTICO, não view: stat direto das duas árvores (o events usa o
// store como presença de view; para o inventário o stat é quem conta).
func (s *MetadataService) diagPresenceResolve(kind string, version common.GameVersion, keys []string) []treePresence {
	out := make([]treePresence, len(keys))
	parallelFor(len(keys), func(i int) {
		out[i] = diagFilePresence(kind, keys[i], version)
	})
	return out
}

// diagFilePresence faz o stat direto do binário nas duas árvores.
// Nota: macro compartilha um único binário entre os chunks — a presença
// repete por chunk, e o sumário reflete a unidade de tradução do kind.
func diagFilePresence(kind, id string, version common.GameVersion) treePresence {
	rel, ok := originalRelPath(kind, id, version)
	if !ok {
		return treePresence{}
	}
	inData := accessorExists(rel, common.SourceData)
	inMods := accessorExists(rel, common.SourceMods)
	return treePresence{inData: inData, inMods: inMods}
}

func accessorExists(rel string, src common.FileSource) bool {
	acc, err := common.NewFileAccessorFrom(rel, src)
	return err == nil && acc.Exists
}

// treeDiag acumula o inventário de arquivos de um (kind, versão) e emite o
// sumário de diagnóstico ao final da montagem da árvore.
type treeDiag struct {
	kind    string
	version common.GameVersion

	inData  int      // arquivos em data/ (fonte da verdade)
	inMods  int      // arquivos em mods/ (tradução)
	missing []string // em data/ SEM tradução (falta para traduzir)
	extra   []string // só em mods/ (ausente em data/ — regra 4, investigar)
}

func newTreeDiag(kind string, version common.GameVersion) *treeDiag {
	return &treeDiag{kind: kind, version: version}
}

func (t *treeDiag) add(p treePresence, key string, rel string) {
	if p.inData {
		t.inData++
	}
	if p.inMods {
		t.inMods++
		if !p.inData {
			t.extra = append(t.extra, fmt.Sprintf("%s (%s)", key, rel))
		}
	}
	if p.inData && !p.inMods {
		t.missing = append(t.missing, key)
	}
}

// emit loga o inventário: INFO sempre (o tradutor sabe quanto falta), WARN
// para cada arquivo só em mods/ (ausente em data/ — investigação). Listas
// completas ficam no arquivo de diagnóstico.
func (t *treeDiag) emit() {
	if t.inData == 0 && t.inMods == 0 {
		return // kind sem arquivos em nenhuma árvore: nada a reportar
	}
	loggingService.Info("%s %s: arquivos em data/ = %d — tradução em mods/ = %d — faltam %d para traduzir",
		t.kind, t.version, t.inData, t.inMods, len(t.missing))

	for _, key := range t.extra {
		loggingService.Warn("mods com arquivo ausente em data/ — não é exibido (regra 4): %s", key)
	}

	details := map[string]any{
		"kind":    t.kind,
		"version": t.version.String(),
	}
	level := func() string {
		if len(t.missing) > 0 || len(t.extra) > 0 {
			return "warn"
		}
		return "info"
	}()
	if level == "warn" {
		details["arquivos_sem_traducao"] = t.missing
		details["arquivos_somente_mods"] = t.extra
		loggingService.DiagWarn(fmt.Sprintf("%s/%s", t.kind, t.version),
			fmt.Sprintf("inventário: %d em data/, %d traduzidos, %d sem tradução, %d só em mods",
				t.inData, t.inMods, len(t.missing), len(t.extra)),
			details)
		return
	}
	loggingService.DiagInfo(fmt.Sprintf("%s/%s", t.kind, t.version),
		fmt.Sprintf("inventário: %d em data/, %d traduzidos, %d sem tradução, %d só em mods",
			t.inData, t.inMods, len(t.missing), len(t.extra)),
		details)
}

// emitTreeDiag resolve a presença das keys em PARALELO (stat por arquivo),
// acumula o inventário e emite o sumário: INFO sempre (o tradutor sabe
// quanto falta), WARN por arquivo só em mods/ (ausente em data/ —
// investigação). extraMods soma arquivos só em mods/ que não são keys da
// coleção (ex.: sobras de eventos na árvore mods/ — invisíveis na árvore).
func (s *MetadataService) emitTreeDiag(kind string, version common.GameVersion, keys []string, extraModsRel []string) {
	diag := newTreeDiag(kind, version)
	presence := s.diagPresenceResolve(kind, version, keys)
	for i, k := range keys {
		rel, _ := originalRelPath(kind, k, version)
		diag.add(presence[i], k, rel)
	}
	diag.extra = append(diag.extra, extraModsRel...)
	diag.emit()
}

// modsOnlyFilesForEvents varre a árvore mods/ de eventos e devolve os
// arquivos que NÃO existem em data/ — um evento descoberto só pelo mods é
// invisível para a árvore (a descoberta é data-driven), então só o log
// denuncia. Devolve caminhos relativos à raiz de localização.
func modsOnlyFilesForEvents(version common.GameVersion) []string {
	lroot := common.GetLocalizationRootForVersion(version, common.DefaultLocalization)
	modsRoot := filepath.Join(common.GameFilesRoot, common.ModsFolder, lroot, "event")
	dataIds := map[string]bool{}
	for _, id := range event.GetAllEventIDs(version) {
		dataIds[strings.ToLower(id)] = true
	}

	var extra []string
	walkErr := filepath.WalkDir(modsRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // diretório ausente: nada a reportar
		}
		if d.IsDir() || !strings.EqualFold(filepath.Ext(d.Name()), ".bin") {
			return nil
		}
		id := strings.TrimSuffix(d.Name(), filepath.Ext(d.Name()))
		if dataIds[strings.ToLower(id)] {
			return nil
		}
		rel, rerr := filepath.Rel(filepath.Join(common.GameFilesRoot, common.ModsFolder, lroot), path)
		if rerr != nil {
			rel = path
		}
		extra = append(extra, rel)
		return nil
	})
	if walkErr != nil {
		common.LogWarning("varredura de mods/ para eventos falhou: %v", walkErr)
	}
	sort.Strings(extra)
	return extra
}
