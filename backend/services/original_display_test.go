package services

import (
	"os"
	"path/filepath"
	"testing"

	"ffxresources/backend/common"
	"ffxresources/backend/dto"
	"ffxresources/backend/fileFormats/event"
	"ffxresources/backend/formatters/hash"
)

// displayRowIsRef: ref quando texto é "$h" e h == hash próprio da row.
func displayRowIsRef(r dto.TextRow) bool {
	bare, ok := hash.Strip(r.Text[common.DefaultLocalization])
	return ok && bare == r.Hash[common.DefaultLocalization]
}

const sharedOriginal = "Frase compartilhada entre eventos"

// seedDisplayRoot prepara o ambiente sintético comum aos testes de display.
func seedDisplayRoot(t *testing.T) *MetadataService {
	t.Helper()
	seedEventsTempRoot(t)
	event.ClearEvents(common.GameVersionFFX2)
	t.Cleanup(func() { event.ClearEvents(common.GameVersionFFX2) })
	if err := ensureVersionReady(common.GameVersionFFX2); err != nil {
		t.Fatalf("prepare version: %v", err)
	}
	clearDedupViewCache()
	t.Cleanup(clearDedupViewCache)
	return NewMetadataService(nil)
}

// seedEventsData grava o binário dos events no gamefiles e o espelha em
// data/ (o export vai sempre para mods/; data/ é a fonte do original).
func seedEventsData(t *testing.T, version common.GameVersion, id string) {
	t.Helper()
	if err := event.ExportEventStringsToLocalizations(version, id); err != nil {
		t.Fatalf("export %s: %v", id, err)
	}
	rel, err := event.EventRelPath(id)
	if err != nil {
		t.Fatalf("relpath %s: %v", id, err)
	}
	for loc := range common.SupportedLanguages {
		lroot := common.GetLocalizationRootForVersion(version, loc)
		modsPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, lroot, rel)
		dataPath := filepath.Join(common.GameFilesRoot, lroot, rel)
		b, rerr := os.ReadFile(modsPath)
		if rerr != nil {
			t.Fatalf("ler mods %s: %v", modsPath, rerr)
		}
		if err := os.MkdirAll(filepath.Dir(dataPath), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dataPath, err)
		}
		if err := os.WriteFile(dataPath, b, 0o644); err != nil {
			t.Fatalf("gravar data %s: %v", dataPath, err)
		}
	}
}

// mutateStoreEvent troca o texto 'us' de strings[0] no STORE (sem gravar).
func mutateStoreEvent(t *testing.T, version common.GameVersion, id, newText string) {
	t.Helper()
	ev := event.GetEvent(version, id)
	if ev == nil || len(ev.Strings) == 0 || ev.Strings[0] == nil {
		t.Fatalf("evento %s não semeado", id)
	}
	fs := ev.Strings[0].GetLocalizedContent(common.DefaultLocalization)
	if fs == nil {
		t.Fatalf("evento %s sem conteúdo 'us'", id)
	}
	fs.SetRegularString(newText)
}

// TestGetEntryDisplayPointerIsOriginalHash: def traduzida (mods) + twin não
// traduzida — o grupo estrutural sobrevive à tradução da def e a twin
// colapsa em ref do PONTEIRO DO ORIGINAL.
func TestGetEntryDisplayPointerIsOriginalHash(t *testing.T) {
	svc := seedDisplayRoot(t)
	seedEventsSynthetic(common.GameVersionFFX2, "zev001", []string{sharedOriginal})
	seedEventsSynthetic(common.GameVersionFFX2, "zev002", []string{sharedOriginal})

	// data/ recebe o ORIGINAL pristine dos dois eventos.
	seedEventsData(t, common.GameVersionFFX2, "zev001")
	seedEventsData(t, common.GameVersionFFX2, "zev002")

	// zev001 é traduzido no store e gravado em mods/ (zev002 fica sem mods:
	// continua não traduzido, lendo o original de data/).
	const translation = "Tradução da frase compartilhada"
	mutateStoreEvent(t, common.GameVersionFFX2, "zev001", translation)
	if err := event.ExportEventStringsToLocalizations(common.GameVersionFFX2, "zev001"); err != nil {
		t.Fatalf("export pós-tradução: %v", err)
	}
	clearDedupViewCache()

	defEntry, err := svc.GetEntry(KindEvents, "zev001", common.GameVersionFFX2)
	if err != nil {
		t.Fatalf("get def: %v", err)
	}
	wantPointer := hash.Sum64Hex(sharedOriginal)
	if got := defEntry.Rows[0].Hash[common.DefaultLocalization]; got != wantPointer {
		t.Fatalf("ponteiro da def deveria ser o do ORIGINAL: %q != %q", got, wantPointer)
	}
	if got := defEntry.Rows[0].Text[common.DefaultLocalization]; got != translation {
		t.Fatalf("def deveria carregar a tradução: %q", got)
	}
	if displayRowIsRef(defEntry.Rows[0]) {
		t.Fatal("def nunca deveria colapsar em ref")
	}

	after, err := svc.GetEntry(KindEvents, "zev002", common.GameVersionFFX2)
	if err != nil {
		t.Fatalf("get twin: %v", err)
	}
	if !displayRowIsRef(after.Rows[0]) {
		t.Fatalf("twin não traduzida deveria colapsar sob a def traduzida: %q", after.Rows[0].Text["us"])
	}
	if bare, _ := hash.Strip(after.Rows[0].Text[common.DefaultLocalization]); bare != wantPointer {
		t.Fatalf("ref deveria apontar para o ponteiro do original: %q != %q", bare, wantPointer)
	}
}

// TestGetEntryTranslatedTwinStaysLiteral: traduções divergentes NÃO
// colapsam — dupe sobre texto já traduzido é proibido por decisão.
func TestGetEntryTranslatedTwinStaysLiteral(t *testing.T) {
	svc := seedDisplayRoot(t)
	seedEventsSynthetic(common.GameVersionFFX2, "zev001", []string{sharedOriginal})
	seedEventsSynthetic(common.GameVersionFFX2, "zev002", []string{sharedOriginal})

	seedEventsData(t, common.GameVersionFFX2, "zev001")
	seedEventsData(t, common.GameVersionFFX2, "zev002")

	const a = "Tradução A da frase"
	const b = "Tradução B divergente"
	mutateStoreEvent(t, common.GameVersionFFX2, "zev001", a)
	if err := event.ExportEventStringsToLocalizations(common.GameVersionFFX2, "zev001"); err != nil {
		t.Fatalf("export zev001: %v", err)
	}
	mutateStoreEvent(t, common.GameVersionFFX2, "zev002", b)
	if err := event.ExportEventStringsToLocalizations(common.GameVersionFFX2, "zev002"); err != nil {
		t.Fatalf("export zev002: %v", err)
	}
	clearDedupViewCache()

	e1, err := svc.GetEntry(KindEvents, "zev001", common.GameVersionFFX2)
	if err != nil {
		t.Fatalf("get zev001: %v", err)
	}
	e2, err := svc.GetEntry(KindEvents, "zev002", common.GameVersionFFX2)
	if err != nil {
		t.Fatalf("get zev002: %v", err)
	}

	if displayRowIsRef(e1.Rows[0]) {
		t.Fatalf("tradução A deveria permanecer literal: %q", e1.Rows[0].Text["us"])
	}
	if displayRowIsRef(e2.Rows[0]) {
		t.Fatalf("tradução B divergente deveria permanecer literal: %q", e2.Rows[0].Text["us"])
	}
	if got := e1.Rows[0].Text[common.DefaultLocalization]; got != a {
		t.Fatalf("tradução A: %q", got)
	}
	if got := e2.Rows[0].Text[common.DefaultLocalization]; got != b {
		t.Fatalf("tradução B div: %q", got)
	}

	// E as duas compartilham o PONTEIRO do mesmo original: a identidade
	// estrutural sobrevive às traduções.
	wantPointer := hash.Sum64Hex(sharedOriginal)
	if got := e1.Rows[0].Hash[common.DefaultLocalization]; got != wantPointer {
		t.Fatalf("ponteiro A: %q != %q", got, wantPointer)
	}
	if got := e2.Rows[0].Hash[common.DefaultLocalization]; got != wantPointer {
		t.Fatalf("ponteiro B: %q != %q", got, wantPointer)
	}
}

// stripEventMods remove o binário de mods/ do evento: simula a entrada NÃO
// traduzida (pristine só em data/) — o cenário em que o HashOrder cru nasce
// sem a coluna Original para a def.
func stripEventMods(t *testing.T, version common.GameVersion, id string) {
	t.Helper()
	rel, err := event.EventRelPath(id)
	if err != nil {
		t.Fatalf("relpath %s: %v", id, err)
	}
	for loc := range common.SupportedLanguages {
		lroot := common.GetLocalizationRootForVersion(version, loc)
		modsPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, lroot, rel)
		if err := os.Remove(modsPath); err != nil && !os.IsNotExist(err) {
			t.Fatalf("remover mods %s: %v", modsPath, err)
		}
	}
}

// TestGetEntryRefAnnotationCarriesDefOriginal: def NÃO traduzida (sem
// mods/, pristine só em data/) colapsa a twin em ref e a anotação de ref
// precisa carregar o Original da def — é o texto do tooltip "Repetição de:".
// Regressão: o HashOrder cru (normalizeCollection) só merge Original de
// entradas COM mods/, então a anotação da def não traduzida vinha vazia e o
// tooltip saía truncado ("Repetição de: " sem texto).
func TestGetEntryRefAnnotationCarriesDefOriginal(t *testing.T) {
	svc := seedDisplayRoot(t)
	seedEventsSynthetic(common.GameVersionFFX2, "zev001", []string{sharedOriginal})
	seedEventsSynthetic(common.GameVersionFFX2, "zev002", []string{sharedOriginal})

	seedEventsData(t, common.GameVersionFFX2, "zev001")
	seedEventsData(t, common.GameVersionFFX2, "zev002")
	// Def e twin sem mods/: nenhuma tradução — o caso comum do dedup.
	stripEventMods(t, common.GameVersionFFX2, "zev001")
	stripEventMods(t, common.GameVersionFFX2, "zev002")

	twin, err := svc.GetEntry(KindEvents, "zev002", common.GameVersionFFX2)
	if err != nil {
		t.Fatalf("get twin: %v", err)
	}
	if !displayRowIsRef(twin.Rows[0]) {
		t.Fatalf("twin não traduzida deveria colapsar sob a def: %q", twin.Rows[0].Text["us"])
	}
	link, ok := twin.Refs[dto.RowKey(twin.Rows[0])]
	if !ok {
		t.Fatal("row de ref sem anotação Refs")
	}
	if link.SourceID != "zev001" {
		t.Fatalf("anotação deveria apontar para a def zev001: %q", link.SourceID)
	}
	if link.Original != sharedOriginal {
		t.Fatalf("Original da anotação deveria carregar o pristine da def: %q != %q",
			link.Original, sharedOriginal)
	}
}

// TestGetEntryWithoutOriginalKeepsRawPointers: sem contraparte em data/
// a entrada degrada — sem Original e sem reescrita de ponteiro; o colapso
// cai no ramo por texto, sem ponteiros de origem mista.
func TestGetEntryWithoutOriginalKeepsRawPointers(t *testing.T) {
	svc := seedDisplayRoot(t)
	seedEventsSynthetic(common.GameVersionFFX2, "zev001", []string{sharedOriginal})
	seedEventsSynthetic(common.GameVersionFFX2, "zev002", []string{sharedOriginal})

	// GameFilesRoot é um temp dir: nenhum arquivo em data/ nem mods/.
	e, err := svc.GetEntry(KindEvents, "zev002", common.GameVersionFFX2)
	if err != nil {
		t.Fatalf("get entry: %v", err)
	}
	if e.Rows[0].Original != nil {
		t.Fatal("entrada sem contraparte em data/ não deveria ter Original")
	}
	wantRaw := hash.Sum64Hex(sharedOriginal)
	if got := e.Rows[0].Hash[common.DefaultLocalization]; got != wantRaw {
		t.Fatalf("ponteiro RAW deveria prevalecer: %q != %q", got, wantRaw)
	}
}
