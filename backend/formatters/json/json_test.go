package json_test

import (
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
	if ev001.Metadata.EventID != "ev001" || ev001.Metadata.FileName != "ev001.bin" {
		t.Fatalf("metadata not preserved: %+v", ev001.Metadata)
	}
}

func TestJSONEventsFormatterShapesOutput(t *testing.T) {
	raw, err := json.NewJSONEventsFormatter().Marshal(sampleCollection())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	out := string(raw)
	for _, want := range []string{`"rows"`, `"metadata"`, `"event_id"`, `"hash"`, "ev001", "Tom & Jerry"} {
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
	if entry.Metadata.ChunkIndex == nil || *entry.Metadata.ChunkIndex != 0 {
		t.Fatalf("chunk index not preserved: %+v", entry.Metadata)
	}
}

func TestMetadataOmitsEmptyFields(t *testing.T) {
	// Events não fornecem file_info/chunk_index: não podem sair no JSON.
	raw, err := json.NewJSONEventsFormatter().Marshal(sampleCollection())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	out := string(raw)
	for _, want := range []string{`"file_info"`, `"chunk_index"`, `"name":`, `:null`, `"hash":{}`} {
		if strings.Contains(out, want) {
			t.Fatalf("unexpected %s in output:\n%s", want, out)
		}
	}
	// Macro não fornece event_id/file_info: omitidos; chunk_index presente.
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
	for _, want := range []string{`"event_id"`, `"file_info"`} {
		if strings.Contains(macroOut, want) {
			t.Fatalf("unexpected %s in macro output:\n%s", want, macroOut)
		}
	}
	if !strings.Contains(macroOut, `"chunk_index"`) {
		t.Fatalf("expected chunk_index in macro output:\n%s", macroOut)
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
