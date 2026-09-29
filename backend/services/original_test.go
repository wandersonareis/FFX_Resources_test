package services

import (
	"testing"

	"ffxresources/backend/common"
	"ffxresources/backend/dto"
	"ffxresources/backend/formatters/hash"
)

// dataRow monta uma row pristine (como o leitor de data/ entrega).
func dataRow(index int, us string) dto.TextRow {
	text := map[string]string{common.DefaultLocalization: us}
	return dto.TextRow{
		Index: index,
		Hash:  map[string]string{common.DefaultLocalization: hash.Sum64Hex(us)},
		Text:  text,
	}
}

// rawRow monta uma row cru (mods-first: ponteiro = hash do texto atual).
func rawRow(index int, textUS string) dto.TextRow {
	text := map[string]string{common.DefaultLocalization: textUS}
	return dto.TextRow{
		Index: index,
		Hash:  map[string]string{common.DefaultLocalization: hash.Sum64Hex(textUS)},
		Text:  text,
	}
}

// O ponteiro de exibição é o hash do ORIGINAL: def traduzida aponta para o
// original; twin não traduzida é no-op (hashes já coincidem)
// ⇒ par def/twin compartilha o ponteiro e o dedup de display os agrupa.
func TestWithOriginalRewritesPointerToOriginalHash(t *testing.T) {
	const shared = "Frase original compartilhada" // ≥ MinDedupRunes
	const translation = "Tradução da frase"

	current := dto.FileEntry{
		Rows: []dto.TextRow{rawRow(0, translation), rawRow(1, shared)},
	}
	orig := dto.FileEntry{Rows: []dto.TextRow{dataRow(0, shared), dataRow(1, shared)}}

	ref := entryRef{kind: KindEvents, id: "zev001", version: common.GameVersionFFX2}
	out, diverged := withOriginal(ref, current, orig)

	wantPointer := hash.Sum64Hex(shared)
	if got := out.Rows[0].Hash[common.DefaultLocalization]; got != wantPointer {
		t.Fatalf("def traduzida: hash deveria ser o do ORIGINAL: %q != %q", got, wantPointer)
	}
	if got := out.Rows[1].Hash[common.DefaultLocalization]; got != wantPointer {
		t.Fatalf("twin não traduzida: hash deveria ser o do ORIGINAL: %q != %q", got, wantPointer)
	}

	// Ponteiro do original ≠ hash do texto traduzido: é o flag "traduzido".
	if out.Rows[0].Hash[common.DefaultLocalization] == hash.Sum64Hex(translation) {
		t.Fatal("ponteiro da def não deveria ser o hash da tradução")
	}

	if got := out.Rows[0].Original[common.DefaultLocalization]; got != shared {
		t.Fatalf("Original ausente na def: %q", got)
	}
	if got := out.Rows[0].Text[common.DefaultLocalization]; got != translation {
		t.Fatalf("Texto traduzido alterado: %q", got)
	}
	if !diverged.empty() {
		t.Fatal("merge limpo não deveria reportar divergência")
	}
}

// Row sem contraparte em data/ degrada POR ROW: sem Original e sem
// reescrita de ponteiro — e o relatório identifica a direção da divergência.
func TestWithOriginalReportsDivergence(t *testing.T) {
	ref := entryRef{kind: KindEvents, id: "zev001", version: common.GameVersionFFX2}
	current := dto.FileEntry{Rows: []dto.TextRow{rawRow(0, "Tradução")}}
	// data/ divergiu: só tem a row 1, ausente no estado atual.
	orig := dto.FileEntry{Rows: []dto.TextRow{dataRow(1, "Outra frase")}}

	out, diverged := withOriginal(ref, current, orig)

	if out.Rows[0].Original != nil {
		t.Fatal("merge parcial não deveria anexar Original em row sem contraparte")
	}
	wantText := hash.Sum64Hex("Tradução")
	if got := out.Rows[0].Hash[common.DefaultLocalization]; got != wantText {
		t.Fatalf("hash deveria manter o ponteiro do texto do arquivo: %q != %q", got, wantText)
	}

	// Relatório: 1 row só em data/ (Index 1), 1 row só no estado atual.
	if diverged.empty() {
		t.Fatal("divergência deveria ser reportada")
	}
	if len(diverged.onlyInData) != 1 || diverged.onlyInData[0].Index != 1 {
		t.Fatalf("só em data/: %+v", diverged.onlyInData)
	}
	if len(diverged.onlyInMods) != 1 || diverged.onlyInMods[0].Index != 0 {
		t.Fatalf("só em mods/: %+v", diverged.onlyInMods)
	}
	if diverged.dataRows != 1 || diverged.modsRows != 1 {
		t.Fatalf("contagens: data=%d mods=%d", diverged.dataRows, diverged.modsRows)
	}
}

// rows de entrada não são mutadas: withOriginal copia o slice antes de
// escrever Original/Hash (o cache do cru fica intacto).
func TestWithOriginalDoesNotMutateInputRows(t *testing.T) {
	const shared = "Frase original compartilhada"
	row := rawRow(0, shared)
	entry := dto.FileEntry{Rows: []dto.TextRow{row}}
	orig := dto.FileEntry{Rows: []dto.TextRow{dataRow(0, shared)}}

	withOriginal(entryRef{kind: KindEvents, id: "zev001", version: common.GameVersionFFX2}, entry, orig)

	if row.Original != nil {
		t.Fatal("row de entrada mutada pelo merge")
	}
	if row.Hash[common.DefaultLocalization] != hash.Sum64Hex(shared) {
		t.Fatalf("hash de entrada mutado: %q", row.Hash[common.DefaultLocalization])
	}
}
