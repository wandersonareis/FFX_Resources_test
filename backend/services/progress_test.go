package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ffxresources/backend/common"
	"ffxresources/backend/core/progress"
	"ffxresources/backend/fileFormats/event"
	"ffxresources/backend/loggingService"
)

func sprintf(format string, args ...any) string { return fmt.Sprintf(format, args...) }

// fakeReporter é a ponte de teste do core/progress: captura os eventos sem
// tocar no wails.
type fakeReporter struct {
	events []string
}

func (f *fakeReporter) record(format string, args ...any) {
	f.events = append(f.events, sprintf(format, args...))
}

func (f *fakeReporter) Begin(label string, total int) { f.record("begin %s total=%d", label, total) }
func (f *fakeReporter) Step(item string)              { f.record("step %s", item) }
func (f *fakeReporter) Issue(item, message string)    { f.record("issue %s: %s", item, message) }
func (f *fakeReporter) End()                          { f.record("end") }

// TestLoadFromBinaryEmitsProgress: a carga de eventos emite Begin com o
// total da descoberta, um Step por evento e End — pelo canal neutro.
func TestLoadFromBinaryEmitsProgress(t *testing.T) {
	fake := &fakeReporter{}
	progress.Set(fake)
	t.Cleanup(func() { progress.Set(nil) })

	root := t.TempDir()
	prevRoot := common.GameFilesRoot
	common.GameFilesRoot = root
	t.Cleanup(func() { common.GameFilesRoot = prevRoot })

	if err := ensureVersionReady(common.GameVersionFFX2); err != nil {
		t.Fatalf("prepare version: %v", err)
	}

	seedEventsSynthetic(common.GameVersionFFX2, "zev001", []string{"Frase do zev001"})
	seedEventsSynthetic(common.GameVersionFFX2, "zev002", []string{"Frase do zev002"})
	seedEventsData(t, common.GameVersionFFX2, "zev001") // export → mods + copy → data
	seedEventsData(t, common.GameVersionFFX2, "zev002")

	b := event.NewEventsBinaryFile(common.GameVersionFFX2)
	if err := b.LoadFromBinary(); err != nil {
		t.Fatalf("load: %v", err)
	}

	joined := strings.Join(fake.events, "\n")
	if !strings.Contains(joined, "begin Carregando eventos… total=2") {
		t.Fatalf("begin com o total da descoberta ausente:\n%s", joined)
	}
	if got := strings.Count(joined, "step "); got != 2 {
		t.Fatalf("esperava 2 steps, achei %d:\n%s", got, joined)
	}
	if !strings.Contains(joined, "\nend") || !strings.HasSuffix(joined, "end") {
		t.Fatalf("end ausente:\n%s", joined)
	}
}

// TestLoadFromBinarySkipsBrokenEvent: evento com binário corrompido é
// pulado com Issue (toast/log individual) e o restante carrega.
func TestLoadFromBinarySkipsBrokenEvent(t *testing.T) {
	root := t.TempDir()
	prevRoot := common.GameFilesRoot
	common.GameFilesRoot = root
	t.Cleanup(func() { common.GameFilesRoot = prevRoot })

	if err := ensureVersionReady(common.GameVersionFFX2); err != nil {
		t.Fatalf("prepare version: %v", err)
	}

	seedEventsSynthetic(common.GameVersionFFX2, "zev001", []string{"Frase do zev001"})
	seedEventsSynthetic(common.GameVersionFFX2, "zev002", []string{"Frase do zev002"})
	seedEventsData(t, common.GameVersionFFX2, "zev001")
	seedEventsData(t, common.GameVersionFFX2, "zev002")

	// zev002: TODOS os binários de leitura (data/ por localização) viram
	// DIRETÓRIOS → a leitura falha em todos, o evento sai vazio e é PULADO
	// (Issue individual) enquanto o resto carrega.
	for loc := range common.SupportedLanguages {
		lroot := common.GetLocalizationRootForVersion(common.GameVersionFFX2, loc)
		rel, _ := event.EventRelPath("zev002")
		broken := filepath.Join(common.GameFilesRoot, lroot, rel)
		if err := os.Remove(broken); err != nil {
			t.Fatalf("remover binário %s: %v", broken, err)
		}
		if err := os.MkdirAll(broken, 0o755); err != nil {
			t.Fatalf("transformar em diretório: %v", err)
		}
	}

	prevMods := common.DisableMods
	common.DisableMods = true
	t.Cleanup(func() { common.DisableMods = prevMods })

	fake := &fakeReporter{}
	progress.Set(fake)
	t.Cleanup(func() { progress.Set(nil) })

	b := event.NewEventsBinaryFile(common.GameVersionFFX2)
	if err := b.LoadFromBinary(); err != nil {
		t.Fatalf("load: %v", err)
	}

	joined := strings.Join(fake.events, "\n")
	t.Logf("DEBUG fake.events=%#v objects=%d", fake.events, b.Objects.Len())
	if !strings.Contains(joined, "issue zev002") {
		t.Fatalf("issue do arquivo corrompido ausente:\n%s", joined)
	}
	if !strings.Contains(joined, "step zev001") {
		t.Fatalf("evento válido deveria continuar carregando:\n%s", joined)
	}
	if b.Objects == nil || b.Objects.Len() != 1 {
		t.Fatalf("a carga deveria seguir com o evento válido: %v", b.Objects)
	}
}

// ProgressService: máquina de estados Begin/Step/End + Issue com cap de
// toasts por ciclo (o arquivo recebe tudo).
func TestProgressServiceCycle(t *testing.T) {
	loggingService.ResetForTest(t.TempDir())
	t.Cleanup(func() { loggingService.ResetForTest(t.TempDir()) })

	p := NewProgressService(nil, nil) // ctx vazio: EventsEmit é no-op
	p.Begin("Exportando events (ffx2)…", 3)
	if _, total, processed, pct, _ := p.stateSnapshot(); processed != 0 || total != 3 || pct != 0 {
		t.Fatalf("begin: processed=%d total=%d pct=%d", processed, total, pct)
	}

	p.Step("zev001")
	p.Step("zev002")
	if _, _, processed, pct, item := p.stateSnapshot(); processed != 2 || pct != 66 || item != "zev002" {
		t.Fatalf("step: processed=%d pct=%d item=%q", processed, pct, item)
	}

	// End encerra: Steps fora de ciclo são ignorados.
	p.End()
	p.Step("depois-do-end")
	if _, _, processed, _, _ := p.stateSnapshot(); processed != 2 {
		t.Fatalf("step após end deveria ser ignorado (processed=%d)", processed)
	}
	p.End() // idempotente
}

// PreloadVersions é idempotente e aborta quando a árvore muda no meio.
func TestPreloadVersionsIdempotentAndInvalidated(t *testing.T) {

	// Ambiente sem gamefiles válidos: a pré-carga falha por versão — mas o
	// Set marca/consome igualmente (idempotência é sobre a fila, não sobre
	// o sucesso).
	version := common.GameVersionFFX2
	if !markPreloaded(version) {
		t.Fatal("primeira marca deveria ser nova")
	}
	if markPreloaded(version) {
		t.Fatal("segunda marca deveria ser ignorada")
	}

	treeGeneration.Add(1)
	InvalidateViewCaches() // geração nova zera o Set de pré-carga
	if !markPreloaded(version) {
		t.Fatal("invalidação deveria liberar a versão de novo")
	}
}
