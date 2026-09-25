package reader

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/encoding"
	"ffxresources/backend/datastore"
	"ffxresources/backend/fileFormats/macrodic"
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

// InitializeInternals prepara os internals da versão indicada: charsets
// (embutidos, via core/encoding) e macros das localizações. A versão é
// recebida do chamador — o reader não consulta a versão global, ficando sem
// acoplamento a interactions e testável por versão sem instâncias.
func InitializeInternals(version common.GameVersion) error {
	if err := PrepareVersion(version); err != nil {
		return err
	}

	common.LogVerbose("Macro Lookup Table:\n")
	logMacroLookup(version)

	return nil
}

// PrepareVersion publica os charsets da versão (delegados ao core/encoding:
// tabelas embutidas, sem I/O) e carrega os macros das localizações para o
// datastore. É idempotente do ponto de vista dos dados (sobrescreve os mapas)
// e independente da versão global ativa, para servir FFX/FFX-2/LastMiss.
// O reader fica com o I/O legítimo: ler os macrodic.dcp — a carga dos charsets
// nunca falha por recurso ausente.
func PrepareVersion(version common.GameVersion) error {
	if err := ffxencoding.PrepareVersionCharsets(version); err != nil {
		return err
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
