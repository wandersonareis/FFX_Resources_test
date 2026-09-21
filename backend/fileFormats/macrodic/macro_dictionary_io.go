package macrodic

import (
	"path/filepath"
	"sort"

	"ffxresources/backend/common"
	ffxencoding "ffxresources/backend/core/encoding"
	"fmt"
)

// DefaultFirstLocalizations returns the default localization followed by the
// remaining supported localizations in sorted order, so the default is always
// read first and the others populate afterwards.
func DefaultFirstLocalizations() []string {
	keys := make([]string, 0, len(common.SupportedLanguages))
	for loc := range common.SupportedLanguages {
		if loc == common.DefaultLocalization {
			continue
		}
		keys = append(keys, loc)
	}
	sort.Strings(keys)
	return append([]string{common.DefaultLocalization}, keys...)
}

// ReadMacroDictionaryContainers reads the macro dictionary binary of the default
// localization first, then populates with the other available localizations.
// A missing default file is an error; other missing files are skipped.
func ReadMacroDictionaryContainers(version common.GameVersion) (map[string]*MacroDictionaryBinaryFile, error) {
	if err := ffxencoding.EnsureAllCharsetsLoaded(version); err != nil {
		return nil, fmt.Errorf("charset maps not loaded: %w", err)
	}
	result := make(map[string]*MacroDictionaryBinaryFile)
	for _, loc := range DefaultFirstLocalizations() {
		path := filepath.Join(common.GetLocalizationRootForVersion(version, loc), "menu", "macrodic.dcp")
		accessor, err := common.NewFileAccessor(path)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve macro dictionary file: %w", err)
		}
		data, err := accessor.ReadBytes()
		if err != nil {
			return nil, fmt.Errorf("failed to read macro dictionary file: %w", err)
		}
		if len(data) == 0 {
			if loc == common.DefaultLocalization {
				return nil, fmt.Errorf("default macro dictionary not found: %s", path)
			}
			common.LogVerbose("Skipping missing macro dictionary for localization: %s", loc)
			continue
		}
		c, err := NewMacroDictionaryBinaryFileFromBytes(data, loc, version)
		if err != nil {
			return nil, fmt.Errorf("failed to parse macro dictionary for localization %s: %w", loc, err)
		}
		result[loc] = c
	}
	return result, nil
}

// SortedLocalizationKeys returns the sorted localization keys of the given
// containers for deterministic export output.
func SortedLocalizationKeys(containers map[string]*MacroDictionaryBinaryFile) []string {
	keys := make([]string, 0, len(containers))
	for loc := range containers {
		keys = append(keys, loc)
	}
	sort.Strings(keys)
	return keys
}

// PublishMacroDictionaryContainers publishes every given container into the
// datastore, default localization first, so MCR lookups resolve for all
// available languages.
func PublishMacroDictionaryContainers(containers map[string]*MacroDictionaryBinaryFile) {
	for _, loc := range DefaultFirstLocalizations() {
		if c, ok := containers[loc]; ok && c != nil {
			_ = c.PublishStrings()
		}
	}
}

// SaveMacroDictionaryBinaries writes each container binary back to its game
// file (mods path per localization), like objectfile SaveBinaryFile does.
func SaveMacroDictionaryBinaries(containers map[string]*MacroDictionaryBinaryFile) error {
	for _, loc := range SortedLocalizationKeys(containers) {
		c := containers[loc]
		if c == nil || len(c.Bytes) == 0 {
			continue
		}
		path := common.MacroBinaryPath(loc)
		if err := common.WriteBytesToFile(path, c.Bytes); err != nil {
			return fmt.Errorf("failed to write macro dictionary binary for localization %s: %w", loc, err)
		}
		common.LogVerbose("Wrote macro dictionary binary: %s (%d bytes)", path, len(c.Bytes))
	}
	return nil
}
