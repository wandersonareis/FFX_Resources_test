// Caminhos de artefato por escopo de export: o dedup do marshal é por
// arquivo (mapa `seen` local à serialização), então o NOME do artefato
// precisa refletir o escopo do pedido — um subconjunto nunca pode colidir
// com o bulk completo (sobrescreveria o arquivo completo do tradutor).
package services

import (
	"fmt"
	"path/filepath"

	"ffxresources/backend/common"
	jsonfmt "ffxresources/backend/formatters/json"
)

// eventsExportPath resolve o caminho do artefato JSON de events:
//   - ids vazio (tudo)     → events_all_localizations[_<v>].json
//   - 1 entrada            → event_<id>_all_localizations[_<v>].json
//   - subconjunto          → events_sel_<n>_<id1>_all_localizations[_<v>].json
func eventsExportPath(version common.GameVersion, ids []string) (string, error) {
	editsPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")
	if err := common.EnsurePathExists(editsPath); err != nil {
		return "", fmt.Errorf("error creating edits directory: %w", err)
	}
	var name string
	switch len(ids) {
	case 0:
		name = "events_all_localizations.json"
	case 1:
		name = "event_" + ids[0] + "_all_localizations.json"
	default:
		name = fmt.Sprintf("events_sel_%d_%s_all_localizations.json", len(ids), ids[0])
	}
	return filepath.Join(editsPath, common.WithVersionSuffixFor(name, version)), nil
}

// objectsBulkPath resolve o caminho do artefato JSON ÚNICO de objects
// (export completo): objects_all_localizations[_<v>].json em edits/.
func objectsBulkPath(version common.GameVersion) (string, error) {
	editsPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")
	if err := common.EnsurePathExists(editsPath); err != nil {
		return "", fmt.Errorf("error creating edits directory: %w", err)
	}
	return filepath.Join(editsPath, common.WithVersionSuffixFor("objects_all_localizations.json", version)), nil
}

// macroExportPath resolve o caminho do artefato JSON do dicionário:
//   - ids vazio (tudo)  → edits/macrodic/macro_dictionary_all_localizations
//     (artefato canônico, mesmo do import padrão);
//   - 1 chunk           → macro_chunk_NN_all_localizations
//   - subconjunto       → macro_sel_<n>_<chunk1>_all_localizations
func macroExportPath(version common.GameVersion, ids []string) (string, error) {
	canonical, err := jsonfmt.MacroJSONPath(version)
	if err != nil {
		return "", err
	}
	var name string
	switch len(ids) {
	case 0:
		name = filepath.Base(canonical)
	case 1:
		name = "macro_" + ids[0] + "_all_localizations.json"
	default:
		name = fmt.Sprintf("macro_sel_%d_%s_all_localizations.json", len(ids), ids[0])
	}
	return filepath.Join(filepath.Dir(canonical), common.WithVersionSuffixFor(name, version)), nil
}
