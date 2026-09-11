package reader

import (
	"ffxresources/backend/common"
	"ffxresources/backend/datastore"
	"ffxresources/backend/fileFormats/macrodic"
	"ffxresources/backend/interactions"
	"ffxresources/backend/models"
	"path/filepath"
)

// PrepareStringMacros reads one localization macro dictionary binary into a
// container and publishes its strings into the datastore. Missing files are
// skipped so unavailable localizations never break initialization.
func PrepareStringMacros(filename, localization string, version int) error {
	resolvedFile, err := common.NewFileAccessor(filename)
	if err != nil {
		return err
	}
	data := resolvedFile.ReadBytes()
	if len(data) == 0 {
		common.LogVerbose("Skipping missing macro dictionary file: %s", filename)
		return nil
	}
	c, err := macrodic.NewMacroDictionaryTextContainer(data, localization, version)
	if err != nil {
		common.LogVerbose("Skipping unreadable macro dictionary file %s: %v", filename, err)
		return nil
	}
	c.PublishStrings()
	return nil
}

func InitializeInternals() error {
	gameVersion := interactions.CurrentGameVersion()
	version := int(gameVersion)
	for _, cs := range common.Charsets {
		if err := PrepareCharset(gameVersion, cs); err != nil {
			return err
		}
	}

	// Default localization first, then populate with the other available ones.
	for _, loc := range macrodic.DefaultFirstLocalizations() {
		path := filepath.Join(common.GetLocalizationRoot(loc), "menu", "macrodic.dcp")
		if err := PrepareStringMacros(path, loc, version); err != nil {
			return err
		}
	}

	common.LogVerbose("Macro Lookup Table:\n")
	logMacroLookup(gameVersion)

	return nil
}

// InitializeAllInternals carrega charsets das duas versões sem mexer no
// datastore de macros (útil quando FFX e FFX-2 precisam coexistir).
func InitializeAllInternals() error {
	for _, cs := range common.Charsets {
		if err := PrepareAllCharsets(cs); err != nil {
			return err
		}
	}
	return nil
}

func logMacroLookup(gameVersion models.GameVersion) {
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
