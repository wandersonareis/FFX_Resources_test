package hash_test

import (
	"testing"

	"ffxresources/backend/dto"
	"ffxresources/backend/formatters/hash"
)

func TestSum64HexStable(t *testing.T) {
	a := hash.Sum64Hex("Hi")
	b := hash.Sum64Hex("Hi")
	if a != b || len(a) != 16 {
		t.Fatalf("unstable or bad length: %q %q", a, b)
	}
	if hash.Sum64Hex("Hi") == hash.Sum64Hex("Bye") {
		t.Fatalf("unexpected equality for different texts")
	}
}

func TestTextsSkipsEmpty(t *testing.T) {
	out := hash.Texts(map[string]string{"us": "Hi", "sp": ""})
	if _, ok := out["sp"]; ok {
		t.Fatalf("empty text should not produce hash: %+v", out)
	}
	if out["us"] == "" {
		t.Fatalf("missing hash for non-empty text")
	}
	if hash.Texts(map[string]string{}) != nil {
		t.Fatalf("expected nil for empty input")
	}
}

func TestValidateNoCollisionDetects(t *testing.T) {
	h := hash.Sum64Hex("A")
	c := dto.Collection{
		"ev001": {
			Rows: []dto.TextRow{
				{Index: 0, Hash: map[string]string{"us": h}, Text: map[string]string{"us": "A"}},
				{Index: 1, Hash: map[string]string{"us": h}, Text: map[string]string{"us": "B"}},
			},
		},
	}
	if err := hash.ValidateNoCollision(c); err == nil {
		t.Fatalf("expected collision error for same hash with different texts")
	}
}

func TestValidateNoCollisionAllowsDedup(t *testing.T) {
	h := hash.Sum64Hex("Same")
	c := dto.Collection{
		"ev001": {
			Rows: []dto.TextRow{
				{Index: 0, Hash: map[string]string{"us": h}, Text: map[string]string{"us": "Same"}},
				{Index: 1, Hash: map[string]string{"us": h}, Text: map[string]string{"us": "Same"}},
			},
		},
	}
	if err := hash.ValidateNoCollision(c); err != nil {
		t.Fatalf("legit dedup should pass: %v", err)
	}
}

func TestDedupIndexGroups(t *testing.T) {
	h := hash.Sum64Hex("Hi")
	rows := []dto.TextRow{
		{Index: 0, Hash: map[string]string{"us": h}, Text: map[string]string{"us": "Hi"}},
		{Index: 1, Hash: map[string]string{"us": h}, Text: map[string]string{"us": "Hi"}},
		{Index: 2, Text: map[string]string{"us": "Other"}},
	}
	got := hash.DedupIndex(rows, "us")
	if len(got[h]) != 2 || got[h][0] != 0 || got[h][1] != 1 {
		t.Fatalf("unexpected grouping: %+v", got)
	}
}
