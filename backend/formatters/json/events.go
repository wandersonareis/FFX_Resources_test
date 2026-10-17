// Caminho do artefato de events: naming é decisão do CHAMADOR — os
// helpers aqui resolvem o caminho canônico .json; a serialização é do
// JSONFormatter (formatter.go/shared.go) e a gravação da camada de
// arquivo padronizada (formatters.NewWriter).
package json

import (
	"fmt"
	"path/filepath"

	"ffxresources/backend/common"
	"ffxresources/backend/dto"
)

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
