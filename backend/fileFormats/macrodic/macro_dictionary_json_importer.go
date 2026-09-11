package macrodic

import (
	"encoding/json"
	"sort"

	"ffxresources/backend/core/converter"
	"ffxresources/backend/core/encoding"
	"ffxresources/backend/models"
	"fmt"
)

// MacroStringJsonImport represents a single macro string for JSON deserialization.
// Name maps each localization to its text. SimplifiedName maps each localization
// to its abbreviated text and is optional per language:
//   - If present and non-empty for a language: that language gets an independent
//     SimplifiedName pointer (SimplifiedOffset != RegularOffset in binary).
//   - If absent or empty: the pointer falls back to the Name pointer
//     (SimplifiedOffset == RegularOffset, both point to the same text).
type MacroStringJsonImport struct {
	Index          int               `json:"index"`
	Name           map[string]string `json:"name,omitempty"`
	SimplifiedName map[string]string `json:"simplifiedName,omitempty"`
}

// MacroChunkJsonImport represents a chunk of macro strings for JSON deserialization.
type MacroChunkJsonImport struct {
	ChunkIndex int                     `json:"chunkIndex"`
	Strings    []MacroStringJsonImport `json:"strings"`
}

// MacroDictionaryJsonImport is the merged macro dictionary import: one entry per
// (chunk, string) holding the texts of every localization.
type MacroDictionaryJsonImport struct {
	Chunks []MacroChunkJsonImport `json:"chunks"`
}

// ImportFromJson imports a merged JSON representation, rebuilding one binary
// container per localization found in the data. Segment texts are converted with
// the per-language charset; a missing/empty SimplifiedName entry falls back to
// the Name bytes, so both point to the same text in the rebuilt binary. Entries
// are placed by index (gaps stay nil and rebuild as zero entries).
func ImportFromJson(data *MacroDictionaryJsonImport, version int) (map[string]*MacroDictionaryBinaryFile, error) {
	if data == nil {
		return nil, fmt.Errorf("nil import data")
	}

	locSet := make(map[string]struct{})
	for _, chunk := range data.Chunks {
		for _, s := range chunk.Strings {
			for loc := range s.Name {
				locSet[loc] = struct{}{}
			}
			for loc := range s.SimplifiedName {
				locSet[loc] = struct{}{}
			}
		}
	}
	locKeys := make([]string, 0, len(locSet))
	for loc := range locSet {
		locKeys = append(locKeys, loc)
	}
	sort.Strings(locKeys)

	filesByLoc := make(map[string][]*MacroDictionaryTextFile, len(locKeys))
	for _, loc := range locKeys {
		files := make([]*MacroDictionaryTextFile, MacroDictionaryChunkCount)
		charset := ffxencoding.GetCharsetForLanguage(loc)
		for i := range files {
			files[i] = &MacroDictionaryTextFile{Charset: charset, Version: version, Index: i}
		}
		filesByLoc[loc] = files
	}

	for _, chunk := range data.Chunks {
		if chunk.ChunkIndex < 0 || chunk.ChunkIndex >= MacroDictionaryChunkCount {
			continue
		}
		for _, s := range chunk.Strings {
			if s.Index < 0 {
				continue
			}
			gameVersion := models.GameVersion(version)
			for _, loc := range locKeys {
				charset := ffxencoding.GetCharsetForLanguage(loc)
				nameBytes := converter.StringToBytes(s.Name[loc], charset, gameVersion)
				seg := &MacroDictionaryTextSegment{NameBytes: nameBytes}
				if simpText, ok := s.SimplifiedName[loc]; ok && simpText != "" {
					seg.SimplifiedNameBytes = converter.StringToBytes(simpText, charset, gameVersion)
				} else {
					seg.SimplifiedNameBytes = nameBytes
				}
				segs := filesByLoc[loc][chunk.ChunkIndex].Segments
				for len(segs) <= s.Index {
					segs = append(segs, nil)
				}
				segs[s.Index] = seg
				filesByLoc[loc][chunk.ChunkIndex].Segments = segs
			}
		}
	}

	result := make(map[string]*MacroDictionaryBinaryFile, len(locKeys))
	for _, loc := range locKeys {
		raw, err := BuildContainerBinary(filesByLoc[loc])
		if err != nil {
			return nil, fmt.Errorf("failed to rebuild macro dictionary for localization %s: %w", loc, err)
		}
		c, err := NewMacroDictionaryBinaryFileFromBytes(raw, loc, version)
		if err != nil {
			return nil, fmt.Errorf("failed to parse rebuilt macro dictionary for localization %s: %w", loc, err)
		}
		result[loc] = c
	}
	return result, nil
}

// UnmarshalJson parses merged macro dictionary JSON bytes.
func UnmarshalJson(jsonData []byte) (*MacroDictionaryJsonImport, error) {
	var imp MacroDictionaryJsonImport
	if err := json.Unmarshal(jsonData, &imp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal macro dictionary JSON: %w", err)
	}
	return &imp, nil
}
