package json

import (
	"fmt"
	"path/filepath"
	"sort"

	"ffxresources/backend/common"
	"ffxresources/backend/dto"
)

// JSONEventsFormatter formata a Collection textual de eventos para JSON.
//
// Formato novo (quebra autorizada): mapa raiz chaveado pelo nome do arquivo
// sem extensão (= eventID), cada entrada com metadata (mesmos campos do
// gerador legado models.NewEventFileInfo) e rows [{index, hash, text}].
// Cada formato futuro organiza esse DTO à sua maneira, sem reaproveitar
// structs de outro formato.
type JSONEventsFormatter struct{}

// NewJSONEventsFormatter constrói o formatter JSON de eventos.
func NewJSONEventsFormatter() JSONEventsFormatter {
	return JSONEventsFormatter{}
}

// Extension retorna a extensão produzida por este formatter.
func (JSONEventsFormatter) Extension() string {
	return extensionJSON
}

// Marshal organiza a Collection pronta no JSON de saída.
// Recebe apenas DTO pronto; não toca no domínio.
func (JSONEventsFormatter) Marshal(c dto.Collection) ([]byte, error) {
	return marshalCollection(c)
}

// Unmarshal parseia o JSON de volta para a Collection (DTO),
// com rows ordenados por Index para reconstrução posicional do binário.
func (JSONEventsFormatter) Unmarshal(data []byte) (dto.Collection, error) {
	return unmarshalCollection(data)
}

// EventsJSONPath resolve o caminho do JSON de eventos em edits/.
// Uma entrada → event_<id>_all_localizations.json; várias → events_all_localizations.json.
func EventsJSONPath(c dto.Collection, version common.GameVersion) (string, error) {
	editsPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")
	if err := common.EnsurePathExists(editsPath); err != nil {
		return "", fmt.Errorf("error creating edits directory: %w", err)
	}
	keys := c.SortedKeys()
	var fileName string
	if len(keys) == 1 {
		fileName = "event_" + keys[0] + "_all_localizations.json"
	} else {
		fileName = "events_all_localizations.json"
	}
	return filepath.Join(editsPath, common.WithVersionSuffixFor(fileName, version)), nil
}

// WriteEvents serializa a Collection e escreve o arquivo JSON.
// Recebe apenas DTO pronto e devolve o caminho escrito.
func (f JSONEventsFormatter) WriteEvents(c dto.Collection, version common.GameVersion) (string, error) {
	if len(c) == 0 {
		return "", fmt.Errorf("no events with string data to export")
	}
	raw, err := f.Marshal(c)
	if err != nil {
		return "", err
	}
	filePath, err := EventsJSONPath(c, version)
	if err != nil {
		return "", err
	}
	if err := common.WriteBytesToFile(filePath, raw); err != nil {
		return "", fmt.Errorf("error writing JSON file %s: %w", filePath, err)
	}
	common.LogVerbose("Exported event JSON file: %s", filePath)
	common.LogVerbose("Total events exported: %d", len(c))
	return filePath, nil
}

// ReadEvents lê o arquivo JSON de eventos e devolve a Collection (DTO).
// Outro responsável (applier) aplica o DTO de volta no binário.
func (f JSONEventsFormatter) ReadEvents(filePath string) (dto.Collection, error) {
	common.LogVerbose("Loading events JSON file: %s", filePath)
	raw, err := common.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load events JSON file: %w", err)
	}
	c, err := f.Unmarshal(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to load events JSON file: %w", err)
	}
	common.LogVerbose("Successfully loaded %d events from JSON", len(c))
	return c, nil
}

// DefaultEventsJSONPath resolve o caminho do bulk events_all_localizations.json.
func DefaultEventsJSONPath(version common.GameVersion) (string, error) {
	editsPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")
	if !common.IsPathExists(editsPath) {
		return "", fmt.Errorf("edits directory not found: %s", editsPath)
	}
	jsonFilePath := filepath.Join(editsPath, common.WithVersionSuffixFor("events_all_localizations.json", version))
	if !common.IsPathExists(jsonFilePath) {
		return "", fmt.Errorf("events JSON file not found: %s", jsonFilePath)
	}
	return jsonFilePath, nil
}

// SortedKeysOf devolve as chaves ordenadas (atalho determinístico p/ logs).
func SortedKeysOf(c dto.Collection) []string {
	keys := c.SortedKeys()
	out := make([]string, len(keys))
	copy(out, keys)
	sort.Strings(out)
	return out
}
