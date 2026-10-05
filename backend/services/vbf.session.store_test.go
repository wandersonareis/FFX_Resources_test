package services

import (
	"os"
	"path/filepath"
	"testing"

	"ffxresources/backend/builders"
	"ffxresources/backend/common"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/dto"
	"ffxresources/backend/fileFormats/event"
	"ffxresources/backend/fileFormats/eventtable"
	"ffxresources/backend/formatters/hash"
)

// O store de sessão é puro em memória: aqui não há container — os testes
// registram entradas montadas à mão (como o GetVbfTextEntry faria) e
// verificam a régua por clique. A gravação usa raiz temporária (o binário
// compilado vai para mods/ do temp, sem tocar fixtures).

func sessStr(t *testing.T, version common.GameVersion, us string) *event.LocalizedFieldStringObject {
	t.Helper()
	if err := ffxencoding.PrepareVersionCharsets(version); err != nil {
		t.Skipf("charsets: %v", err)
	}
	fs := event.NewEmptyFieldString(ffxencoding.GetCharsetForLanguage(common.DefaultLocalization), version)
	fs.SetRegularString(us)
	return event.NewLocalizedFieldStringObjectWithContent(common.DefaultLocalization, fs)
}

func sessRow(index int, origUS, translated string) dto.TextRow {
	text := map[string]string{common.DefaultLocalization: translated}
	return dto.TextRow{
		Index:    index,
		Hash:     map[string]string{common.DefaultLocalization: hash.Sum64Hex(origUS)},
		Text:     text,
		Original: map[string]string{common.DefaultLocalization: origUS},
	}
}

func sessEntry(id string, rows ...dto.TextRow) dto.FileEntry {
	return dto.FileEntry{Metadata: dto.Metadata{ID: id}, Rows: rows}
}

func isRefRow(r dto.TextRow) bool {
	bare, ok := hash.Strip(r.Text[common.DefaultLocalization])
	return ok && bare == r.Hash[common.DefaultLocalization]
}

// Clicou 236 primeiro, depois 235: a repetição não traduzida de 235 colapsa
// contra a SESSÃO e a view anota o link para a def. Re-clicar 235 —
// idempotente — devolve a MESMA view, sem re-decode.
func TestVbfSessionCrossFileCollapse(t *testing.T) {
	vbfSessions = map[string]*vbfSession{}
	t.Cleanup(vbfSessionResetAll)
	version := common.GameVersionFFX2

	const orig = "Frase de batalha compartilhada entre btl bins"
	// 1º clique: tudo literal (não há def na sessão).
	first := vbfSessionUpsert("fake.vbf", vbfTarget{Kind: KindBattleText, ID: "bika07_236", Version: version},
		sessEntry("bika07_236", sessRow(0, orig, orig), sessRow(1, "Único de 236", "Único de 236")),
		[]*event.LocalizedFieldStringObject{sessStr(t, version, orig), sessStr(t, version, "Único de 236")})
	if len(first.Refs) != 0 {
		t.Fatalf("1º clique não deveria ter refs: %+v", first.Refs)
	}

	// 2º clique: cópia (235) com a mesma row não traduzida → ref linkada.
	second := vbfSessionUpsert("fake.vbf", vbfTarget{Kind: KindBattleText, ID: "bika07_237", Version: version},
		sessEntry("bika07_237", sessRow(0, orig, orig)),
		[]*event.LocalizedFieldStringObject{sessStr(t, version, orig)})
	if !isRefRow(second.Rows[0]) {
		t.Fatalf("cópia deveria colapsar: %q", second.Rows[0].Text[common.DefaultLocalization])
	}
	link, ok := second.Refs[dto.RowKey(second.Rows[0])]
	if !ok {
		t.Fatalf("ref sem anotação: %+v", second.Refs)
	}
	if link.Text != orig || link.SourceID != "bika07_236" || link.SourceIndex != 0 {
		t.Fatalf("link: %+v", link)
	}

	// Re-clique em 237: cache da sessão — mesmo resultado, sem re-decode.
	cached, ok := vbfSessionCached("fake.vbf", vbfTarget{Kind: KindBattleText, ID: "bika07_237", Version: version})
	if !ok {
		t.Fatal("re-clique deveria servir do cache de sessão")
	}
	if cached.Rows[0].Text[common.DefaultLocalization] != second.Rows[0].Text[common.DefaultLocalization] {
		t.Fatalf("view de re-clique divergiu da inserida")
	}

	// Reset de sessão: tudo recomeça (o que foi colapsado volta ao literal).
	vbfSessionResetAll()
	if _, ok := vbfSessionCached("fake.vbf", vbfTarget{Kind: KindBattleText, ID: "bika07_237", Version: version}); ok {
		t.Fatal("reset não descartou a sessão")
	}
}

// Traduzido nunca colapsa na sessão: cada arquivo clicado mostra o que os
// próprios mods têm.
func TestVbfSessionTranslatedStaysLiteral(t *testing.T) {
	vbfSessions = map[string]*vbfSession{}
	t.Cleanup(vbfSessionResetAll)
	version := common.GameVersionFFX2

	const orig = "Repetição idêntica que o jogo carrega como dupe"
	vbfSessionUpsert("fake.vbf", vbfTarget{Kind: KindBattleText, ID: "bika07_400", Version: version},
		sessEntry("bika07_400", sessRow(0, orig, "Tradução do 400")),
		[]*event.LocalizedFieldStringObject{sessStr(t, version, "Tradução do 400")})
	view := vbfSessionUpsert("fake.vbf", vbfTarget{Kind: KindBattleText, ID: "bika07_401", Version: version},
		sessEntry("bika07_401", sessRow(0, orig, "Tradução do 401 — própria")),
		[]*event.LocalizedFieldStringObject{sessStr(t, version, "Tradução do 401 — própria")})
	if isRefRow(view.Rows[0]) {
		t.Fatalf("cópia traduzida não pode colapsar")
	}
}

// clearDedupViewCache invalida a sessão inteira: escrita em mods/ muda o
// texto que a sessão serviu e as defs de outros arquivos seguiriam velhas.
func TestVbfSessionInvalidatesOnWrite(t *testing.T) {
	vbfSessions = map[string]*vbfSession{}
	t.Cleanup(vbfSessionResetAll)
	version := common.GameVersionFFX2

	vbfSessionUpsert("fake.vbf", vbfTarget{Kind: KindBattleText, ID: "bika07_236", Version: version},
		sessEntry("bika07_236", sessRow(0, "Original qualquer de sessão", "Original qualquer de sessão")),
		[]*event.LocalizedFieldStringObject{sessStr(t, version, "Original qualquer de sessão")})
	clearDedupViewCache()
	if _, ok := vbfSessionCached("fake.vbf", vbfTarget{Kind: KindBattleText, ID: "bika07_236", Version: version}); ok {
		t.Fatal("clearDedupViewCache deveria descartar a sessão de .vbf")
	}
}

// O apply-from-vbf é o MESMO motor de data/ com o escopo da sessão: a
// edição do def se espalha pelas cópias abertas e cada binário tocado é
// gravado em mods/. Tradução divergente não se espalha.
func TestVbfSessionApplyPropagatesToOpenedCopies(t *testing.T) {
	seedEventsTempRoot(t)
	vbfSessions = map[string]*vbfSession{}
	t.Cleanup(vbfSessionResetAll)
	version := common.GameVersionFFX2

	const orig = "Frase replicada entre os btl abertos no clique"
	const target = "Tradução aplicada no def"
	vbfSessionUpsert("fake.vbf", vbfTarget{Kind: KindBattleText, ID: "bika07_236", Version: version},
		sessEntry("bika07_236", sessRow(0, orig, orig), sessRow(1, "Row própria do 236", "Row própria do 236")),
		[]*event.LocalizedFieldStringObject{sessStr(t, version, orig), sessStr(t, version, "Row própria do 236")})
	vbfSessionUpsert("fake.vbf", vbfTarget{Kind: KindBattleText, ID: "bika07_237", Version: version},
		sessEntry("bika07_237", sessRow(0, orig, orig)),
		[]*event.LocalizedFieldStringObject{sessStr(t, version, orig)})

	scope, ok := vbfSessionApplyScope("fake.vbf", KindBattleText, version)
	if !ok {
		t.Fatal("escopo da sessão indisponível")
	}
	bare := hash.Sum64Hex(orig)
	defRow := dto.TextRow{
		Index: 0,
		Hash:  map[string]string{common.DefaultLocalization: bare},
		Text:  map[string]string{common.DefaultLocalization: target},
	}
	refRow := dto.TextRow{
		Index: 0,
		Hash:  map[string]string{common.DefaultLocalization: bare},
		Text:  map[string]string{common.DefaultLocalization: hash.Prefix(bare)},
	}
	ownRow := dto.TextRow{
		Index: 1,
		Hash:  map[string]string{common.DefaultLocalization: hash.Sum64Hex("Row própria do 236")},
		Text:  map[string]string{common.DefaultLocalization: "Row própria do 236"},
	}
	// O lote: def editada (236) + ref própria dele; 237 só participa pela
	// propagação (não veio no lote).
	c := dto.Collection{"bika07_236": {Rows: []dto.TextRow{defRow, refRow, ownRow}}}
	if err := builders.ApplyTextDTOWithScope(KindBattleText, c, scope); err != nil {
		t.Fatalf("apply: %v", err)
	}

	// Binários dos 3 lados tocados: mods/ recebeu a tradução replicada.
	for _, id := range []string{"bika07_236", "bika07_237"} {
		rel, _ := eventtable.RelPathLoc(version, common.DefaultLocalization, KindBattleText, id)
		data, err := os.ReadFile(filepath.Join(common.GameFilesRoot, common.ModsFolder, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatalf("%s: mods/%s: %v", id, rel, err)
		}
		if len(data) == 0 {
			t.Fatalf("%s: arquivo de mods vazio", id)
		}
		// Round-trip do binário salvo: o texto 'us' tem que ser o do def.
		charset := ffxencoding.GetCharsetForLanguage(common.DefaultLocalization)
		fs, perr := event.FromFieldStringData(data, charset, version)
		if perr != nil {
			t.Fatalf("%s: reler mods/%s: %v", id, rel, perr)
		}
		if len(fs) == 0 {
			t.Fatalf("%s: sem strings em mods/%s", id, rel)
		}
		if got := fs[0].GetRegularString(); got != target {
			t.Fatalf("%s: mods/%s: %q, esperava %q", id, rel, got, target)
		}
	}
}
