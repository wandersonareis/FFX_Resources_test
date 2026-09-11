package macrodic

import (
	"encoding/json"
	"sort"
)

// MacroStringJsonExport represents a single macro string for JSON serialization.
// Name maps each localization to its text, covering all languages in one object
// (objectfile ExportToJson pattern). SimplifiedName maps each localization to its
// abbreviated text and is omitted when empty: a present entry means the binary
// holds distinct Name/SimplifiedName offsets, an absent one means both offsets
// are equal and point to the same text.
type MacroStringJsonExport struct {
	Index          int               `json:"index"`
	Name           map[string]string `json:"name,omitempty"`
	SimplifiedName map[string]string `json:"simplifiedName,omitempty"`
}

// MacroChunkJsonExport represents a chunk of macro strings for JSON serialization.
type MacroChunkJsonExport struct {
	ChunkIndex int                     `json:"chunkIndex"`
	Strings    []MacroStringJsonExport `json:"strings"`
}

// MacroDictionaryJsonExport is the merged macro dictionary export: one entry per
// (chunk, string) holding the texts of every localization.
type MacroDictionaryJsonExport struct {
	Chunks []MacroChunkJsonExport `json:"chunks"`
}

// fillMacroLocalized adds a localization text to the map, mirroring the objectfile
// fillLocalized helper: empty texts are skipped, the map is allocated on demand.
func fillMacroLocalized(m map[string]string, text, locKey string) map[string]string {
	if text == "" {
		return m
	}
	if m == nil {
		m = make(map[string]string)
	}
	m[locKey] = text
	return m
}

// ExportToJson exports all given localization containers into a single merged JSON
// document, following the objectfile ExportToJson pattern: it iterates over every
// localization and fills the per-entry Name/SimplifiedName maps. Entries without
// any text in any language are skipped.
func ExportToJson(containers map[string]*MacroDictionaryTextContainer) *MacroDictionaryJsonExport {
	locKeys := SortedLocalizationKeys(containers)

	type entryKey struct {
		chunk int
		index int
	}
	names := make(map[entryKey]map[string]string)
	simplified := make(map[entryKey]map[string]string)

	for _, loc := range locKeys {
		c := containers[loc]
		if c == nil {
			continue
		}
		for chunkIndex, strings := range c.MapAllStrings() {
			for stringIndex, s := range strings {
				if s == nil {
					continue
				}
			k := entryKey{chunk: chunkIndex, index: stringIndex}
			// Always register the key: the format is positional, so even
			// textless entries are preserved (unlike objectfile's hasContent
			// skip, which suits its ID-keyed layout).
			if _, ok := names[k]; !ok {
				names[k] = nil
			}
			names[k] = fillMacroLocalized(names[k], s.GetRegularString(), loc)
			if s.HasDistinctSimplified() {
				simplified[k] = fillMacroLocalized(simplified[k], s.GetSimplifiedString(), loc)
			}
			}
		}
	}

	keys := make([]entryKey, 0, len(names))
	for k := range names {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(a, b int) bool {
		if keys[a].chunk != keys[b].chunk {
			return keys[a].chunk < keys[b].chunk
		}
		return keys[a].index < keys[b].index
	})

	chunks := make([]MacroChunkJsonExport, 0)
	for _, k := range keys {
		if len(chunks) == 0 || chunks[len(chunks)-1].ChunkIndex != k.chunk {
			chunks = append(chunks, MacroChunkJsonExport{
				ChunkIndex: k.chunk,
				Strings:    make([]MacroStringJsonExport, 0),
			})
		}
		last := &chunks[len(chunks)-1]
		last.Strings = append(last.Strings, MacroStringJsonExport{
			Index:          k.index,
			Name:           names[k],
			SimplifiedName: simplified[k],
		})
	}

	return &MacroDictionaryJsonExport{Chunks: chunks}
}

// MarshalToJson serializes the merged export to JSON bytes with proper indentation.
func MarshalToJson(export *MacroDictionaryJsonExport) ([]byte, error) {
	return json.MarshalIndent(export, "", "  ")
}

// SortedLocalizationKeys returns the sorted localization keys of the given
// containers for deterministic export output.
func SortedLocalizationKeys(containers map[string]*MacroDictionaryTextContainer) []string {
	keys := make([]string, 0, len(containers))
	for loc := range containers {
		keys = append(keys, loc)
	}
	sort.Strings(keys)
	return keys
}
