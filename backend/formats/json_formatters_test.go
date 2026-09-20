package formats

import (
	"strings"
	"testing"

	"ffxresources/backend/common"
	"ffxresources/backend/datastore"
	"ffxresources/backend/fileFormats/event"
)

func sampleEvents() []event.EventFileData {
	return []event.EventFileData{
		{
			ID: "ev002",
			Strings: []event.EventStringData{
				{Index: 0, Text: map[string]string{"us": "Hi"}},
			},
		},
		{
			ID: "ev001",
			Strings: []event.EventStringData{
				{Index: 0, Text: map[string]string{"us": "Tom & Jerry", "sp": "Tom y Jerry"}},
			},
		},
	}
}

func TestJSONEventsFormatterRoundTrip(t *testing.T) {
	f := NewJSONEventsFormatter()
	if f.Extension() != ".json" {
		t.Fatalf("unexpected extension: %s", f.Extension())
	}
	raw, err := f.Marshal(sampleEvents(), common.GameVersionFFX)
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
	if back[0].ID != "ev001" || back[1].ID != "ev002" {
		t.Fatalf("expected sorted IDs, got %q %q", back[0].ID, back[1].ID)
	}
	if got := back[0].Strings[0].Text["us"]; got != "Tom & Jerry" {
		t.Fatalf("roundtrip mismatch: %q", got)
	}
}

func TestJSONEventsFormatterShapesOutput(t *testing.T) {
	raw, err := NewJSONEventsFormatter().Marshal(sampleEvents(), common.GameVersionFFX)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	out := string(raw)
	for _, want := range []string{`"data"`, `"strings"`, `"metadata"`, `"event_id"`, "ev001", "Tom & Jerry"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %s:\n%s", want, out)
		}
	}
	if strings.Contains(out, `\u0026`) {
		t.Fatalf("unexpected unicode escape:\n%s", out)
	}
}

func TestJSONEventsFormatterReadsLegacyBarePayload(t *testing.T) {
	raw := []byte(`{"strings":{"ev001":{"id":"ev001","strings":[{"index":0,"text":{"us":"Hi"}}]}}}`)
	back, err := NewJSONEventsFormatter().Unmarshal(raw)
	if err != nil {
		t.Fatalf("unmarshal bare: %v", err)
	}
	if len(back) != 1 || back[0].ID != "ev001" {
		t.Fatalf("bare fallback mismatch: %+v", back)
	}
}

func TestJSONObjectFormatterRoundTrip(t *testing.T) {
	f := NewJSONObjectFormatter()
	if f.Extension() != ".json" {
		t.Fatalf("unexpected extension: %s", f.Extension())
	}
	data := datastore.ObjectTextData{
		Entries: []*datastore.ObjectTextEntry{
			{ID: 0, Name: map[string]string{"us": "Potion & More"}},
		},
	}
	raw, err := f.Marshal(data)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	// Nota: o Strings interno preserva o escape legado (\u0026), byte a byte
	// como o formato original. O roundtrip abaixo prova o texto intacto.
	back, err := f.Unmarshal(raw)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(back.Entries) != 1 || back.Entries[0].ID != 0 {
		t.Fatalf("roundtrip mismatch: %+v", back.Entries)
	}
	if got := back.Entries[0].Name["us"]; got != "Potion & More" {
		t.Fatalf("roundtrip text mismatch: %q", got)
	}
}

func TestJSONMacroFormatterUnmarshal(t *testing.T) {
	raw := []byte(`{"chunks":[{"chunkIndex":0,"strings":[{"index":0,"name":{"us":"Tom & Jerry"}}]}]}`)
	imp, err := NewJSONMacroFormatter().Unmarshal(raw)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := imp.Chunks[0].Strings[0].Name["us"]; got != "Tom & Jerry" {
		t.Fatalf("mismatch: %+v", imp)
	}
}

func TestMarshalNoEscapeKeepsAmpersand(t *testing.T) {
	raw, err := marshalNoEscape(map[string]string{"text": "Tom & Jerry <test>"})
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
