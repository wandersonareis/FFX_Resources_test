package json

import (
	"fmt"
	"path/filepath"

	"ffxresources/backend/common"
	"ffxresources/backend/dto"
)

// MacroJSONFileName é o nome do JSON do dicionário de macros em
// edits/macrodic, com textos de todas as localizações num documento só.
const MacroJSONFileName = "macro_dictionary_all_localizations.json"

// JSONMacroFormatter formata a Collection textual do dicionário de macros
// para JSON.
//
// Formato: mapa raiz chaveado por chunk ("chunk_NN"), cada entrada com
// metadata (ChunkIndex para reconstrução posicional) e rows
// [{index, name, hash, text}], onde name é "name"/"simplifiedName".
type JSONMacroFormatter struct{}

// NewJSONMacroFormatter constrói o formatter JSON do dicionário de macros.
func NewJSONMacroFormatter() JSONMacroFormatter {
	return JSONMacroFormatter{}
}

// Extension retorna a extensão produzida por este formatter.
func (JSONMacroFormatter) Extension() string {
	return extensionJSON
}

// Marshal organiza a Collection pronta no JSON de saída.
// Recebe apenas DTO pronto; não toca no domínio.
func (JSONMacroFormatter) Marshal(c dto.Collection) ([]byte, error) {
	return marshalCollection(c)
}

// Unmarshal parseia o JSON de volta para a Collection (DTO),
// com rows ordenados por Index para reconstrução posicional do binário.
func (JSONMacroFormatter) Unmarshal(data []byte) (dto.Collection, error) {
	return unmarshalCollection(data)
}

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

// WriteMacro serializa a Collection e escreve o arquivo JSON.
// Recebe apenas DTO pronto e devolve o caminho escrito.
func (f JSONMacroFormatter) WriteMacro(c dto.Collection, version common.GameVersion) (string, error) {
	if len(c) == 0 {
		return "", fmt.Errorf("no macro data in DTO to export")
	}
	filePath, err := MacroJSONPath(version)
	if err != nil {
		return "", err
	}
	return f.WriteMacroFile(c, filePath)
}

// WriteMacroFile serializa a Collection e escreve no caminho dado.
// Recebe apenas DTO pronto e devolve o caminho escrito.
func (f JSONMacroFormatter) WriteMacroFile(c dto.Collection, filePath string) (string, error) {
	if len(c) == 0 {
		return "", fmt.Errorf("no macro data in DTO to export")
	}
	raw, err := f.Marshal(c)
	if err != nil {
		return "", err
	}
	if err := common.WriteBytesToFile(filePath, raw); err != nil {
		return "", fmt.Errorf("error writing JSON file %s: %w", filePath, err)
	}
	common.LogVerbose("Exported macro dictionary JSON file: %s", filePath)
	return filePath, nil
}

// ReadMacro lê o arquivo JSON do dicionário de macros e devolve a
// Collection (DTO). Outro responsável (applier) reconstrói os binários.
func (f JSONMacroFormatter) ReadMacro(filePath string) (dto.Collection, error) {
	common.LogVerbose("Loading macro dictionary JSON file: %s", filePath)
	raw, err := common.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load macro dictionary JSON file: %w", err)
	}
	c, err := f.Unmarshal(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to load macro dictionary JSON file: %w", err)
	}
	common.LogVerbose("Successfully loaded %d macro chunk(s) from JSON", len(c))
	return c, nil
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
