package xlsx

import (
	"strings"
	"testing"

	"ffxresources/backend/dto"

	"github.com/xuri/excelize/v2"
)

func sampleCollection() dto.Collection {
	c := dto.Collection{}
	one := dto.FileEntry{
		Metadata: dto.Metadata{
			Key: "ffx/event/obj_ps3/az/azit0000/azit0000.bin",
			ID:  "azit0000",
		},
		Rows: []dto.TextRow{
			{Index: 0, Text: map[string]string{"us": "Frase de teste comum ao arquivo"},
				Hash: map[string]string{"us": "aaaa0000aaaa0000"}},
			{Index: 1, Name: "name_simplified",
				Text: map[string]string{"us": "你好世界，这是一段很长的文字用来测试",
					jpKey: "テスト"},
				Hash: map[string]string{"us": "bbbb1111bbbb1111", jpKey: "cccc2222cccc2222"}},
		},
	}
	one.Metadata = one.Metadata.WithRowCount(len(one.Rows))
	two := dto.FileEntry{
		Metadata: dto.Metadata{
			Key: "ffx/menu/macrodic.dcp",
			ID:  "chunk_07",
		},
		Rows: []dto.TextRow{
			{Index: 3, Name: "name",
				Text: map[string]string{"us": strings.Repeat("texto longo ", 20)}, // > 80 → clamp
				Hash: map[string]string{"us": "dddd3333dddd3333"}},
		},
	}
	two.Metadata = two.Metadata.WithRowCount(len(two.Rows))
	c["azit0000"] = one
	c["chunk_07"] = two
	return c
}

const jpKey = "jp"

// As larguras saem CALCULADAS: todas começam em xlsxWidthBase (10) e
// crescem pelo maior texto observado (runes), teto 80 com clamp.
func TestMarshalLargurasCalculadas(t *testing.T) {
	raw, err := NewXLSXFormatter().Marshal(sampleCollection())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	f, err := excelize.OpenReader(strings.NewReader(string(raw)))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer f.Close()

	cases := []struct {
		col  string
		want float64
	}{
		{"E", float64(xlsxWidthBase)},                                     // lang: conteúdo curto → fica na base
		{"F", float64(len("aaaa0000aaaa0000"))},                           // hash: maior hex observado
		{"H", float64(xlsxWidthBase)},                                     // divergent: "false" < base
		{"A", float64(len("ffx/event/obj_ps3/az/azit0000/azit0000.bin"))}, // key: maior key
		{"B", float64(xlsxWidthBase)},                                     // id: "chunk_07" (8) < base
		{"C", float64(len("name_simplified"))},                            // name: maior name
		{"D", float64(xlsxWidthBase)},                                     // index: dígitos < base
		{"G", 80},                                                         // text: estourou o teto → clamp
	}
	for _, tc := range cases {
		got, err := f.GetColWidth("Sheet1", tc.col)
		if err != nil {
			t.Fatalf("GetColWidth(%s): %v", tc.col, err)
		}
		if got != tc.want {
			t.Errorf("coluna %s: largura %v, esperado %v", tc.col, got, tc.want)
		}
	}
}

// A coluna text tem WrapText em range: a linha quebra e a altura da linha
// é automática ao abrir (nunca SetRowHeight).
func TestMarshalWrapNaColunaText(t *testing.T) {
	raw, err := NewXLSXFormatter().Marshal(sampleCollection())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	f, err := excelize.OpenReader(strings.NewReader(string(raw)))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer f.Close()

	styleID, err := f.GetCellStyle("Sheet1", "G2")
	if err != nil {
		t.Fatalf("GetCellStyle: %v", err)
	}
	style, err := f.GetStyle(styleID)
	if err != nil {
		t.Fatalf("GetStyle: %v", err)
	}
	if style.Alignment == nil || !style.Alignment.WrapText {
		t.Error("célula G2 deveria ter WrapText (queca de linha)")
	}
}

// Proteção das colunas: metadata (A–F) e header travados, text (G) e
// divergent (H) destravados — planilha protegida SEM senha (quem
// insistir desprotege em um clique) e validação A:F carregando o aviso
// "Edição não permitida".
func TestMarshalProtegeColunasMetadata(t *testing.T) {
	marshallers := []struct {
		nome  string
		marsh func(dto.Collection) ([]byte, error)
	}{
		{"xlsx", NewXLSXFormatter().Marshal},
		{"xlsx-no-dedup", XLSXFormatterNoDedup{}.Marshal},
	}

	for _, m := range marshallers {
		t.Run(m.nome, func(t *testing.T) {
			raw, err := m.marsh(sampleCollection())
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			f, err := excelize.OpenReader(strings.NewReader(string(raw)))
			if err != nil {
				t.Fatalf("open: %v", err)
			}
			defer f.Close()

			cellStyle := func(cell string) *excelize.Style {
				t.Helper()
				styleID, err := f.GetCellStyle("Sheet1", cell)
				if err != nil {
					t.Fatalf("GetCellStyle(%s): %v", cell, err)
				}
				style, err := f.GetStyle(styleID)
				if err != nil {
					t.Fatalf("GetStyle(%s): %v", cell, err)
				}
				return style
			}

			// Metadata (A–F) e header (linha 1): travadas. Protection
			// nil é o PADRÃO do xlsx (célula travada).
			for _, cell := range []string{"A1", "G1", "H1", "A2", "C3", "F2"} {
				if style := cellStyle(cell); style.Protection != nil && !style.Protection.Locked {
					t.Errorf("%s (metadata/header) deveria estar travada", cell)
				}
			}

			// text (G) e divergent (H): destravadas (Locked false
			// explícito no estilo).
			for _, cell := range []string{"G2", "H2", "G5", "H5"} {
				if style := cellStyle(cell); style.Protection == nil || style.Protection.Locked {
					t.Errorf("%s (text/divergent) deveria estar destravada", cell)
				}
			}

			// O Protection entrou NO textStyle: WrapText preservado.
			if style := cellStyle("G2"); style.Alignment == nil || !style.Alignment.WrapText {
				t.Error("G2 perdeu o WrapText ao ganhar Protection")
			}

			// Proteção da planilha: seleção/cópia e AutoFilter
			// liberados; estrutura (inserir/deletar/formatar) segue
			// bloqueada.
			prot, err := f.GetSheetProtection("Sheet1")
			if err != nil {
				t.Fatalf("GetSheetProtection: %v", err)
			}
			if !prot.SelectLockedCells || !prot.SelectUnlockedCells || !prot.AutoFilter {
				t.Errorf("esperava seleção+cópia e autofilter liberados, got %+v", prot)
			}
			if prot.InsertRows || prot.DeleteRows || prot.FormatCells {
				t.Errorf("estrutura deveria seguir bloqueada, got %+v", prot)
			}

			// Validação A:F: regra custom sempre-inválida com o aviso
			// customizado em estilo Stop.
			dvs, err := f.GetDataValidations("Sheet1")
			if err != nil {
				t.Fatalf("GetDataValidations: %v", err)
			}
			if len(dvs) != 1 {
				t.Fatalf("validações: %d, esperado 1", len(dvs))
			}
			dv := dvs[0]
			if dv.Sqref != "A:F" {
				t.Errorf("sqref %q, esperado \"A:F\"", dv.Sqref)
			}
			if dv.Type != "custom" || dv.Formula1 != "FALSE" {
				t.Errorf("regra custom sempre-inválida, got type=%q formula1=%q", dv.Type, dv.Formula1)
			}
			if !dv.ShowErrorMessage {
				t.Error("validação deveria mostrar a mensagem de erro")
			}
			if dv.ErrorStyle == nil || *dv.ErrorStyle != "stop" {
				t.Errorf("errorStyle %v, esperado \"stop\"", dv.ErrorStyle)
			}
			if dv.ErrorTitle == nil || *dv.ErrorTitle != metadataEditErrorTitle {
				t.Errorf("título %v, esperado %q", dv.ErrorTitle, metadataEditErrorTitle)
			}
			if dv.Error == nil || *dv.Error != metadataEditErrorMsg {
				t.Errorf("mensagem %v, esperado %q", dv.Error, metadataEditErrorMsg)
			}
		})
	}
}

// Round-trip bulk: 2 entradas (events + macro) voltam com keys, ids e
// textos preservados — a coluna key é a identidade da entrada.
func TestUnmarshalRoundTripBulk(t *testing.T) {
	c := sampleCollection()
	raw, err := NewXLSXFormatter().Marshal(c)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got, err := NewXLSXFormatter().Unmarshal(raw)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got) != len(c) {
		t.Fatalf("entradas: %d, esperado %d", len(got), len(c))
	}
	for id, want := range c {
		entry, ok := got[id]
		if !ok {
			t.Fatalf("entrada %s sumiu", id)
		}
		if entry.Metadata.Key != want.Metadata.Key {
			t.Errorf("%s: key %q != %q", id, entry.Metadata.Key, want.Metadata.Key)
		}
		if len(entry.Rows) != len(want.Rows) {
			t.Fatalf("%s: rows %d != %d", id, len(entry.Rows), len(want.Rows))
		}
		for i, row := range want.Rows {
			g := entry.Rows[i]
			if g.Index != row.Index || g.Name != row.Name {
				t.Errorf("%s row %d: (%d,%q) != (%d,%q)", id, i, g.Index, g.Name, row.Index, row.Name)
			}
			for lang, text := range row.Text {
				if g.Text[lang] != text {
					t.Errorf("%s row %d %s: texto %q != %q", id, i, lang, g.Text[lang], text)
				}
			}
		}
	}
}
