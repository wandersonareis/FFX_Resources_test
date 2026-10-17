// Caminho do artefato do dicionário de macros: naming é decisão do
// CHAMADOR — os helpers aqui resolvem o caminho canônico .json; a
// serialização é do JSONFormatter (formatter.go/shared.go) e a gravação
// da camada de arquivo padronizada (formatters.NewWriter).
package json

import (
	"fmt"
	"path/filepath"

	"ffxresources/backend/common"
)

// MacroJSONFileName é o nome do JSON do dicionário de macros em
// edits/macrodic, com textos de todas as localizações num documento só.
const MacroJSONFileName = "macro_dictionary_all_localizations.json"

// macroEditsPath garante o diretório edits/macrodic.
func macroEditsPath() (string, error) {
	path := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits", "macrodic")
	if err := common.EnsurePathExists(path); err != nil {
		return "", fmt.Errorf("error creating macro edits directory: %w", err)
	}
	return path, nil
}

// MacroJSONPath resolve o caminho do JSON do dicionário de macros.
func MacroJSONPath(version common.GameVersion) (string, error) {
	dir, err := macroEditsPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, common.WithVersionSuffix(MacroJSONFileName)), nil
}

// DefaultMacroJSONPath resolve o caminho do JSON padrão do dicionário.
func DefaultMacroJSONPath(version common.GameVersion) (string, error) {
	dir := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits", "macrodic")
	if !common.IsPathExists(dir) {
		return "", fmt.Errorf("macro edits directory not found: %s", dir)
	}
	jsonFilePath := filepath.Join(dir, common.WithVersionSuffix(MacroJSONFileName))
	if !common.IsPathExists(jsonFilePath) {
		return "", fmt.Errorf("macro dictionary JSON file not found: %s", jsonFilePath)
	}
	return jsonFilePath, nil
}
