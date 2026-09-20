package json

import (
	"fmt"
	"path/filepath"

	"ffxresources/backend/common"
	"ffxresources/backend/dto"
)

// JSONObjectFormatter formata a Collection textual de objectsfile para JSON.
//
// Formato: mapa raiz chaveado pelo nome do arquivo sem extensão, cada
// entrada com metadata (mesmos campos do gerador legado
// models.NewObjectFileMetadataKeyed) e rows [{index, name, hash, text}],
// onde index = posição do objeto na lista e name = chave do segmento
// ("name", "ability1", "T"/"sT" de weapons, ...).
type JSONObjectFormatter struct{}

// NewJSONObjectFormatter constrói o formatter JSON de objectsfile.
func NewJSONObjectFormatter() JSONObjectFormatter {
	return JSONObjectFormatter{}
}

// Extension retorna a extensão produzida por este formatter.
func (JSONObjectFormatter) Extension() string {
	return extensionJSON
}

// Marshal organiza a Collection pronta no JSON de saída.
// Recebe apenas DTO pronto; não toca no domínio.
func (JSONObjectFormatter) Marshal(c dto.Collection) ([]byte, error) {
	return marshalCollection(c)
}

// Unmarshal parseia o JSON de volta para a Collection (DTO),
// com rows ordenados por Index para reconstrução posicional.
func (JSONObjectFormatter) Unmarshal(data []byte) (dto.Collection, error) {
	return unmarshalCollection(data)
}

// ObjectsJSONPath resolve o caminho do JSON de um arquivo de objetos em
// edits/: <chave>_all_localizations.json (chave = basename sem extensão).
func ObjectsJSONPath(key string, version common.GameVersion) (string, error) {
	editsPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")
	if err := common.EnsurePathExists(editsPath); err != nil {
		return "", fmt.Errorf("error creating edits directory: %w", err)
	}
	return filepath.Join(editsPath, common.WithVersionSuffixFor(key+"_all_localizations.json", version)), nil
}

// WriteObjects serializa a Collection e escreve um arquivo JSON por entrada.
// Recebe apenas DTO pronto e devolve os caminhos escritos.
func (f JSONObjectFormatter) WriteObjects(c dto.Collection, version common.GameVersion) ([]string, error) {
	if len(c) == 0 {
		return nil, fmt.Errorf("no objects with text data to export")
	}
	var paths []string
	for _, key := range c.SortedKeys() {
		singleRaw, err := f.Marshal(dto.Collection{key: c[key]})
		if err != nil {
			return nil, err
		}
		filePath, err := ObjectsJSONPath(key, version)
		if err != nil {
			return nil, err
		}
		if err := common.WriteBytesToFile(filePath, singleRaw); err != nil {
			return nil, fmt.Errorf("error writing JSON file %s: %w", filePath, err)
		}
		common.LogVerbose("Exported objects JSON file: %s", filePath)
		paths = append(paths, filePath)
	}
	common.LogVerbose("Total object files exported: %d", len(paths))
	return paths, nil
}

// ReadObjects lê um arquivo JSON de objetos e devolve a Collection (DTO).
// Outro responsável (applier) aplica o DTO de volta nos objetos/binário.
func (f JSONObjectFormatter) ReadObjects(filePath string) (dto.Collection, error) {
	common.LogVerbose("Loading objects JSON file: %s", filePath)
	raw, err := common.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load objects JSON file: %w", err)
	}
	c, err := f.Unmarshal(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to load objects JSON file: %w", err)
	}
	common.LogVerbose("Successfully loaded %d object file(s) from JSON", len(c))
	return c, nil
}
