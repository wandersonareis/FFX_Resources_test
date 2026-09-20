package builders_test

import (
	"testing"

	"ffxresources/backend/builders"
	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/event"
)

func seedEvent(id string) {
	ev := &event.EventFile{
		ID:      id,
		Version: common.GameVersionFFX,
		Strings: []*event.LocalizedFieldStringObject{
			event.NewLocalizedFieldStringObject(),
		},
	}
	event.SetEvent(common.GameVersionFFX, id, ev)
}

func TestResolveEmptyReturnsAll(t *testing.T) {
	event.ClearEvents(common.GameVersionFFX)
	seedEvent("ev002")
	seedEvent("ev001")
	got, err := builders.ResolveEventTargets(common.GameVersionFFX, nil)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if len(got) != 2 || got[0] != "ev001" || got[1] != "ev002" {
		t.Fatalf("expected sorted all, got %v", got)
	}
}

func TestResolveSelectedStrict(t *testing.T) {
	event.ClearEvents(common.GameVersionFFX)
	seedEvent("ev001")
	got, err := builders.ResolveEventTargets(common.GameVersionFFX, []string{"ev001", "missing"})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if len(got) != 1 || got[0] != "ev001" {
		t.Fatalf("expected only found, got %v", got)
	}
}

func TestResolveNoneFoundErrorsWithoutFallback(t *testing.T) {
	event.ClearEvents(common.GameVersionFFX)
	seedEvent("ev001")
	if _, err := builders.ResolveEventTargets(common.GameVersionFFX, []string{"nope", "missing"}); err == nil {
		t.Fatalf("expected error when zero found, must not fall back to all")
	}
}

func TestBuildEventsDTOKeysAndRows(t *testing.T) {
	event.ClearEvents(common.GameVersionFFX)
	seedEvent("ev001")
	c, err := builders.BuildEventsDTO(common.GameVersionFFX, []string{"ev001"})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	entry, ok := c["ev001"]
	if !ok {
		t.Fatalf("missing key ev001: %v", c.SortedKeys())
	}
	if len(entry.Rows) != 1 || entry.Rows[0].Index != 0 {
		t.Fatalf("unexpected rows: %+v", entry.Rows)
	}
	if entry.Metadata.EventID != "ev001" || entry.Metadata.FileName != "ev001.bin" {
		t.Fatalf("metadata mismatch: %+v", entry.Metadata)
	}
}
