// Caminho do artefato de objects: naming é decisão do CHAMADOR — o helper
// aqui resolve o caminho canônico .json; a serialização é do JSONFormatter
// (formatter.go/shared.go) e a gravação da camada de arquivo padronizada
// (formatters.NewWriter).
package json

import (
	"fmt"
	"path/filepath"

	"ffxresources/backend/common"
)

// ObjectsJSONPath resolve o caminho do JSON de um arquivo de objetos em
// edits/: <chave>_all_localizations.json (chave = basename sem extensão).
func ObjectsJSONPath(key string, version common.GameVersion) (string, error) {
	editsPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")
	if err := common.EnsurePathExists(editsPath); err != nil {
		return "", fmt.Errorf("error creating edits directory: %w", err)
	}
	return filepath.Join(editsPath, common.WithVersionSuffixFor(key+"_all_localizations.json", version)), nil
}
