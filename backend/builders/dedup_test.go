package builders_test

import (
	"testing"

	"ffxresources/backend/builders"
	"ffxresources/backend/common"
	"ffxresources/backend/dto"
	"ffxresources/backend/formatters/hash"
)

// rowOf monta um TextRow com texto 'us' e hash coerente (a identidade do
// dedup é o hash do texto — byte-exato por XXH64).
func rowOf(index int, us string, other map[string]string) dto.TextRow {
	text := make(map[string]string, len(other)+1)
	for lang, t := range other {
		text[lang] = t
	}
	text[common.DefaultLocalization] = us
	return dto.TextRow{Index: index, Hash: hash.Texts(text), Text: text}
}

func usOf(r dto.TextRow) string { return r.Text[common.DefaultLocalization] }

func TestDedupDTOMarksRefsAfterFirstDef(t *testing.T) {
	const shared = "Texto compartilhado entre entradas"
	c := dto.Collection{
		"bbb": {Rows: []dto.TextRow{rowOf(0, shared, nil), rowOf(1, shared, nil)}},
		"aaa": {Rows: []dto.TextRow{rowOf(7, shared, nil)}},
	}
	dedup := builders.DedupDTO(c)

	// aaa vem antes na ordem de chave: a def fica lá, bbb recebe refs.
	if usOf(dedup["aaa"].Rows[0]) != shared {
		t.Fatalf("def alterada: %q", usOf(dedup["aaa"].Rows[0]))
	}
	wantRef := "$" + dedup["aaa"].Rows[0].Hash[common.DefaultLocalization]
	if got := usOf(dedup["bbb"].Rows[0]); got != wantRef {
		t.Fatalf("ref: got %q want %q", got, wantRef)
	}
	if got := usOf(dedup["bbb"].Rows[1]); got != wantRef {
		t.Fatalf("2ª ref do mesmo hash: got %q want %q", got, wantRef)
	}
}

func TestDedupDTORespectsMinRunesAndEmpty(t *testing.T) {
	const shared = "ok" // < MinDedupRunes: nunca vira ref
	c := dto.Collection{
		"aaa": {Rows: []dto.TextRow{rowOf(0, shared, nil)}},
		"bbb": {Rows: []dto.TextRow{rowOf(0, shared, nil)}},
		"ccc": {Rows: []dto.TextRow{rowOf(0, "", map[string]string{"jp": "テスト"})}},
	}
	dedup := builders.DedupDTO(c)

	if usOf(dedup["bbb"].Rows[0]) != shared {
		t.Fatalf("texto curto não deveria virar ref: %q", usOf(dedup["bbb"].Rows[0]))
	}
	// Row sem texto 'us' (só outros idiomas) passa intacta.
	if usOf(dedup["ccc"].Rows[0]) != "" {
		t.Fatalf("row sem us alterada: %q", usOf(dedup["ccc"].Rows[0]))
	}
	if dedup["ccc"].Rows[0].Text["jp"] != "テスト" {
		t.Fatalf("idioma ≠ us alterado: %q", dedup["ccc"].Rows[0].Text["jp"])
	}
}

func TestDedupDTODoesNotMutateInput(t *testing.T) {
	const shared = "Texto compartilhado entre entradas"
	c := dto.Collection{
		"aaa": {Rows: []dto.TextRow{rowOf(0, shared, nil)}},
		"bbb": {Rows: []dto.TextRow{rowOf(0, shared, nil)}},
	}
	builders.DedupDTO(c)

	if usOf(c["bbb"].Rows[0]) != shared {
		t.Fatalf("entrada mutada pelo dedup: %q", usOf(c["bbb"].Rows[0]))
	}
	// E o dedup é idempotente: refs de uma cópia permanecem refs.
	dedup := builders.DedupDTO(builders.DedupDTO(c))
	wantRef := "$" + c["aaa"].Rows[0].Hash[common.DefaultLocalization]
	if got := usOf(dedup["bbb"].Rows[0]); got != wantRef {
		t.Fatalf("dedup não idempotente: %q != %q", got, wantRef)
	}
}

