package json

import (
	"fmt"
	"path/filepath"

	"ffxresources/backend/common"
	"ffxresources/backend/dto"
)

// JSONHelpFormatter formata a Collection textual dos painéis de ajuda
// (kind help) para JSON. O artefato segue o padrão dos demais formatos
// (gravado em mods/edits) e é ÚNICO por versão: todos os 6 painéis no mesmo
// arquivo (help_all_localizations_ffx.json), com as entradas diferenciadas
// pela metadata.key (nome dos arquivos .sps2).
//
// Os textos gêmeos deduplicam entre entradas (marshalCollectionLangs
// compartilha o mapa $hash) e o import reconstrói cada binário separadamente.
type JSONHelpFormatter struct{}

func NewJSONHelpFormatter() JSONHelpFormatter {
	return JSONHelpFormatter{}
}

func (JSONHelpFormatter) Extension() string {
	return extensionJSON
}

func (JSONHelpFormatter) Marshal(c dto.Collection) ([]byte, error) {
	return marshalCollection(c)
}

func (JSONHelpFormatter) MarshalLangs(c dto.Collection, langs []string) ([]byte, error) {
	return marshalCollectionLangs(c, langs)
}

func (JSONHelpFormatter) Unmarshal(data []byte) (dto.Collection, error) {
	return unmarshalCollection(data)
}

// HelpJSONPath resolve o caminho do artefato JSON único de help em
// mods/edits (padrão dos demais kinds): help_all_localizations[_<versão>].json.
func HelpJSONPath(version common.GameVersion) (string, error) {
	editsPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")
	if err := common.EnsurePathExists(editsPath); err != nil {
		return "", fmt.Errorf("error creating edits directory: %w", err)
	}
	return filepath.Join(editsPath, common.WithVersionSuffixFor("help_all_localizations.json", version)), nil
}

// WriteHelp serializa a Collection e escreve o artefato JSON único de help
// em mods/edits (os painéis da Collection, no padrão dos demais formatos).
// Devolve o caminho escrito.
func (f JSONHelpFormatter) WriteHelp(c dto.Collection, version common.GameVersion, langs []string) ([]string, error) {
	if len(c) == 0 {
		return nil, fmt.Errorf("no help panels with string data to export")
	}
	filePath, err := HelpJSONPath(version)
	if err != nil {
		return nil, err
	}
	raw, err := f.MarshalLangs(c, langs)
	if err != nil {
		return nil, err
	}
	if err := common.WriteBytesToFile(filePath, raw); err != nil {
		return nil, fmt.Errorf("error writing JSON file %s: %w", filePath, err)
	}
	common.LogVerbose("Exported help JSON file: %s", filePath)
	return []string{filePath}, nil
}

// ReadHelp lê o artefato JSON de help e devolve a Collection (DTO),
// com refs $hash expandidas (unmarshalCollection).
func (f JSONHelpFormatter) ReadHelp(filePath string) (dto.Collection, error) {
	data, err := common.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return f.Unmarshal(data)
}
