package json_test

import (
	encodingjson "encoding/json"
	"strings"
	"testing"

	"ffxresources/backend/common"
	"ffxresources/backend/dto"
	"ffxresources/backend/formatters/hash"
	"ffxresources/backend/formatters/json"
)

func sampleCollection() dto.Collection {
	return dto.Collection{
		"ev002": {
			Metadata: dto.NewEventMetadata("ev002", common.GameVersionFFX),
			Rows: []dto.TextRow{
				{Index: 0, Hash: hash.Texts(map[string]string{"us": "Hi"}), Text: map[string]string{"us": "Hi"}},
			},
		},
		"ev001": {
			Metadata: dto.NewEventMetadata("ev001", common.GameVersionFFX),
			Rows: []dto.TextRow{
				{
					Index: 0,
					Hash:  hash.Texts(map[string]string{"us": "Tom & Jerry", "sp": "Tom y Jerry"}),
					Text:  map[string]string{"us": "Tom & Jerry", "sp": "Tom y Jerry"},
				},
			},
		},
	}
}

func TestJSONEventsFormatterRoundTrip(t *testing.T) {
	f := json.NewJSONEventsFormatter()
	if f.Extension() != ".json" {
		t.Fatalf("unexpected extension: %s", f.Extension())
	}
	raw, err := f.Marshal(sampleCollection())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	back, err := f.Unmarshal(raw)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(back) != 2 {
		t.Fatalf("expected 2 events, got %d", len(back))
	}
	ev001, ok := back["ev001"]
	if !ok {
		t.Fatalf("missing ev001: %+v", back.SortedKeys())
	}
	if len(ev001.Rows) != 1 || ev001.Rows[0].Index != 0 {
		t.Fatalf("rows not preserved: %+v", ev001.Rows)
	}
	if got := ev001.Rows[0].Text["us"]; got != "Tom & Jerry" {
		t.Fatalf("roundtrip mismatch: %q", got)
	}
	if got := ev001.Rows[0].Hash["us"]; got != hash.Sum64Hex("Tom & Jerry") {
		t.Fatalf("hash mismatch: %q", got)
	}
	if ev001.Metadata.ID != "ev001" {
		t.Fatalf("metadata not preserved: %+v", ev001.Metadata)
	}
}

func TestJSONEventsFormatterShapesOutput(t *testing.T) {
	raw, err := json.NewJSONEventsFormatter().Marshal(sampleCollection())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	out := string(raw)
	for _, want := range []string{`"rows"`, `"metadata"`, `"key"`, `"id"`, `"hash"`, "ev001", "Tom & Jerry"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %s:\n%s", want, out)
		}
	}
	if strings.Contains(out, `\u0026`) {
		t.Fatalf("unexpected unicode escape:\n%s", out)
	}
}

func TestJSONEventsFormatterSortsRows(t *testing.T) {
	c := dto.Collection{
		"ev001": {
			Metadata: dto.NewEventMetadata("ev001", common.GameVersionFFX),
			Rows: []dto.TextRow{
				{Index: 2, Text: map[string]string{"us": "C"}},
				{Index: 0, Text: map[string]string{"us": "A"}},
				{Index: 1, Text: map[string]string{"us": "B"}},
			},
		},
	}
	raw, err := json.NewJSONEventsFormatter().Marshal(c)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	back, err := json.NewJSONEventsFormatter().Unmarshal(raw)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	rows := back["ev001"].Rows
	for i, want := range []int{0, 1, 2} {
		if rows[i].Index != want {
			t.Fatalf("rows not sorted: %+v", rows)
		}
	}
}

func TestJSONObjectFormatterRoundTrip(t *testing.T) {
	f := json.NewJSONObjectFormatter()
	if f.Extension() != ".json" {
		t.Fatalf("unexpected extension: %s", f.Extension())
	}
	version := common.GameVersionFFX
	c := dto.Collection{
		"command": {
			Metadata: dto.NewObjectMetadata(version, "battle/kernel", "command.bin", "ffx/battle/kernel/command.bin"),
			Rows: []dto.TextRow{
				{
					Index: 0,
					Name:  "name",
					Hash:  hash.Texts(map[string]string{"us": "Potion & More"}),
					Text:  map[string]string{"us": "Potion & More"},
				},
				{
					Index: 1,
					Name:  "description",
					Hash:  hash.Texts(map[string]string{"us": "Restores HP"}),
					Text:  map[string]string{"us": "Restores HP"},
				},
			},
		},
	}
	raw, err := f.Marshal(c)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	back, err := f.Unmarshal(raw)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	entry, ok := back["command"]
	if !ok {
		t.Fatalf("missing command: %+v", back.SortedKeys())
	}
	if len(entry.Rows) != 2 {
		t.Fatalf("roundtrip mismatch: %+v", entry.Rows)
	}
	if got := entry.Rows[0].Text["us"]; got != "Potion & More" {
		t.Fatalf("roundtrip text mismatch: %q", got)
	}
	if entry.Rows[0].Name != "name" || entry.Rows[1].Name != "description" {
		t.Fatalf("field names not preserved: %+v", entry.Rows)
	}
	if entry.Metadata.Key != "ffx/battle/kernel/command.bin" {
		t.Fatalf("metadata key mismatch: %+v", entry.Metadata)
	}
}

// TestJSONPreservesLayoutFieldOrder garante que campos de um mesmo objeto
// (mesmo Index) saem na ordem do layout, não em ordem alfabética por Name.
func TestJSONPreservesLayoutFieldOrder(t *testing.T) {
	f := json.NewJSONObjectFormatter()
	version := common.GameVersionFFX
	order := []string{"name", "simplifiedName", "description", "simplifiedDescription"}
	rows := make([]dto.TextRow, 0, len(order))
	for _, name := range order {
		text := map[string]string{"us": name + " text"}
		rows = append(rows, dto.TextRow{Index: 0, Name: name, Hash: hash.Texts(text), Text: text})
	}
	c := dto.Collection{
		"command": {
			Metadata: dto.NewObjectMetadata(version, "battle/kernel", "command.bin", "ffx/battle/kernel/command.bin"),
			Rows:     rows,
		},
	}
	raw, err := f.Marshal(c)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	back, err := f.Unmarshal(raw)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	got := back["command"].Rows
	if len(got) != len(order) {
		t.Fatalf("row count: %d vs %d", len(got), len(order))
	}
	for i, want := range order {
		if got[i].Name != want {
			t.Fatalf("field order at %d: got %q, want %q (rows=%+v)", i, got[i].Name, want, got)
		}
	}
}

func TestJSONMacroFormatterRoundTrip(t *testing.T) {
	f := json.NewJSONMacroFormatter()
	if f.Extension() != ".json" {
		t.Fatalf("unexpected extension: %s", f.Extension())
	}
	c := dto.Collection{
		"chunk_00": {
			Metadata: dto.NewMacroMetadata(common.GameVersionFFX, 0),
			Rows: []dto.TextRow{
				{
					Index: 0,
					Name:  "name",
					Hash:  hash.Texts(map[string]string{"us": "Tom & Jerry"}),
					Text:  map[string]string{"us": "Tom & Jerry"},
				},
				{
					Index: 0,
					Name:  "simplifiedName",
					Hash:  hash.Texts(map[string]string{"us": "Tom & Jerry!"}),
					Text:  map[string]string{"us": "Tom & Jerry!"},
				},
			},
		},
	}
	raw, err := f.Marshal(c)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	back, err := f.Unmarshal(raw)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	entry, ok := back["chunk_00"]
	if !ok {
		t.Fatalf("missing chunk_00: %+v", back.SortedKeys())
	}
	if len(entry.Rows) != 2 {
		t.Fatalf("roundtrip mismatch: %+v", entry.Rows)
	}
	if entry.Metadata.ID != "chunk_00" {
		t.Fatalf("chunk id not preserved: %+v", entry.Metadata)
	}
}

func TestMetadataOmitsEmptyFields(t *testing.T) {
	// Novo formato: só key/row_count/id/is_dir podem sair no JSON.
	raw, err := json.NewJSONEventsFormatter().Marshal(sampleCollection())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	out := string(raw)
	for _, want := range []string{`"event_id"`, `"file_info"`, `"chunk_index"`, `"version"`, `"dir_pattern"`, `"file_name"`, `"shortened"`, `"mid_path"`, `"event_file_path"`, `"localization_pattern"`, `"name":`, `:null`, `"hash":{}`} {
		if strings.Contains(out, want) {
			t.Fatalf("unexpected %s in output:\n%s", want, out)
		}
	}
	// Macro carrega id=chunk_NN; row_count omitido quando zero.
	macroRaw, err := json.NewJSONMacroFormatter().Marshal(dto.Collection{
		"chunk_00": {
			Metadata: dto.NewMacroMetadata(common.GameVersionFFX, 0),
			Rows:     []dto.TextRow{{Index: 0, Name: "name", Text: map[string]string{"us": "Hi"}}},
		},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	macroOut := string(macroRaw)
	for _, want := range []string{`"event_id"`, `"file_info"`, `"chunk_index"`} {
		if strings.Contains(macroOut, want) {
			t.Fatalf("unexpected %s in macro output:\n%s", want, macroOut)
		}
	}
	if !strings.Contains(macroOut, `"chunk_00"`) {
		t.Fatalf("expected chunk id in macro output:\n%s", macroOut)
	}
}

func TestHashPrefixStrip(t *testing.T) {
	if got := hash.Prefix("abc"); got != "$abc" {
		t.Fatalf("prefix mismatch: %q", got)
	}
	if got := hash.Prefix("$abc"); got != "$abc" {
		t.Fatalf("prefix must be idempotent: %q", got)
	}
	if bare, isRef := hash.Strip("$abc"); !isRef || bare != "abc" {
		t.Fatalf("strip mismatch: %q %v", bare, isRef)
	}
	if bare, isRef := hash.Strip("abc"); isRef || bare != "abc" {
		t.Fatalf("strip must pass through: %q %v", bare, isRef)
	}
}

func dedupCollection() dto.Collection {
	long := "{PC:04:WAKKA}{\n}{PC:01:YUNA}! Onde você está?!"
	return dto.Collection{
		"boss_fight": {
			Metadata: dto.NewEventMetadata("boss_fight", common.GameVersionFFX),
			Rows: []dto.TextRow{
				{Index: 0, Hash: hash.Texts(map[string]string{"us": long}), Text: map[string]string{"us": long}},
				{Index: 7, Hash: hash.Texts(map[string]string{"us": long}), Text: map[string]string{"us": long}},
				{Index: 8, Hash: hash.Texts(map[string]string{"us": "ok", "sp": long}), Text: map[string]string{"us": "ok", "sp": long}},
			},
		},
	}
}

func TestMarshalDedupsDefaultLangRepeats(t *testing.T) {
	f := json.NewJSONEventsFormatter()
	raw, err := f.Marshal(dedupCollection())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	// Inspeciona o JSON bruto estruturalmente: primeira ocorrência integral,
	// repetição como ref, hashes com $ em todos os idiomas.
	var doc map[string]dto.FileEntry
	if err := encodingjson.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("reparse: %v", err)
	}
	long := "{PC:04:WAKKA}{\n}{PC:01:YUNA}! Onde você está?!"
	h := hash.Sum64Hex(long)
	rows := doc["boss_fight"].Rows
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %+v", rows)
	}
	if rows[0].Text["us"] != long {
		t.Fatalf("first occurrence must stay literal: %q", rows[0].Text["us"])
	}
	if rows[0].Hash["us"] != "$"+h || rows[1].Hash["us"] != "$"+h {
		t.Fatalf("hashes must gain $ prefix: %+v", rows)
	}
	if rows[1].Text["us"] != "$"+h {
		t.Fatalf("repeat must become ref: %q", rows[1].Text["us"])
	}
	if rows[2].Hash["sp"] != "$"+h {
		t.Fatalf("non-default hash must gain $ too: %q", rows[2].Hash["sp"])
	}
	if rows[2].Text["sp"] != long {
		t.Fatalf("non-default text must stay literal: %q", rows[2].Text["sp"])
	}
	back, err := f.Unmarshal(raw)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	brows := back["boss_fight"].Rows
	if brows[1].Text["us"] != long {
		t.Fatalf("ref not resolved: %q", brows[1].Text["us"])
	}
	if brows[1].Hash["us"] != h {
		t.Fatalf("hash must come back bare: %q", brows[1].Hash["us"])
	}
	// "ok" (< 5 runes) fica literal mesmo repetido.
	if brows[2].Text["us"] != "ok" {
		t.Fatalf("short text must stay literal: %q", brows[2].Text["us"])
	}
}

func TestMarshalSkipsShortCJKTexts(t *testing.T) {
	// "你好" tem 2 runes mas 6 bytes: conta como curta, fica literal.
	text := "你好"
	if got := len([]rune(text)); got != 2 {
		t.Fatalf("fixture assumption broken: %d runes", got)
	}
	c := dto.Collection{
		"ev001": {
			Metadata: dto.NewEventMetadata("ev001", common.GameVersionFFX),
			Rows: []dto.TextRow{
				{Index: 0, Hash: hash.Texts(map[string]string{"us": text}), Text: map[string]string{"us": text}},
				{Index: 1, Hash: hash.Texts(map[string]string{"us": text}), Text: map[string]string{"us": text}},
			},
		},
	}
	raw, err := json.NewJSONEventsFormatter().Marshal(c)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	back, err := json.NewJSONEventsFormatter().Unmarshal(raw)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, row := range back["ev001"].Rows {
		if row.Text["us"] != text {
			t.Fatalf("short CJK text must stay literal: %q", row.Text["us"])
		}
	}
}

func TestUnmarshalExpandsEditedValue(t *testing.T) {
	// Cenário do usuário: texto da row 0 editado sob hash antigo; a ref da
	// row 7 deve explodir para o NOVO valor usando o hash existente.
	h := hash.Sum64Hex("{PC:04:WAKKA} original")
	raw := []byte(`{"boss_fight": {"metadata": {}, "rows": [
		{"index": 0, "hash": {"us": "$` + h + `"}, "text": {"us": "Texto alterado"}},
		{"index": 7, "hash": {"us": "$` + h + `"}, "text": {"us": "$` + h + `"}}
	]}}`)
	back, err := json.NewJSONEventsFormatter().Unmarshal(raw)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	rows := back["boss_fight"].Rows
	if rows[0].Text["us"] != "Texto alterado" {
		t.Fatalf("edited literal not kept: %q", rows[0].Text["us"])
	}
	if rows[1].Text["us"] != "Texto alterado" {
		t.Fatalf("ref not expanded to edited value: %q", rows[1].Text["us"])
	}
	if rows[0].Hash["us"] != h || rows[1].Hash["us"] != h {
		t.Fatalf("hashes must come back bare: %+v", rows)
	}
}

func TestUnmarshalKeepsOrphanAndLiteralDollar(t *testing.T) {
	h := hash.Sum64Hex("algum texto longo aqui")
	raw := []byte(`{"ev001": {"metadata": {}, "rows": [
		{"index": 0, "hash": {"us": "$deadbeefdeadbeef"}, "text": {"us": "$deadbeefdeadbeef"}},
		{"index": 1, "hash": {"us": "$` + h + `"}, "text": {"us": "$nao-eh-ref"}}
	]}}`)
	back, err := json.NewJSONEventsFormatter().Unmarshal(raw)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	rows := back["ev001"].Rows
	if rows[0].Text["us"] != "$deadbeefdeadbeef" {
		t.Fatalf("orphan ref must be kept: %q", rows[0].Text["us"])
	}
	if rows[0].Hash["us"] != "deadbeefdeadbeef" {
		t.Fatalf("orphan hash must be stripped: %q", rows[0].Hash["us"])
	}
	if rows[1].Text["us"] != "$nao-eh-ref" {
		t.Fatalf("literal dollar text must be kept: %q", rows[1].Text["us"])
	}
}

func TestMarshalUnmarshalIdentity(t *testing.T) {
	in := dedupCollection()
	raw, err := json.NewJSONEventsFormatter().Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	back, err := json.NewJSONEventsFormatter().Unmarshal(raw)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	inRows := in["boss_fight"].Rows
	backRows := back["boss_fight"].Rows
	if len(inRows) != len(backRows) {
		t.Fatalf("row count mismatch: %d vs %d", len(inRows), len(backRows))
	}
	for i := range inRows {
		if inRows[i].Text["us"] != backRows[i].Text["us"] {
			t.Fatalf("row %d text mismatch: %q vs %q", i, inRows[i].Text["us"], backRows[i].Text["us"])
		}
		if inRows[i].Hash["us"] != backRows[i].Hash["us"] {
			t.Fatalf("row %d hash mismatch: %q vs %q", i, inRows[i].Hash["us"], backRows[i].Hash["us"])
		}
	}
}

func TestMarshalNoEscapeKeepsAmpersand(t *testing.T) {
	c := dto.Collection{
		"ev001": {
			Metadata: dto.NewEventMetadata("ev001", common.GameVersionFFX),
			Rows: []dto.TextRow{
				{Index: 0, Text: map[string]string{"us": "Tom & Jerry <test>"}},
			},
		},
	}
	raw, err := json.NewJSONEventsFormatter().Marshal(c)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(raw), "Tom & Jerry <test>") {
		t.Fatalf("expected literals: %s", raw)
	}
	if strings.Contains(string(raw), `\u`) {
		t.Fatalf("unexpected unicode escape: %s", raw)
	}
}

func TestMarshalLangsFilter(t *testing.T) {
	raw, err := json.NewJSONEventsFormatter().MarshalLangs(sampleCollection(), []string{"us"})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	out := string(raw)
	if strings.Contains(out, `"sp"`) {
		t.Fatalf("sp must be filtered:\n%s", out)
	}
	if !strings.Contains(out, `"us"`) {
		t.Fatalf("us must be kept:\n%s", out)
	}
	back, err := json.NewJSONEventsFormatter().Unmarshal(raw)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := back["ev001"].Rows[0].Text["sp"]; ok {
		t.Fatalf("sp must not roundtrip: %+v", back["ev001"].Rows[0])
	}
	if back["ev001"].Rows[0].Text["us"] != "Tom & Jerry" {
		t.Fatalf("us lost: %+v", back["ev001"].Rows[0])
	}
	// nil = todos (compatível com Marshal).
	full, err := json.NewJSONEventsFormatter().MarshalLangs(sampleCollection(), nil)
	if err != nil {
		t.Fatalf("marshal nil langs: %v", err)
	}
	if !strings.Contains(string(full), `"sp"`) {
		t.Fatalf("nil langs must keep all:\n%s", full)
	}
}
