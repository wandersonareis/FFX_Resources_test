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

// A tabela é a UNIÃO dos dois lados: row que só existe na tradução fica sem
// Original e marcada MissingInOriginal; row que só existe no original entra
// ordenada, com Text vazio e marcada MissingInTranslated. O texto de um lado
// nunca vaza para o outro.
func TestWithOriginalMontaUniaoDasRows(t *testing.T) {
	ref := entryRef{kind: KindEvents, id: "zev001", version: common.GameVersionFFX2}
	current := dto.FileEntry{Rows: []dto.TextRow{
		rawRow(0, "Traduzida"),
		rawRow(2, "Só na tradução"),
	}}
	orig := dto.FileEntry{Rows: []dto.TextRow{
		dataRow(0, "Original"),
		dataRow(1, "Só no original"),
	}}

	out, diverged := withOriginal(ref, current, orig)

	if len(out.Rows) != 3 {
		t.Fatalf("união com %d rows, esperava 3: %+v", len(out.Rows), out.Rows)
	}
	// Ordenado por Index depois da união: 0 (casada), 1 (só original),
	// 2 (só tradução).
	if out.Rows[0].Index != 0 || out.Rows[1].Index != 1 || out.Rows[2].Index != 2 {
		t.Fatalf("ordem pós-união: %d, %d, %d", out.Rows[0].Index, out.Rows[1].Index, out.Rows[2].Index)
	}

	// Row casada: ponteiro = hash do original, sem marcação de órfã.
	casada := out.Rows[0]
	if casada.Original[common.DefaultLocalization] != "Original" {
		t.Fatalf("row casada sem Original: %v", casada.Original)
	}
	if casada.MissingInOriginal || casada.MissingInTranslated {
		t.Fatal("row casada marcada como órfã")
	}
	if casada.Hash[common.DefaultLocalization] != hash.Sum64Hex("Original") {
		t.Fatalf("ponteiro da row casada = %q, queria o do original", casada.Hash[common.DefaultLocalization])
	}

	// Só no original: Original preenchido, Traduzido vazio e marcado.
	somOriginal := out.Rows[1]
	if !somOriginal.MissingInTranslated {
		t.Fatal("row só no original não marcada como ausente da tradução")
	}
	if somOriginal.MissingInOriginal {
		t.Fatal("row só no original marcada como ausente do original")
	}
	if len(somOriginal.Text) != 0 {
		t.Fatalf("texto do original vazou para a coluna Traduzido: %v", somOriginal.Text)
	}
	if somOriginal.Original == nil || somOriginal.Original[common.DefaultLocalization] != "Só no original" {
		t.Fatalf("coluna Original da row órfã: %v", somOriginal.Original)
	}

	// Só na tradução: Original vazio e marcado; texto atual preservado.
	somTraducao := out.Rows[2]
	if !somTraducao.MissingInOriginal {
		t.Fatal("row só na tradução não marcada como ausente do original")
	}
	if somTraducao.MissingInTranslated {
		t.Fatal("row só na tradução marcada como ausente da tradução")
	}
	if somTraducao.Original != nil {
		t.Fatalf("coluna Original preenchida para row sem contraparte: %v", somTraducao.Original)
	}
	if somTraducao.Text[common.DefaultLocalization] != "Só na tradução" {
		t.Fatalf("texto atual perdido: %v", somTraducao.Text)
	}

	// O relatório de divergência continua descrevendo os dois lados.
	if len(diverged.onlyInData) != 1 || diverged.onlyInData[0].Index != 1 {
		t.Fatalf("só em data/: %+v", diverged.onlyInData)
	}
	if len(diverged.onlyInMods) != 1 || diverged.onlyInMods[0].Index != 2 {
		t.Fatalf("só em mods/: %+v", diverged.onlyInMods)
	}
	if diverged.dataRows != 2 || diverged.modsRows != 2 {
		t.Fatalf("contagens: data=%d mods=%d", diverged.dataRows, diverged.modsRows)
	}

	// Nada das entradas foi mutado.
	if len(current.Rows) != 2 || current.Rows[1].MissingInOriginal {
		t.Fatal("entradas mutadas pela união")
	}
}
