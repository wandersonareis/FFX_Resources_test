package macrodic

import (
	"path/filepath"
	"sort"

	"ffxresources/backend/common"
	"ffxresources/backend/models"
	"fmt"
)

// MacroDictionaryJSONFileName is the merged macro dictionary JSON file name,
// holding the texts of every available localization in one document.
const MacroDictionaryJSONFileName = "macro_dictionary_all_localizations.json"

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
func ReadMacroDictionaryContainers(version int) (map[string]*MacroDictionaryTextContainer, error) {
	result := make(map[string]*MacroDictionaryTextContainer)
	for _, loc := range DefaultFirstLocalizations() {
		path := filepath.Join(common.GetLocalizationRoot(loc), "menu", "macrodic.dcp")
		accessor, err := common.NewFileAccessor(path)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve macro dictionary file: %w", err)
		}
		data := accessor.ReadBytes()
		if len(data) == 0 {
			if loc == common.DefaultLocalization {
				return nil, fmt.Errorf("default macro dictionary not found: %s", path)
			}
			common.LogVerbose("Skipping missing macro dictionary for localization: %s", loc)
			continue
		}
		c, err := NewMacroDictionaryTextContainer(data, loc, version)
		if err != nil {
			return nil, fmt.Errorf("failed to parse macro dictionary for localization %s: %w", loc, err)
		}
		result[loc] = c
	}
	return result, nil
}

// PublishMacroDictionaryContainers publishes every given container into the
// datastore, default localization first, so MCR lookups resolve for all
// available languages.
func PublishMacroDictionaryContainers(containers map[string]*MacroDictionaryTextContainer) {
	for _, loc := range DefaultFirstLocalizations() {
		if c, ok := containers[loc]; ok && c != nil {
			c.PublishStrings()
		}
	}
}

// SaveMacroDictionaryBinaries writes each container binary back to its game
// file (mods path per localization), like objectfile SaveBinaryFile does.
func SaveMacroDictionaryBinaries(containers map[string]*MacroDictionaryTextContainer) error {
	for _, loc := range SortedLocalizationKeys(containers) {
		c := containers[loc]
		if c == nil || len(c.Bytes) == 0 {
			continue
		}
		path := models.MacroBinaryPath(loc)
		if err := common.WriteBytesToFile(path, c.Bytes); err != nil {
			return fmt.Errorf("failed to write macro dictionary binary for localization %s: %w", loc, err)
		}
		common.LogVerbose("Wrote macro dictionary binary: %s (%d bytes)", path, len(c.Bytes))
	}
	return nil
}

// SaveMacroDictionaryJson exports all given containers into a single merged JSON
// file holding every available localization, like objectfile exports do.
func SaveMacroDictionaryJson(containers map[string]*MacroDictionaryTextContainer, fileName string) error {
	raw, err := MarshalToJson(ExportToJson(containers))
	if err != nil {
		return fmt.Errorf("failed to marshal macro dictionary JSON: %w", err)
	}
	path := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits", "macrodic", common.WithVersionSuffix(fileName))
	if err := common.WriteBytesToFile(path, raw); err != nil {
		return fmt.Errorf("failed to write macro dictionary JSON file %s: %w", path, err)
	}
	common.LogVerbose("Exported macro dictionary JSON file: %s", path)
	return nil
}

// LoadMacroDictionaryJson reads a merged macro dictionary JSON file.
func LoadMacroDictionaryJson(fileName string) (*MacroDictionaryJsonImport, error) {
	path := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits", "macrodic", common.WithVersionSuffix(fileName))
	if !common.IsPathExists(path) {
		return nil, fmt.Errorf("macro dictionary JSON file not found: %s", path)
	}
	raw, err := common.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read macro dictionary JSON file %s: %w", path, err)
	}
	return UnmarshalJson(raw)
}
