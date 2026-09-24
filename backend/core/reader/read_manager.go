package reader

import (
	"ffxresources/backend/common"
	"ffxresources/backend/datastore"
	"ffxresources/backend/fileFormats/macrodic"
	"ffxresources/backend/interactions"
	"path/filepath"
)

// PrepareStringMacros reads one localization macro dictionary binary into a
// container and publishes its strings into the datastore. Missing files are
// skipped so unavailable localizations never break initialization.
func PrepareStringMacros(filename, localization string, version common.GameVersion) error {
	c := macrodic.NewMacroDictionaryBinaryFile(localization, version)
	if err := c.LoadFromBinary(); err != nil {
		common.LogVerbose("Skipping missing macro dictionary file: %s (%v)", filename, err)
		return nil
	}
	if err := c.PublishStrings(); err != nil {
		common.LogVerbose("Skipping unreadable macro dictionary file %s: %v", filename, err)
		return nil
	}
	return nil
}

func InitializeInternals() error {
	gameVersion := interactions.CurrentGameVersion()
	if err := PrepareVersion(gameVersion); err != nil {
		return err
	}

	common.LogVerbose("Macro Lookup Table:\n")
	logMacroLookup(gameVersion)

	return nil
}

// PrepareVersion carrega os charsets da versão e publica os macros no
// datastore. É idempotente do ponto de vista dos dados (sobrescreve os mapas)
// e independente da versão global ativa, para servir FFX/FFX-2/LastMiss.
func PrepareVersion(version common.GameVersion) error {
	for _, cs := range common.Charsets {
		if err := PrepareCharset(version, cs); err != nil {
			return err
		}
	}

	// Default localization first, then populate with the other available ones.
	for _, loc := range macrodic.DefaultFirstLocalizations() {
		path := filepath.Join(common.GetLocalizationRootForVersion(version, loc), "menu", "macrodic.dcp")
		if err := PrepareStringMacros(path, loc, version); err != nil {
			return err
		}
	}

	return nil
}

func logMacroLookup(gameVersion common.GameVersion) {
	if common.IsVerboseMode() {
		datastore.GetMacros(gameVersion).ForEach(func(_ int, strings datastore.IGlobalLocalizedMacroStringObject) {
			for loc := range common.SupportedLanguages {
				content, _ := strings.GetLocalizedContent(loc)
				if content == nil || content.IsEmpty() {
					continue
				}
				macro := content.GetString()
				common.LogVerbose("  %s: %s\n", loc, macro)
			}
		})
	}
}
