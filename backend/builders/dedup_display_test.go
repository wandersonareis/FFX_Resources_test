package builders_test

import (
	"testing"

	"ffxresources/backend/builders"
	"ffxresources/backend/common"
	"ffxresources/backend/dto"
	"ffxresources/backend/formatters/hash"
)

// rawOf monta uma row RAW (ponteiro = hash do texto, sem Original).
func rawOf(index int, us string, other map[string]string) dto.TextRow {
	return rowOf(index, us, other)
}

// mergedOf monta uma row merged: Pointer = hash do ORIGINAL (regra
// withOriginal), Text = mods-first e Original = data/.
func mergedOf(index int, originalUS, textUS string, other map[string]string) dto.TextRow {
	text := make(map[string]string, len(other)+2)
	for lang, v := range other {
		text[lang] = v
	}
	text[common.DefaultLocalization] = textUS
	orig := make(map[string]string, len(text))
	for lang, v := range text {
		orig[lang] = v
	}
	orig[common.DefaultLocalization] = originalUS
	return dto.TextRow{
		Index:    index,
		Hash:     map[string]string{common.DefaultLocalization: hash.Sum64Hex(originalUS)},
		Text:     text,
		Original: orig,
	}
}

func usText(r dto.TextRow) string { return r.Text[common.DefaultLocalization] }

// isRef verifica o contrato do frontend: texto "$h" com h == hash próprio.
func isRef(r dto.TextRow) bool {
	bare, ok := hash.Strip(usText(r))
	return ok && bare == r.Hash[common.DefaultLocalization]
}

// Gêmeos NÃO traduzidos colapsam: ref aponta para o PONTEIRO do original.
func TestDedupDisplayCollapsesUntranslatedTwins(t *testing.T) {
	const shared = "Frase original compartilhada entre linhas"
	c := dto.Collection{
		"aaa": {Rows: []dto.TextRow{rawOf(0, shared, nil)}},
		"bbb": {Rows: []dto.TextRow{rawOf(3, shared, nil)}},
	}
	ho := builders.NewHashOrder(c)

	// bbb é twin não traduzida de aaa: colapsa no ponteiro do original.
	twin := mergedOf(3, shared, shared, nil)
	out := builders.DedupDisplayDTO(dto.FileEntry{Rows: []dto.TextRow{twin}}, 1, ho)
	if !isRef(out.Rows[0]) {
		t.Fatal("twin não traduzida deveria colapsar em ref")
	}
	if got := usText(out.Rows[0]); got != "$"+hash.Sum64Hex(shared) {
		t.Fatalf("ref deveria apontar para o original: %q", got)
	}
}

// Célula TXA traduzida NÃO colapsa — dupe sobre texto traduzido é proibido,
// mesmo sob o MESMO ponteiro de original: é carga real de revisão.
func TestDedupDisplayNeverCollapsesTranslatedRows(t *testing.T) {
	const shared = "Frase original compartilhada entre linhas"
	const translated = "Tradução diferente desta twin"
	c := dto.Collection{
		"aaa": {Rows: []dto.TextRow{rawOf(0, shared, nil)}},
		"bbb": {Rows: []dto.TextRow{rawOf(3, shared, nil)}},
	}
	ho := builders.NewHashOrder(c)

	twin := mergedOf(3, shared, translated, nil)
	out := builders.DedupDisplayDTO(dto.FileEntry{Rows: []dto.TextRow{twin}}, 1, ho)
	if isRef(out.Rows[0]) {
		t.Fatal("célula traduzida deveria permanecer literal")
	}
	if got := usText(out.Rows[0]); got != translated {
		t.Fatalf("texto traduzido alterado: %q", got)
	}
}

// Def (1ª ocorrência) traduzida NÃO colapsa: permanece literal — e a twin
// não traduzida do MESMO original colapsa apontando para o ponteiro do
// ORIGINAL (não da tradução): o apply resolve defs[ponteiro] → tradução.
func TestDedupDisplayTranslatedDefCatchesUntranslatedTwin(t *testing.T) {
	const shared = "Frase original compartilhada entre linhas"
	const translated = "Tradução da frase compartilhada"
	c := dto.Collection{
		// RAW: def ainda não traduzida (ponteiro = hash do original).
		"aaa": {Rows: []dto.TextRow{rawOf(0, shared, nil)}},
		"bbb": {Rows: []dto.TextRow{rawOf(3, shared, nil)}},
	}
	ho := builders.NewHashOrder(c)

	//Estado atual: A tradução aplicada; a twin segue não traduzida. Ambas
	// saem do GetEntry com ponteiro reescrito para o original (withOriginal).
	def := mergedOf(0, shared, translated, nil)
	twin := mergedOf(3, shared, shared, nil)

	defOut := builders.DedupDisplayDTO(dto.FileEntry{Rows: []dto.TextRow{def}}, 0, ho)
	if isRef(defOut.Rows[0]) {
		t.Fatal("def nunca deveria colapsar em ref")
	}
	twinOut := builders.DedupDisplayDTO(dto.FileEntry{Rows: []dto.TextRow{twin}}, 1, ho)
	if !isRef(twinOut.Rows[0]) {
		t.Fatalf("twin não traduzida deveria colapsar sob a def traduzida: %q", usText(twinOut.Rows[0]))
	}

	// Payload do apply: def (literal traduzida, ponteiro do original) +
	// ref resolvem entre si — nenhuma mudança de contrato nos appliers.
	lote := dto.Collection{
		"aaa": {Rows: defOut.Rows},
		"bbb": {Rows: twinOut.Rows},
	}
	if err := hash.ValidateNoCollision(lote); err != nil {
		t.Fatalf("guarda de colisão: %v", err)
	}
}

// Tradução IDÊNTICA à def: ainda sob o mesmo ponteiro, é tradução
// propagada (textos iguais). Permanece literal — o texto já traduzido
// invalida o dupe de forma proposital.
func TestDedupDisplaySameTranslationStaysLiteral(t *testing.T) {
	const shared = "Frase original compartilhada entre linhas"
	const translated = "Tradução aplicada em ambas as twins"
	c := dto.Collection{
		"aaa": {Rows: []dto.TextRow{rawOf(0, shared, nil)}},
		"bbb": {Rows: []dto.TextRow{rawOf(3, shared, nil)}},
	}
	ho := builders.NewHashOrder(c)

	// Ambas traduzidas com o MESMO texto (propagação anterior): ponteiro
	// do original para as duas, texto ≠ original ⇒ nenhuma colapsa.
	a := mergedOf(0, shared, translated, nil)
	b := mergedOf(3, shared, translated, nil)
	out := builders.DedupDisplayDTO(dto.FileEntry{Rows: []dto.TextRow{a, b}}, 1, ho)
	if isRef(out.Rows[0]) || isRef(out.Rows[1]) {
		t.Fatal("células traduzidas deveriam permanecer literais")
	}
}

// Régua de runes é avaliada sobre o ORIGINAL: partícula curta nunca vira ref.
func TestDedupDisplayRespectsMinRunesOnOriginal(t *testing.T) {
	const shared = "ok" // < MinDedupRunes
	c := dto.Collection{
		"aaa": {Rows: []dto.TextRow{rawOf(0, shared, nil)}},
		"bbb": {Rows: []dto.TextRow{rawOf(1, shared, nil)}},
	}
	ho := builders.NewHashOrder(c)
	twin := mergedOf(1, shared, shared, nil)
	out := builders.DedupDisplayDTO(dto.FileEntry{Rows: []dto.TextRow{twin}}, 1, ho)
	if isRef(out.Rows[0]) {
		t.Fatal("original curto deveria ficar literal")
	}
}

// Sem Original na row (divergência de estrutura): modo RAW — colapso por
// texto, como antes da separação.
func TestDedupRawCollapsesByDefText(t *testing.T) {
	const shared = "Texto compartilhado entre entradas"
	c := dto.Collection{
		"aaa": {Rows: []dto.TextRow{rawOf(0, shared, nil), rawOf(1, shared, nil)}},
	}
	ho := builders.NewHashOrder(c)

	// aaa[0] = def (1ª ocorrência); aaa[1] = segunda ocorrência: ref por
	// texto igual ao da def.
	entry := dto.FileEntry{Rows: []dto.TextRow{rawOf(0, shared, nil), rawOf(1, shared, nil)}}
	out := builders.DedupDisplayDTO(entry, 0, ho)
	if !isRef(out.Rows[1]) {
		t.Fatalf("segunda ocorrência deveria colapsar: %q", usText(out.Rows[1]))
	}
	if isRef(out.Rows[0]) {
		t.Fatal("primeira ocorrência deveria ser a def")
	}
}

// Dupe de texto traduzido DIFERENTE em modo RAW: não colapsa.
func TestDedupRawNeverCollapsesDivergentText(t *testing.T) {
	const shared = "Texto compartilhado entre entradas"
	c := dto.Collection{
		"aaa": {Rows: []dto.TextRow{rawOf(0, shared, nil)}},
		"bbb": {Rows: []dto.TextRow{rawOf(1, shared, nil)}},
	}
	ho := builders.NewHashOrder(c)

	edited := rawOf(1, shared, nil)
	edited.Text[common.DefaultLocalization] = "Tradução divergente"
	out := builders.DedupDisplayDTO(dto.FileEntry{Rows: []dto.TextRow{edited}}, 1, ho)
	if isRef(out.Rows[0]) {
		t.Fatal("texto divergente não deveria colapsar")
	}
}

// HashOrder conta TODAS as rows (com ou sem texto) na ordem — a posição
// global tem que bater entre o RAW e a cópia merged.
func TestHashOrderOffsetsMatchAcrossDomains(t *testing.T) {
	const a = "Frase única da primeira entrada"
	const b = "Frase da segunda entrada"
	c := dto.Collection{
		"aaa": {Rows: []dto.TextRow{rawOf(0, a, nil), rawOf(1, "", nil)}},
		"bbb": {Rows: []dto.TextRow{rawOf(0, b, nil)}},
	}
	ho := builders.NewHashOrder(c)

	off, ok := ho.Offset("bbb")
	if !ok {
		t.Fatal("entrada bbb fora da ordem")
	}
	if off != 2 { // aaa tem 2 rows
		t.Fatalf("offset de bbb: %d", off)
	}
	first, ok := ho.RawFirst(hash.Sum64Hex(a))
	if !ok || first != 0 {
		t.Fatalf("1ª ocorrência de a: %d (ok=%v)", first, ok)
	}
}
