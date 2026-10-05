package services

import (
	"os"
	"testing"

	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/event"
	jsonfmt "ffxresources/backend/formatters/json"
)

// TestExportEntryFiltersBlankRows: o export de ENTRADA ÚNICA (menu de
// contexto → ExportEntry) obedece ao mesmo contrato do export em lote —
// rows sem 'us' utilizável (vazio, só espaço ou "-") não chegam ao
// artefato, em JSON nem em .strings.
func TestExportEntryFiltersBlankRows(t *testing.T) {
	seedEventsTempRoot(t)
	event.ClearEvents(common.GameVersionFFX2)
	defer event.ClearEvents(common.GameVersionFFX2)
	if err := ensureVersionReady(common.GameVersionFFX2); err != nil {
		t.Fatalf("prepare version: %v", err)
	}
	clearDedupViewCache()
	t.Cleanup(clearDedupViewCache)

	seedEventsSynthetic(common.GameVersionFFX2, "zev001", []string{
		"Primeira linha", // única traduzível
		"-",              // texto próprio do jogo: sai
		"  ",             // só espaço: sai
		"",               // vazio: sai
	})

	paths, err := NewMetadataService(nil).ExportEntry(KindEvents, common.GameVersionFFX2, "zev001", nil)
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if len(paths) != 2 {
		t.Fatalf("paths=%v, queria JSON + .strings", paths)
	}
	for _, p := range paths {
		if _, statErr := os.Stat(p); statErr != nil {
			t.Fatalf("artefato ausente: %s (%v)", p, statErr)
		}
	}

	// O JSON volta SEM a row em branco e com a contagem recalculada.
	out, err := jsonfmt.NewJSONEventsFormatter().ReadEvents(paths[0])
	if err != nil {
		t.Fatalf("releitura: %v", err)
	}
	entry, ok := out["zev001"]
	if !ok {
		t.Fatal("entrada zev001 ausente no artefato")
	}
	if n := len(entry.Rows); n != 1 {
		for _, r := range entry.Rows {
			t.Logf("row %+v", r)
		}
		t.Fatalf("rows=%d, queria só a traduzível", n)
	}
	if got := entry.Rows[0].Text[common.DefaultLocalization]; got != "Primeira linha" {
		t.Fatalf("texto %q, queria %q", got, "Primeira linha")
	}
	if entry.Metadata.RowCount != 1 {
		t.Fatalf("row_count=%d, queria 1", entry.Metadata.RowCount)
	}
}

// TestExportEntryDropsFullyBlankEntry: entrada só com rows em branco não
// gera artefato — é omitida do lote inteiro (não escreve arquivo vazio).
func TestExportEntryDropsFullyBlankEntry(t *testing.T) {
	seedEventsTempRoot(t)
	event.ClearEvents(common.GameVersionFFX2)
	defer event.ClearEvents(common.GameVersionFFX2)
	if err := ensureVersionReady(common.GameVersionFFX2); err != nil {
		t.Fatalf("prepare version: %v", err)
	}
	clearDedupViewCache()
	t.Cleanup(clearDedupViewCache)

	seedEventsSynthetic(common.GameVersionFFX2, "zev001", []string{"-", "  "})

	_, err := NewMetadataService(nil).ExportEntry(KindEvents, common.GameVersionFFX2, "zev001", nil)
	if err == nil {
		t.Fatal("entrada só com rows em branco deveria falhar (nada a exportar)")
	}
}

// Guarda de contrato: isBlankText é o mesmo predicado do casamento de
// import — vazio, só espaço (e tab) ou o hífen do jogo.
func TestIsBlankTextContract(t *testing.T) {
	for _, c := range []struct {
		in   string
		want bool
	}{
		{"", true},
		{"   ", true},
		{"\t", true},
		{"-", true},
		{" - ", true},
	} {
		if got := isBlankText(c.in); got != c.want {
			t.Fatalf("isBlankText(%q)=%v, queria %v", c.in, got, c.want)
		}
	}
	if isBlankText("- ou não") {
		t.Fatal("texto com hífen e conteúdo não é blank")
	}
}
