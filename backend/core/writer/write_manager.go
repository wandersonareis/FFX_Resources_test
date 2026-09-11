package writer

import (
	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/macrodic"
	"ffxresources/backend/interactions"
	"fmt"
	"path/filepath"
)

// getLocalizationKeys returns all available localization keys
// This function returns the localization keys from the common package
func getLocalizationKeys() []string {
	var keys []string
	for key := range common.SupportedLanguages {
		keys = append(keys, key)
	}
	return keys
}

// ExportMacroDictionaryToJSON reads every available localization container and
// exports them into a single merged JSON file holding all languages, like
// objectfile exports do.
func ExportMacroDictionaryToJSON() {
	version := interactions.NewInteractionService().FFXAppConfig().GetGameVersion()
	containers, err := macrodic.ReadMacroDictionaryContainers(version)
	if err != nil {
		common.LogError("Error reading macro dictionary containers: %v\n", err)
		return
	}

	if len(containers) == 0 {
		common.LogError("No macro dictionary data found to export to JSON")
		return
	}

	if common.IsVerboseMode() {
		for _, loc := range macrodic.SortedLocalizationKeys(containers) {
			common.LogVerbose("Processing macro dictionary for localization: %s\n", loc)
		}
	}

	if err := macrodic.SaveMacroDictionaryJson(containers, macrodic.MacroDictionaryJSONFileName); err != nil {
		common.LogError("Error writing macro dictionary JSON: %v\n", err)
		return
	}

	if common.IsVerboseMode() {
		common.LogVerbose("Arquivo JSON de dicionário de macros exportado (%d localizações)\n", len(containers))
	}
}

// WriteMacroDictionaryForLocalizationJSON exports a single localization into
// its own merged-shape JSON file.
func WriteMacroDictionaryForLocalizationJSON(localization string) {
	version := interactions.NewInteractionService().FFXAppConfig().GetGameVersion()
	containers, err := macrodic.ReadMacroDictionaryContainers(version)
	if err != nil {
		common.LogError("Error reading macro dictionary containers: %v\n", err)
		return
	}

	c, ok := containers[localization]
	if !ok || c == nil {
		common.LogError("No macro data found for localization: %s\n", localization)
		return
	}

	fileName := fmt.Sprintf("macro_dictionary_%s.json", localization)
	if err := macrodic.SaveMacroDictionaryJson(map[string]*macrodic.MacroDictionaryTextContainer{localization: c}, fileName); err != nil {
		common.LogError("Error writing macro dictionary JSON: %v\n", err)
		return
	}

	if common.IsVerboseMode() {
		common.LogVerbose("Arquivo JSON de dicionário de macros exportado (%s)\n", localization)
	}
}

// EditAndSaveMacrodicFromJson reads a merged macro dictionary JSON file,
// rebuilds one binary container per localization found in it and saves every
// rebuilt binary back to its game file, like objectfile save flows do.
//
// Parameters:
//   - jsonFilePath: Path to the merged JSON file to read
func EditAndSaveMacrodicFromJson(jsonFilePath string) error {
	if common.IsVerboseMode() {
		common.LogVerbose("Carregando dados do dicionário de macros do arquivo: %s\n", jsonFilePath)
	}
	resolvedFile, err := common.NewFileAccessor(jsonFilePath)
	if err != nil {
		return fmt.Errorf("erro ao resolver caminho do arquivo JSON: %v", err)
	}
	raw, err := common.ReadFile(resolvedFile.ResolvedPath)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo JSON: %v", err)
	}
	imp, err := macrodic.UnmarshalJson(raw)
	if err != nil {
		return fmt.Errorf("erro ao fazer parse do JSON: %v", err)
	}

	version := interactions.NewInteractionService().FFXAppConfig().GetGameVersion()
	containers, err := macrodic.ImportFromJson(imp, version)
	if err != nil {
		return fmt.Errorf("erro ao importar dados do JSON: %v", err)
	}

	if err := macrodic.SaveMacroDictionaryBinaries(containers); err != nil {
		return fmt.Errorf("erro ao salvar binários: %v", err)
	}

	if common.IsVerboseMode() {
		common.LogVerbose("Dicionário de macros reconstruído com sucesso!\n")
	}
	return nil
}

// ExampleMacroDictionaryUsage demonstrates how to use the macro dictionary
// container-based export functions.
func ExampleMacroDictionaryUsage() {
	common.LogInfo("=== Exemplo de Exportação de Dicionário de Macros ===")
	common.LogInfo("\n1. Exportando dicionário de macros para todas as localizações:")
	ExportMacroDictionaryToJSON()

	common.LogInfo("\n2. Exportando dicionário de macros para localização japonesa:")
	WriteMacroDictionaryForLocalizationJSON("jp")

	common.LogInfo("\n3. Exportando dicionário de macros para localização inglesa:")
	WriteMacroDictionaryForLocalizationJSON("us")

	common.LogInfo("\n=== Exportação de dicionário de macros concluída ===")
}

// MacroDictionaryEditsPath returns the edits/macrodic directory path.
func MacroDictionaryEditsPath() string {
	return filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits", "macrodic")
}
