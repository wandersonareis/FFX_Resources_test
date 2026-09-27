package builders_test

import (
	"testing"

	"ffxresources/backend/builders"
	"ffxresources/backend/common"
	"ffxresources/backend/dto"
	"ffxresources/backend/formatters/hash"
)

// ResolveDedupRefs expande refs "$hash" contra as defs do próprio
// collection — o mecanismo de aplicação do lote vindo do view dedupado.
func TestResolveDedupRefsExpandsAgainstOwnDefs(t *testing.T) {
	const shared = "Texto compartilhado entre entradas"
	const translated = "Tradução do texto compartilhado"

	// def em aaa (editada), ref em bbb — como o frontend devolve o lote.
	// Na ref, o hash permanece o do texto ORIGINAL (ponteiro opaco).
	defRow := rowOf(0, shared, nil)
	defRow.Text[common.DefaultLocalization] = translated
	refRow := rowOf(3, shared, nil)
	refRow.Text[common.DefaultLocalization] = "$" + hash.Sum64Hex(shared)
	c := dto.Collection{
		"aaa": {Rows: []dto.TextRow{defRow}},
		"bbb": {Rows: []dto.TextRow{refRow}},
	}
	resolved := builders.ResolveDedupRefs(c)

	// A ref recebe o texto ATUAL da def (a tradução), não o original.
	if got := resolved["bbb"].Rows[0].Text[common.DefaultLocalization]; got != translated {
		t.Fatalf("ref não resolveu para a def atual: %q", got)
	}
	if got := resolved["aaa"].Rows[0].Text[common.DefaultLocalization]; got != translated {
		t.Fatalf("def alterada: %q", got)
	}

	// A entrada não é mutada (o hash da ref é o do texto original).
	if got := c["bbb"].Rows[0].Text[common.DefaultLocalization]; got != "$"+hash.Sum64Hex(shared) {
		t.Fatalf("entrada mutada: %q", got)
	}
}

// Ref órfã (def fora do collection) é identidade: mantém o valor atual
// (o texto vence o hash — nunca grava "$hash" literal no binário).
func TestResolveDedupRefsKeepsOrphanRefs(t *testing.T) {
	const shared = "Texto que só existe como ref aqui"
	refRow := rowOf(3, shared, nil)
	refRow.Text[common.DefaultLocalization] = "$" + hash.Sum64Hex(shared)
	c := dto.Collection{
		"bbb": {Rows: []dto.TextRow{refRow}},
	}
	resolved := builders.ResolveDedupRefs(c)

	if got := resolved["bbb"].Rows[0].Text[common.DefaultLocalization]; got != "$"+hash.Sum64Hex(shared) {
		t.Fatalf("ref órfã deveria permanecer intacta: %q", got)
	}
}
