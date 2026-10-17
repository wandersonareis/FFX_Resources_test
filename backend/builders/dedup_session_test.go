package builders_test

import (
	"testing"

	"ffxresources/backend/builders"
	"ffxresources/backend/common"
	"ffxresources/backend/dto"
	"ffxresources/backend/formatters/hash"
)

// rowLinked monta uma row no DOMÍNIO DISPLAY (já merged com o original):
// Hash[us] = hash do ORIGINAL, texto atual em Text. Original é a cópia do
// pristine — é ele que prova "não traduzida" no dedup de display.
func rowLinked(index int, origUS, translated string) dto.TextRow {
	text := map[string]string{common.DefaultLocalization: translated}
	h := hash.Sum64Hex(origUS)
	return dto.TextRow{
		Index:    index,
		Hash:     map[string]string{common.DefaultLocalization: h},
		Text:     text,
		Original: map[string]string{common.DefaultLocalization: origUS},
	}
}

func refOf(r dto.TextRow) (string, bool) {
	return hash.Strip(r.Text[common.DefaultLocalization])
}

// A régua de sessão é append-only: a 1ª carga de cada ponteiro é a def e o
// que já saiu não muda — Extends seguintes com o MESMO ponteiro colapsam.
func TestHashOrderExtendFirstClickIsDef(t *testing.T) {
	const orig = "Repetição que atravessa arquivos onClick"
	ho := builders.NewHashOrder(nil)

	baseA := ho.Extend("bika236", []dto.TextRow{rowLinked(0, orig, orig), rowLinked(1, "Único de 236", "Único de 236")})
	if baseA != 0 {
		t.Fatalf("base da 1ª entrada deve ser 0, veio %d", baseA)
	}
	baseB := ho.Extend("bika237", []dto.TextRow{rowLinked(0, orig, orig)})
	if baseB != 2 {
		t.Fatalf("base da 2ª entrada deve ser 2, veio %d", baseB)
	}
	if id, index, name, ok := ho.Resolve(hash.Sum64Hex(orig)); !ok || id != "bika236" || index != 0 || name != "" {
		t.Fatalf("def do ponteiro: %q/%d/%q ok=%v, esperava bika236/0", id, index, name, ok)
	}
}

// A view colapsa a repetição NÃO traduzida cuja def já está na sessão e
// ANOTA o link: texto atual da def + origem (arquivo, índice).
func TestDedupDisplaySessionAnnotatesLinkedRef(t *testing.T) {
	const orig = "Texto repetido salvo em arquivo do container"
	ho := builders.NewHashOrder(nil)
	ho.Extend("bika236", []dto.TextRow{rowLinked(0, orig, orig)})
	twinEntry := dto.FileEntry{Rows: []dto.TextRow{rowLinked(3, orig, orig)}}
	base := ho.Extend("bika237", twinEntry.Rows)

	view := builders.DedupDisplayDTO(twinEntry, base, ho)
	bare, isRef := refOf(view.Rows[0])
	if !isRef {
		t.Fatalf("cópia não colapsou: %q", view.Rows[0].Text[common.DefaultLocalization])
	}
	link, ok := view.Refs[dto.RowKey(view.Rows[0])]
	if !ok {
		t.Fatalf("ref sem anotação de link: %+v", view.Refs)
	}
	if link.Text != orig {
		t.Fatalf("texto do link = def RAW da sessão: %q", link.Text)
	}
	if link.Original != orig {
		t.Fatalf("pristine do link: %q", link.Original)
	}
	if link.SourceID != "bika236" || link.SourceIndex != 0 {
		t.Fatalf("origem do link: %q/%d, esperava bika236/0 (bare %q)", link.SourceID, link.SourceIndex, bare)
	}
}

// Def traduzida NUNCA atrai a gêmea: a cópia ainda não traduzida fica
// literal e em inglês — sem ref apontando para a tradução de outro binário.
// (Decisão: dedupe é só de original para original; tradução repetida fica
// repetida, e gêmea em inglês é exibida em inglês.)
func TestDedupDisplayTranslatedDefKeepsTwinLiteral(t *testing.T) {
	const orig = "Repetição idêntica que o jogo carrega como dupe"
	ho := builders.NewHashOrder(nil)
	def := rowLinked(0, orig, "Tradução da def") // def traduzida
	ho.Extend("bika236", []dto.TextRow{def})
	twin := rowLinked(1, orig, orig) // cópia ainda não traduzida
	entry := dto.FileEntry{Rows: []dto.TextRow{twin}}
	ho.Extend("bika237", []dto.TextRow{twin})
	base, _ := ho.Offset("bika237")
	view := builders.DedupDisplayDTO(entry, base, ho)
	if _, isRef := refOf(view.Rows[0]); isRef {
		t.Fatalf("cópia NÃO traduzida não pode colapsar contra def traduzida: %q",
			view.Rows[0].Text[common.DefaultLocalization])
	}
	if len(view.Refs) != 0 {
		t.Fatalf("sem ref para anotar: %+v", view.Refs)
	}
	if got := view.Rows[0].Text[common.DefaultLocalization]; got != orig {
		t.Fatalf("célula deveria seguir em inglês (estado real do binário): %q", got)
	}

	// Cópia JÁ traduzida (texto divergente do original) nunca colapsa.
	literal := rowLinked(2, orig, "Tradução própria da cópia")
	entry2 := dto.FileEntry{Rows: []dto.TextRow{literal}}
	base2 := ho.Extend("bika238", []dto.TextRow{literal})
	view2 := builders.DedupDisplayDTO(entry2, base2, ho)
	if _, isRef := refOf(view2.Rows[0]); isRef {
		t.Fatalf("tradução divergente não pode virar ref")
	}
}

// Replace (frescura ao re-clicar um arquivo): enquanto a def está pendente a
// cópia colapsa; quando o re-clica chega TRADUZIDO, a def sai de pendente e a
// cópia volta a ser célula própria (em inglês), sem ref.
func TestHashOrderReplaceTranslatedDefStopsCollapse(t *testing.T) {
	const orig = "Definicia carregada no primeiro clique da sessão"
	ho := builders.NewHashOrder(nil)
	ho.Extend("bika236", []dto.TextRow{rowLinked(0, orig, orig)})
	ho.Extend("bika237", []dto.TextRow{rowLinked(0, orig, orig)})

	h := hash.Sum64Hex(orig)
	if !ho.DefPending(h) {
		t.Fatal("def recém-registrada sem tradução deveria estar pendente")
	}

	// 236 foi traduzido fora da sessão (via data/): re-clique re-decodifica
	// e Replace re-registra a entrada. No domínio display o ponteiro é do
	// ORIGINAL e não muda — o que muda é o TEXTO da def, que sai de
	// pendente e com isso deixa de poder atrair a cópia.
	ho.Replace("bika236", []dto.TextRow{rowLinked(0, orig, "Tradução aplicada depois")})
	if ho.DefPending(h) {
		t.Fatal("def traduzida não pode mais estar pendente")
	}
	if id, index, _, ok := ho.Resolve(h); !ok || id != "bika236" || index != 0 {
		t.Fatalf("def do ponteiro: %q/%d ok=%v, esperava bika236/0", id, index, ok)
	}

	// A cópia (ainda em inglês) vira célula própria: sem ref, sem link.
	base, _ := ho.Offset("bika237")
	view := builders.DedupDisplayDTO(
		dto.FileEntry{Rows: []dto.TextRow{rowLinked(0, orig, orig)}},
		base, ho,
	)
	if _, isRef := refOf(view.Rows[0]); isRef {
		t.Fatalf("cópia não pode colapsar contra def traduzida: %q",
			view.Rows[0].Text[common.DefaultLocalization])
	}
	if len(view.Refs) != 0 {
		t.Fatalf("sem ref para anotar: %+v", view.Refs)
	}
	if got := view.Rows[0].Text[common.DefaultLocalization]; got != orig {
		t.Fatalf("célula deveria seguir em inglês: %q", got)
	}
}
