package strings_test

import (
	"strings"
	"testing"

	"ffxresources/backend/common"
	"ffxresources/backend/dto"
	"ffxresources/backend/formatters/hash"
	fmtstrings "ffxresources/backend/formatters/strings"
)

func sampleCollection() dto.Collection {
	return dto.Collection{
		"azit0000": {
			Metadata: dto.NewEventMetadata("azit0000", common.GameVersionFFX).WithRowCount(2),
			Rows: []dto.TextRow{
				{
					Index: 0,
					Hash:  hash.Texts(map[string]string{"us": "Where are you?", "sp": "¿Dónde estás?"}),
					Text:  map[string]string{"us": "Where are you?", "sp": "¿Dónde estás?"},
				},
				{
					Index: 1,
					Hash:  hash.Texts(map[string]string{"us": "Where are you?"}),
					Text:  map[string]string{"us": "Where are you?"},
				},
			},
		},
		"command": {
			Metadata: dto.NewObjectMetadata(common.GameVersionFFX, "battle/kernel", "command.bin", "ffx/battle/kernel/command.bin").WithRowCount(2),
			Rows: []dto.TextRow{
				{
					Index: 0, Name: "name",
					Hash: hash.Texts(map[string]string{"us": "Potion"}),
					Text: map[string]string{"us": "Potion"},
				},
				{
					Index: 0, Name: "description",
					Hash: hash.Texts(map[string]string{"us": "Restores HP"}),
					Text: map[string]string{"us": "Restores HP"},
				},
			},
		},
		"chunk_00": {
			Metadata: dto.NewMacroMetadata(common.GameVersionFFX, 0).WithRowCount(1),
			Rows: []dto.TextRow{
				{
					Index: 0, Name: "name",
					Hash: hash.Texts(map[string]string{"us": "Tom & Jerry"}),
					Text: map[string]string{"us": "Tom & Jerry"},
				},
			},
		},
	}
}

func TestGoldenExampleLines(t *testing.T) {
	// Linhas no estilo do exemplo do usuário: parser deve aceitar
	// def, ref e chaves das 3 famílias (hash mismatch aqui só loga).
	raw := "/*key=ffx/event/obj_ps3/az/azit0000/azit0000.bin row_count=309*/\n" +
		"ffx:azit0000:0:us║$ba7b73bfc15bb6ea = {PC:04:WAKKA}{TEXT_NEWLINE}{PC:01:YUNA}! Where are you?\n" +
		"ffx:azit0000:1:us║$ba7b73bfc15bb6ea = $ba7b73bfc15bb6ea\n" +
		"ffx:azit0042:17:us║$ba7b73bfc15bb6ea = $ba7b73bfc15bb6ea\n"
	c, err := fmtstrings.Unmarshal([]byte(raw))
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	ev, ok := c["azit0000"]
	if !ok || len(ev.Rows) != 2 {
		t.Fatalf("azit0000 rows: %+v", c.SortedKeys())
	}
	if ev.Rows[0].Index != 0 || ev.Rows[1].Index != 1 {
		t.Fatalf("indexes: %+v", ev.Rows)
	}
	if ev.Metadata.Key != "ffx/event/obj_ps3/az/azit0000/azit0000.bin" {
		t.Fatalf("key: %+v", ev.Metadata)
	}
	other, ok := c["azit0042"]
	if !ok || len(other.Rows) != 1 || other.Rows[0].Index != 17 {
		t.Fatalf("azit0042: %+v", c.SortedKeys())
	}
}

func TestMarshalShapesOutput(t *testing.T) {
	raw, err := fmtstrings.Marshal(sampleCollection())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	out := string(raw)
	for _, want := range []string{
		"/*key=ffx/event/obj_ps3/az/azit0000/azit0000.bin row_count=2*/",
		"/*key=ffx/battle/kernel/command.bin row_count=2*/",
		"/*key=ffx/menu/macrodic.dcp row_count=1*/",
		"ffx:azit0000:0:us║$",
		"ffx:azit0000:0:sp║$",
		"ffx:command:name:0:us║$",
		"ffx:command:description:0:us║$",
		"ffx:chunk_00:name:0:us║$",
		"Tom & Jerry",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q:\n%s", want, out)
		}
	}
	// "Where are you?" repete no default lang (>=5 runes) → segunda vira ref.
	h := hash.Sum64Hex("Where are you?")
	if !strings.Contains(out, "= $"+h+"\n") {
		t.Fatalf("expected dedup ref for repeated text:\n%s", out)
	}
}

func TestMarshalUnmarshalRoundTrip(t *testing.T) {
	in := sampleCollection()
	raw, err := fmtstrings.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	back, err := fmtstrings.Unmarshal(raw)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(back) != len(in) {
		t.Fatalf("entry count: %d vs %d", len(back), len(in))
	}
	for _, k := range in.SortedKeys() {
		want, got := in[k], back[k]
		if got.Metadata.Key != want.Metadata.Key || got.Metadata.ID != want.Metadata.ID {
			t.Fatalf("%s metadata: %+v vs %+v", k, got.Metadata, want.Metadata)
		}
		if len(got.Rows) != len(want.Rows) {
			t.Fatalf("%s rows: %d vs %d", k, len(got.Rows), len(want.Rows))
		}
		wantRows := append([]dto.TextRow(nil), want.Rows...)
		dto.SortRows(wantRows)
		for i := range wantRows {
			wr, gr := wantRows[i], got.Rows[i]
			if wr.Index != gr.Index || wr.Name != gr.Name {
				t.Fatalf("%s row %d position: %+v vs %+v", k, i, gr, wr)
			}
			for lang, text := range wr.Text {
				if gr.Text[lang] != text {
					t.Fatalf("%s row %d lang %s: %q vs %q", k, i, lang, gr.Text[lang], text)
				}
				if gr.Hash[lang] != wr.Hash[lang] {
					t.Fatalf("%s row %d lang %s hash: %q vs %q", k, i, lang, gr.Hash[lang], wr.Hash[lang])
				}
			}
		}
	}
}

func TestMarshalLangsFilter(t *testing.T) {
	raw, err := fmtstrings.MarshalLangs(sampleCollection(), []string{"us"})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	out := string(raw)
	if strings.Contains(out, ":sp║$") {
		t.Fatalf("sp lines must be filtered:\n%s", out)
	}
	back, err := fmtstrings.Unmarshal(raw)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := back["azit0000"].Rows[0].Text["sp"]; ok {
		t.Fatalf("sp must not roundtrip: %+v", back["azit0000"].Rows[0])
	}
	if back["azit0000"].Rows[0].Text["us"] != "Where are you?" {
		t.Fatalf("us lost: %+v", back["azit0000"].Rows[0])
	}
}

func TestEscapesRoundTrip(t *testing.T) {
	tricky := "a=b\\c║d\nnewline here and  leading kept"
	c := dto.Collection{
		"ev001": {
			Metadata: dto.NewEventMetadata("ev001", common.GameVersionFFX).WithRowCount(1),
			Rows: []dto.TextRow{
				{Index: 0, Hash: hash.Texts(map[string]string{"us": tricky}), Text: map[string]string{"us": tricky}},
			},
		},
	}
	raw, err := fmtstrings.Marshal(c)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(raw), "║d\n") {
		t.Fatalf("raw separator/newline leaked:\n%s", raw)
	}
	back, err := fmtstrings.Unmarshal(raw)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := back["ev001"].Rows[0].Text["us"]; got != tricky {
		t.Fatalf("escape roundtrip: %q vs %q", got, tricky)
	}
}

func TestEmptyCellsSkipped(t *testing.T) {
	c := dto.Collection{
		"ev001": {
			Metadata: dto.NewEventMetadata("ev001", common.GameVersionFFX),
			Rows: []dto.TextRow{
				{Index: 0, Hash: hash.Texts(map[string]string{"us": "Hi"}), Text: map[string]string{"us": "Hi", "sp": ""}},
			},
		},
	}
	raw, err := fmtstrings.Marshal(c)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(raw), ":sp║$") {
		t.Fatalf("empty sp cell must be skipped:\n%s", raw)
	}
	// Def vazio escrito à mão é preservado como "".
	hand := "/*key=ffx/event/obj_ps3/ev/ev001/ev001.bin row_count=1*/\n" +
		"ffx:ev001:0:us║$" + hash.Sum64Hex("Hi") + " = Hi\n" +
		"ffx:ev001:0:sp║$" + hash.Sum64Hex("") + " = \n"
	back, err := fmtstrings.Unmarshal([]byte(hand))
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if v, ok := back["ev001"].Rows[0].Text["sp"]; !ok || v != "" {
		t.Fatalf("empty def not preserved: %+v", back["ev001"].Rows[0])
	}
}

func TestMalformedLines(t *testing.T) {
	for _, line := range []string{
		"ffx:ev001:0:us $abc = Hi",               // sem ║
		"ffx:ev001:0:us║Hi",                      // sem =
		"ffx:ev001:0:us║abc = Hi",                // sem $ no hash
		"ffx:ev001:0:us║$xyz = Hi",               // hash não-hex
		"ffx:ev001:xx:us║$0000000000000000 = Hi", // index inválido
		"ffx:us║$0000000000000000 = Hi",          // sem index
	} {
		if _, err := fmtstrings.Unmarshal([]byte(line + "\n")); err == nil {
			t.Fatalf("expected error for %q", line)
		}
	}
}

func TestRightAnchorGrowth(t *testing.T) {
	// Campos novos crescem pela esquerda: id/name continuam resolvidos.
	raw := "/*key=ffx/battle/kernel/command.bin row_count=1*/\n" +
		"region:ffx:command:name:0:us║$" + hash.Sum64Hex("Potion") + " = Potion\n"
	back, err := fmtstrings.Unmarshal([]byte(raw))
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	entry, ok := back["command"]
	if !ok {
		t.Fatalf("command not resolved: %+v", back.SortedKeys())
	}
	if entry.Rows[0].Name != "name" || entry.Rows[0].Text["us"] != "Potion" {
		t.Fatalf("wrong row: %+v", entry.Rows[0])
	}
}
