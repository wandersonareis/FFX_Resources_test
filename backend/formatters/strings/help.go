package strings

import (
	"fmt"

	"ffxresources/backend/common"
	"ffxresources/backend/dto"
	"ffxresources/backend/formatters/json"
)

// HelpStrings: espelho do JSON (formatters/json/help.go) — o artefato
// .strings de help é único por versão e nasce ao lado do .json irmão em
// mods/edits (mesmo basename, extensão trocada), com todos os painéis
// diferenciados pela metadata.key (nome dos arquivos .sps2).

// HelpStringsPath resolve o caminho do Strings único de help, ao lado do JSON.
func HelpStringsPath(version common.GameVersion) (string, error) {
	p, err := json.HelpJSONPath(version)
	if err != nil {
		return "", err
	}
	return asStringsPath(p), nil
}

// WriteHelp serializa a Collection e escreve o artefato Strings único de
// help (irmão do JSON, mesmo basename). Devolve o caminho escrito.
func (f StringsFormatter) WriteHelp(c dto.Collection, version common.GameVersion, langs []string) ([]string, error) {
	if len(c) == 0 {
		return nil, fmt.Errorf("no help panels with string data to export")
	}
	filePath, err := HelpStringsPath(version)
	if err != nil {
		return nil, err
	}
	raw, err := f.MarshalLangs(c, langs)
	if err != nil {
		return nil, err
	}
	if err := common.WriteBytesToFile(filePath, raw); err != nil {
		return nil, fmt.Errorf("error writing strings file %s: %v", filePath, err)
	}
	common.LogVerbose("Exported help strings file: %s", filePath)
	return []string{filePath}, nil
}

// ReadHelp lê o artefato Strings de help e devolve a Collection (DTO).
func (f StringsFormatter) ReadHelp(filePath string) (dto.Collection, error) {
	common.LogVerbose("Loading help strings file: %s", filePath)
	raw, err := common.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load help strings file: %v", err)
	}
	c, err := f.Unmarshal(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to load help strings file: %v", err)
	}
	return c, nil
}
