package builders_test

import (
	"testing"

	"ffxresources/backend/builders"
	"ffxresources/backend/common"
	"ffxresources/backend/dto"
	"ffxresources/backend/fileFormats/helpfile"
	"ffxresources/backend/formatters/hash"
)

func TestApplyHelpDivergentCopyStaysIndependent(t *testing.T) {
	const original = "Same help text shared by the definition and copies"
	const defText = "Definition translation A"
	const divergentText = "Copy-specific translation B"
	bare := hash.Sum64Hex(original)

	panel := func(name string) *helpfile.HelpKeyedStringFile {
		return &helpfile.HelpKeyedStringFile{
			Name: name,
			Files: map[string]*helpfile.HelpBinaryFile{
				common.DefaultLocalization: {
					Segments: []*helpfile.HelpSegment{{Index: 0, Text: original}},
				},
			},
		}
	}
	panels := map[string]*helpfile.HelpKeyedStringFile{
		// Divergent key sorts before definition to prove behavior is not
		// dependent on the write order.
		"aaa_divergent":  panel("aaa_divergent"),
		"zzz_definition": panel("zzz_definition"),
		"zzz_reference":  panel("zzz_reference"),
		"zzz_twin":       panel("zzz_twin"),
	}
	scope := builders.HelpApplyScope{
		Ids:   []string{"aaa_divergent", "zzz_definition", "zzz_reference", "zzz_twin"},
		Panel: func(name string) *helpfile.HelpKeyedStringFile { return panels[name] },
		Save:  func(string) error { return nil },
	}
	collection := dto.Collection{
		"aaa_divergent": {Rows: []dto.TextRow{{
			Index: 0, Hash: map[string]string{"us": bare},
			Text: map[string]string{"us": divergentText}, Divergent: true,
		}}},
		"zzz_definition": {Rows: []dto.TextRow{{
			Index: 0, Hash: map[string]string{"us": bare}, Text: map[string]string{"us": defText},
		}}},
		"zzz_reference": {Rows: []dto.TextRow{{
			Index: 0, Hash: map[string]string{"us": bare}, Text: map[string]string{"us": hash.Prefix(bare)},
		}}},
	}
	if err := builders.ApplyHelpDTOWithScope(common.GameVersionFFX, collection, collection.SortedKeys(), scope); err != nil {
		t.Fatalf("apply: %v", err)
	}
	for name, want := range map[string]string{
		"aaa_divergent":  divergentText,
		"zzz_definition": defText,
		"zzz_reference":  defText,
		"zzz_twin":       defText,
	} {
		got := panels[name].Files[common.DefaultLocalization].Segments[0].Text
		if got != want {
			t.Errorf("%s[0] = %q, esperado %q", name, got, want)
		}
	}
}
