// Caminho do artefato de help: naming é decisão do CHAMADOR — o helper
// aqui resolve o caminho canônico .json; a serialização é do JSONFormatter
// (formatter.go/shared.go) e a gravação da camada de arquivo padronizada
// (formatters.NewWriter).
package json

import (
	"fmt"
	"path/filepath"

	"ffxresources/backend/common"
)

// HelpJSONPath resolve o caminho do artefato JSON único de help em
// mods/edits (padrão dos demais kinds): help_all_localizations[_<versão>].json.
func HelpJSONPath(version common.GameVersion) (string, error) {
	editsPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")
	if err := common.EnsurePathExists(editsPath); err != nil {
		return "", fmt.Errorf("error creating edits directory: %w", err)
	}
	return filepath.Join(editsPath, common.WithVersionSuffixFor("help_all_localizations.json", version)), nil
}
