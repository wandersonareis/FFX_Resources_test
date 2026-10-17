package writer

import (
	"ffxresources/backend/builders"
	"ffxresources/backend/common"
	"ffxresources/backend/dto"
	"ffxresources/backend/fileFormats/macrodic"
	"ffxresources/backend/formatters"
	"ffxresources/backend/formatters/json"
	_ "ffxresources/backend/formatters/strings"
	"ffxresources/backend/interactions"
	"fmt"
	"path/filepath"
)

// writeArtifact grava a collection no formato pedido (registry) sobre o
// caminho CANÔNICO .json — a extensão é trocada para a do formato. É a
// camada de arquivo padronizada (formatters.NewWriter) que substitui os
// antigos métodos Write*File por kind dos pacotes de formato.
func writeArtifact(format string, c dto.Collection, jsonPath string, langs []string) (string, error) {
	f, ok := formatters.Get(format)
	if !ok {
		return "", fmt.Errorf("formatter %s não registrado", format)
	}
	return formatters.NewWriter(f, c, formatters.SwapExt(f, jsonPath), langs).Write()
}

// readArtifactFile lê o arquivo no formato pedido (registry) de volta
// para a Collection (DTO).
func readArtifactFile(format, path string) (dto.Collection, error) {
	f, ok := formatters.Get(format)
	if !ok {
		return nil, fmt.Errorf("formatter %s não registrado", format)
	}
	return formatters.NewReader(f, path).Read()
}

// ExportMacroDictionaryToJSON reads every available localization container and
// exports them into a single DTO-based JSON file holding all languages.
func ExportMacroDictionaryToJSON() {
	version := interactions.NewInteractionService().FFXAppConfig().GetGameVersion()
	collection, err := builders.BuildMacroDTO(version, nil)
	if err != nil {
		common.LogError("Error building macro dictionary DTO: %v\n", err)
		return
	}

	if common.IsVerboseMode() {
		for _, key := range collection.SortedKeys() {
			common.LogVerbose("Processing macro dictionary chunk: %s\n", key)
		}
	}

	path, err := json.MacroJSONPath(version)
	if err != nil {
		common.LogError("Error resolving macro dictionary path: %v\n", err)
		return
	}
	if _, err := writeArtifact("json", collection, path, nil); err != nil {
		common.LogError("Error writing macro dictionary JSON: %v\n", err)
		return
	}

	if common.IsVerboseMode() {
		common.LogVerbose("Arquivo JSON de dicionário de macros exportado (%d chunks)\n", len(collection))
	}
}

// ExportMacroDictionaryToStrings reads every available localization container
// and exports them into a single Strings file holding the requested languages
// (langs nil/vazio = todos), ao lado do JSON.
func ExportMacroDictionaryToStrings(langs []string) {
	version := interactions.NewInteractionService().FFXAppConfig().GetGameVersion()
	collection, err := builders.BuildMacroDTO(version, nil)
	if err != nil {
		common.LogError("Error building macro dictionary DTO: %v\n", err)
		return
	}

	path, err := json.MacroJSONPath(version)
	if err != nil {
		common.LogError("Error resolving macro dictionary path: %v\n", err)
		return
	}
	if _, err := writeArtifact("strings", collection, path, langs); err != nil {
		common.LogError("Error writing macro dictionary strings: %v\n", err)
		return
	}

	if common.IsVerboseMode() {
		common.LogVerbose("Arquivo Strings de dicionário de macros exportado (%d chunks)\n", len(collection))
	}
}

// WriteMacroDictionaryForLocalizationJSON exports a single localization into
// its own DTO-based JSON file.
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

	collection, err := builders.BuildMacroDTOFromContainers(version, map[string]*macrodic.MacroDictionaryBinaryFile{localization: c}, nil)
	if err != nil {
		common.LogError("Error building macro dictionary DTO: %v\n", err)
		return
	}

	fileName := fmt.Sprintf("macro_dictionary_%s.json", localization)
	dir := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits", "macrodic")
	if err := common.EnsurePathExists(dir); err != nil {
		common.LogError("Error creating macro edits directory: %v\n", err)
		return
	}
	if _, err := writeArtifact("json", collection, filepath.Join(dir, common.WithVersionSuffix(fileName)), nil); err != nil {
		common.LogError("Error writing macro dictionary JSON: %v\n", err)
		return
	}

	if common.IsVerboseMode() {
		common.LogVerbose("Arquivo JSON de dicionário de macros exportado (%s)\n", localization)
	}
}

// EditAndSaveMacrodicFromJson reads a DTO-based macro dictionary JSON file,
// rebuilds one binary container per localization found in it and saves every
// rebuilt binary back to its game file.
//
// Parameters:
//   - jsonFilePath: Path to the JSON file to read
func EditAndSaveMacrodicFromJson(jsonFilePath string) error {
	if common.IsVerboseMode() {
		common.LogVerbose("Carregando dados do dicionário de macros do arquivo: %s\n", jsonFilePath)
	}
	resolvedFile, err := common.NewFileAccessor(jsonFilePath)
	if err != nil {
		return fmt.Errorf("erro ao resolver caminho do arquivo JSON: %v", err)
	}
	collection, err := readArtifactFile("json", resolvedFile.ResolvedPath)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo JSON: %v", err)
	}

	version := interactions.NewInteractionService().FFXAppConfig().GetGameVersion()
	if err := builders.ApplyMacroDTO(version, collection); err != nil {
		return fmt.Errorf("erro ao importar dados do JSON: %v", err)
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
