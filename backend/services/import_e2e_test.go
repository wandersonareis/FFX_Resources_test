package services

import (
	"os"
	"path/filepath"
	"testing"

	"ffxresources/backend/common"
	"ffxresources/backend/dto"
	"ffxresources/backend/fileFormats/event"
	jsonfmt "ffxresources/backend/formatters/json"
	strfmt "ffxresources/backend/formatters/strings"
)

// Testes fim-a-fim do prepareImport com events sintéticos (sem fixtures de
// disco): artefato escrito em disco → PreviewImport/ImportFile.
//
// Cobre os três cenários pedidos:
//   - ordem dos textos trocada no arquivo,
//   - import parcial (menos textos que o total),
//   - hash não encontrado (store mudou desde o export) — log, não erro.
//
// O artefato carrega o hash do texto vigente no export (o tradutor mexe só
// no valor), então a store intacta casa tudo; editada, casa o que sobrou.
func importE2EPrepare(t *testing.T, raw []byte, name string) dto.ImportSummary {
	t.Helper()
	clearEventsForTest(t)

	seedEventsSynthetic(common.GameVersionFFX2, "zev001", []string{
		"Primeira frase do evento",
		"Segunda frase do evento",
		"Terceira frase do evento",
	})

	impPath := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(impPath, raw, 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
	summary, err := NewMetadataService(nil).PreviewImport(impPath, common.GameVersionFFX2)
	if err != nil {
		t.Fatalf("preview: %v", err)
	}
	return summary
}

// clearEventsForTest zera o store de events da versão (com cleanup).
func clearEventsForTest(t *testing.T) {
	t.Helper()
	event.ClearEvents(common.GameVersionFFX2)
	t.Cleanup(func() { event.ClearEvents(common.GameVersionFFX2) })
	if err := ensureVersionReady(common.GameVersionFFX2); err != nil {
		t.Fatalf("prepare version: %v", err)
	}
	clearDedupViewCache()
	t.Cleanup(clearDedupViewCache)
}

// collectionRaw devolve a store raw de events (mods-first) para montar o
// artefato com hash/texto consistentes antes da edição.
func collectionRawForTest(t *testing.T) dto.Collection {
	t.Helper()
	raw, err := NewMetadataService(nil).GetCollection(KindEvents, common.GameVersionFFX2, nil)
	if err != nil {
		t.Fatalf("collection: %v", err)
	}
	return raw
}

// TestImportE2ESwappedOrder: artefato com as rows em ordem trocada importa
// nas células certas (hash é a chave) e o resumo não tem erro nenhum.
func TestImportE2ESwappedOrder(t *testing.T) {
	seedEventsTempRoot(t)
	clearEventsForTest(t)
	seedEventsSynthetic(common.GameVersionFFX2, "zev001", []string{
		"Primeira frase do evento",
		"Segunda frase do evento",
		"Terceira frase do evento",
	})
	raw := collectionRawForTest(t)
	entry := raw["zev001"]
	if len(entry.Rows) != 3 {
		t.Fatalf("pré-condição: 3 rows, tem %d", len(entry.Rows))
	}

	// Edita as 3 (hash fica = texto do export) e embaralha a ordem.
	edited := make([]dto.TextRow, len(entry.Rows))
	copy(edited, entry.Rows)
	edited[0].Text[common.DefaultLocalization] = "Terceira traduzida"
	edited[1].Text[common.DefaultLocalization] = "Primeira traduzida"
	edited[2].Text[common.DefaultLocalization] = "Segunda traduzida"
	entry.Rows = []dto.TextRow{edited[2], edited[0], edited[1]}

	out, err := jsonfmt.NewJSONEventsFormatter().Marshal(dto.Collection{"zev001": entry})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	summary := importE2EPrepare(t, out, "swapped.json")
	if len(summary.Errors) != 0 {
		t.Fatalf("erros no resumo: %v", summary.Errors)
	}
	if summary.ChangedTexts != 3 {
		t.Fatalf("changed=%d, queria 3", summary.ChangedTexts)
	}
	if summary.Entries[0].Error != "" {
		t.Fatalf("entry com erro: %q", summary.Entries[0].Error)
	}
}

// TestImportE2EPartial: artefato com menos textos que o total (1 de 3).
func TestImportE2EPartial(t *testing.T) {
	seedEventsTempRoot(t)
	clearEventsForTest(t)
	seedEventsSynthetic(common.GameVersionFFX2, "zev001", []string{
		"Primeira frase do evento",
		"Segunda frase do evento",
		"Terceira frase do evento",
	})
	raw := collectionRawForTest(t)
	entry := raw["zev001"]

	// Só a row 2 vai no arquivo, traduzida; hash = texto do export.
	entry.Rows = []dto.TextRow{hEditRow(2, "", "Terceira frase do evento", "Terceira traduzida")}

	out, err := jsonfmt.NewJSONEventsFormatter().Marshal(dto.Collection{"zev001": entry})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	summary := importE2EPrepare(t, out, "partial.json")
	if len(summary.Errors) != 0 {
		t.Fatalf("erros no resumo: %v", summary.Errors)
	}
	if summary.ChangedTexts != 1 {
		t.Fatalf("changed=%d, queria 1 (import parcial)", summary.ChangedTexts)
	}
	if summary.Entries[0].Error != "" {
		t.Fatalf("entry com erro: %q", summary.Entries[0].Error)
	}
}

// TestImportE2EStaleHash: a store mudou desde o export — hash não casa e a
// row sai só no log; o resumo NÃO tem erro (o revisor corrige o artefato).
func TestImportE2EStaleHash(t *testing.T) {
	seedEventsTempRoot(t)
	clearEventsForTest(t)
	seedEventsSynthetic(common.GameVersionFFX2, "zev001", []string{
		"Primeira frase do evento",
		"Segunda frase do evento",
	})
	raw := collectionRawForTest(t)
	entry := raw["zev001"]

	// A store muda DEPOIS do export: texto da row 1 vira outra coisa.
	entry.Rows[0].Text[common.DefaultLocalization] = "Tradução anterior"
	entry.Rows[1].Text[common.DefaultLocalization] = "Segunda traduzida"
	// O artefato é exportado com ESTE estado...
	out, err := jsonfmt.NewJSONEventsFormatter().Marshal(dto.Collection{"zev001": entry})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	// ...e a store evolui: a frase "Tradução anterior" não existe mais.
	// A mutação é no FieldString do STORE (o DTO do GetCollection é cópia).
	ev := event.GetEvent(common.GameVersionFFX2, "zev001")
	if ev == nil {
		t.Fatal("pré-condição: zev001 no store")
	}
	ev.Strings[0].GetLocalizedContent(common.DefaultLocalization).SetRegularString("Tradução refeita")

	// O artefato (hash do texto antigo) não casa mais com a store nova:
	// 1 de 2 vai, e o resto é log.
	impPath := filepath.Join(t.TempDir(), "stale.json")
	if err := os.WriteFile(impPath, out, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	summary, err := NewMetadataService(nil).PreviewImport(impPath, common.GameVersionFFX2)
	if err != nil {
		t.Fatalf("preview: %v", err)
	}
	if len(summary.Errors) != 0 {
		t.Fatalf("hash não encontrado não pode ser erro do resumo: %v", summary.Errors)
	}
	if summary.ChangedTexts != 1 {
		t.Fatalf("changed=%d, queria 1 (só a row que ainda casa)", summary.ChangedTexts)
	}
}

// TestImportE2EUnknownEntry: único erro por entrada é "entrada não existe
// nesta versão" — e vem sozinho no resumo.
func TestImportE2EUnknownEntry(t *testing.T) {
	seedEventsTempRoot(t)
	clearEventsForTest(t)
	seedEventsSynthetic(common.GameVersionFFX2, "zev001", []string{"Frase qualquer aqui"})

	impPath := filepath.Join(t.TempDir(), "unknown.json")
	out, err := jsonfmt.NewJSONEventsFormatter().Marshal(dto.Collection{
		"zz_inexistente": dto.FileEntry{
			Metadata: dto.Metadata{Key: "ffx-2/event/zz/zzinexistente.bin", ID: "zz_inexistente"},
			Rows:     []dto.TextRow{hRow(0, "", "Texto solto")},
		},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(impPath, out, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	summary, err := NewMetadataService(nil).PreviewImport(impPath, common.GameVersionFFX2)
	if err != nil {
		t.Fatalf("preview: %v", err)
	}
	if len(summary.Errors) != 1 {
		t.Fatalf("errors=%v, queria 1", summary.Errors)
	}
	if got := summary.Errors[0]; !contains(got, "entrada não existe nesta versão") {
		t.Fatalf("erro %q, queria entrada inexistente", got)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// TestImportE2EStringsFormatRoundtrip: o mesmo fluxo parcial pelo formato
// .strings (dedup de valor + hash na chave da linha).
func TestImportE2EStringsFormatRoundtrip(t *testing.T) {
	seedEventsTempRoot(t)
	clearEventsForTest(t)
	seedEventsSynthetic(common.GameVersionFFX2, "zev001", []string{
		"Primeira frase do evento",
		"Segunda frase do evento",
	})
	raw := collectionRawForTest(t)
	entry := raw["zev001"]
	entry.Rows = []dto.TextRow{hEditRow(1, "", "Segunda frase do evento", "Segunda traduzida")}

	out, err := strfmt.Marshal(dto.Collection{"zev001": entry})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	summary := importE2EPrepare(t, out, "partial.strings")
	if len(summary.Errors) != 0 {
		t.Fatalf("erros no resumo: %v", summary.Errors)
	}
	if summary.ChangedTexts != 1 {
		t.Fatalf("changed=%d, queria 1", summary.ChangedTexts)
	}
}
