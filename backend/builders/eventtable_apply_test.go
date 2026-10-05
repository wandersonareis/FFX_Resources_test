package builders

import (
	"testing"

	"ffxresources/backend/common"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/dto"
	"ffxresources/backend/fileFormats/event"
	"ffxresources/backend/fileFormats/eventtable"
	"ffxresources/backend/formatters/hash"
)

// O motor de apply (4 fases) resolve a ref "$hash" do view dedupado contra
// a def do lote E propaga a edição por índice de texto para todas as
// cópias do ESCOPO — o que impede o literal "$abc…" de chegar ao binário
// e alinha os gêmeos entre arquivos.
func TestApplyTextDTOResolvesRefsAndPropagates(t *testing.T) {
	if err := ffxencoding.PrepareVersionCharsets(common.GameVersionFFX2); err != nil {
		t.Skipf("charsets: %v", err)
	}
	charset := ffxencoding.GetCharsetForLanguage(common.DefaultLocalization)

	const orig = "Texto de batalha repetido entre arquivos"
	const defText = "Tradução da def"
	bare := hash.Sum64Hex(orig)

	str := func(us string) *event.LocalizedFieldStringObject {
		fs := event.NewEmptyFieldString(charset, common.GameVersionFFX2)
		fs.SetRegularString(us)
		return event.NewLocalizedFieldStringObjectWithContent(common.DefaultLocalization, fs)
	}
	files := map[string]*eventtable.File{
		"bika00_00": {Kind: eventtable.KindBattleText, ID: "bika00_00", Version: common.GameVersionFFX2,
			Strings: []*event.LocalizedFieldStringObject{str(orig), str(orig)}},
		// A cópia em OUTRO arquivo (o def do lote alcança gêmeos fora do
		// lote pelo índice pré-estado).
		"bika00_01": {Kind: eventtable.KindBattleText, ID: "bika00_01", Version: common.GameVersionFFX2,
			Strings: []*event.LocalizedFieldStringObject{str(orig)}},
	}
	scope := TableApplyScope{
		Ids:        []string{"bika00_00", "bika00_01"},
		StringsFor: func(id string) []*event.LocalizedFieldStringObject { return files[id].Strings },
		Save:       func(id string) error { return nil },
	}

	// Payload do display: def literal editada + cópia do MESMO arquivo em
	// ref do MESMO ponteiro.
	defRow := dto.TextRow{
		Index: 0,
		Hash:  map[string]string{common.DefaultLocalization: bare},
		Text:  map[string]string{common.DefaultLocalization: defText},
	}
	refRow := dto.TextRow{
		Index: 1,
		Hash:  map[string]string{common.DefaultLocalization: bare},
		Text:  map[string]string{common.DefaultLocalization: hash.Prefix(bare)},
	}
	if _, isRef := refBare(refRow, hash.Prefix(bare)); !isRef {
		t.Fatalf("setup: cópia não casa como ref")
	}

	c := dto.Collection{"bika00_00": {Rows: []dto.TextRow{defRow, refRow}}}
	if err := applyTextDTO(eventtable.KindBattleText, c, nil, scope); err != nil {
		t.Fatalf("apply: %v", err)
	}
	for _, id := range []string{"bika00_00", "bika00_01"} {
		for i := range files[id].Strings {
			got := files[id].Strings[i].GetLocalizedContent(common.DefaultLocalization).GetRegularString()
			if got != defText {
				t.Fatalf("%s[%d]: %q, esperava a def propagada %q", id, i, got, defText)
			}
		}
	}
}

// Orphan ref: ref cuja def não está no lote nem no escopo não escreve
// literal e mantém a célula como está (identidade — o save não corrompe).
func TestApplyTextDTOOrphanRefKeepsCurrent(t *testing.T) {
	if err := ffxencoding.PrepareVersionCharsets(common.GameVersionFFX2); err != nil {
		t.Skipf("charsets: %v", err)
	}
	charset := ffxencoding.GetCharsetForLanguage(common.DefaultLocalization)
	const orig = "Texto repetido cujo def está em outro arquivo"
	bare := hash.Sum64Hex(orig)

	str := func(us string) *event.LocalizedFieldStringObject {
		fs := event.NewEmptyFieldString(charset, common.GameVersionFFX2)
		fs.SetRegularString(us)
		return event.NewLocalizedFieldStringObjectWithContent(common.DefaultLocalization, fs)
	}
	f := &eventtable.File{Kind: eventtable.KindBattleText, ID: "bika00_02", Version: common.GameVersionFFX2,
		Strings: []*event.LocalizedFieldStringObject{str(orig), str("Segmento vizinho intacto")}}
	scope := TableApplyScope{
		Ids:        []string{"bika00_02"},
		StringsFor: func(id string) []*event.LocalizedFieldStringObject { return f.Strings },
		Save:       func(id string) error { return nil },
	}
	refRow := dto.TextRow{
		Index: 0,
		Hash:  map[string]string{common.DefaultLocalization: bare},
		Text:  map[string]string{common.DefaultLocalization: hash.Prefix(bare)},
	}
	// Lote com def de OUTRO ponteiro: a ref do setup fica órfã.
	other := dto.TextRow{
		Index: 1,
		Hash:  map[string]string{common.DefaultLocalization: hash.Sum64Hex("Texto de outra row")},
		Text:  map[string]string{common.DefaultLocalization: "Texto de outra row"},
	}
	c := dto.Collection{"bika00_02": {Rows: []dto.TextRow{refRow, other}}}
	if err := applyTextDTO(eventtable.KindBattleText, c, nil, scope); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if got := f.Strings[0].GetLocalizedContent(common.DefaultLocalization).GetRegularString(); got != orig {
		t.Fatalf("ref órfã não pode tocar a célula: %q (esperava %q)", got, orig)
	}
}
