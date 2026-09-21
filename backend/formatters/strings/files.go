package strings

import (
	"fmt"
	stdstrings "strings"

	"ffxresources/backend/common"
	"ffxresources/backend/dto"
	"ffxresources/backend/formatters/json"
)

// Este arquivo espelha os writers do JSON (formatters/json): os arquivos
// .strings nascem ao lado dos .json (mesmo diretório, mesmo basename).
// Os caminhos vêm dos helpers do JSON com a extensão trocada, garantindo
// adjacência sem duplicar a lógica de diretórios.

// StringsFormatter formata a Collection textual para o formato Strings.
//
// Formato java-like, uma linha por (row, idioma):
// <versão>:<id>[:<name>]:<index>:<lang>║$<hash> = <literal|$hash>,
// agrupadas por entrada sob /*key=<key> row_count=<n>*/.
// Cada formato futuro organiza esse DTO à sua maneira, sem reaproveitar
// structs de outro formato.
type StringsFormatter struct{}

// NewStringsFormatter constrói o formatter Strings.
func NewStringsFormatter() StringsFormatter {
	return StringsFormatter{}
}

// Extension retorna a extensão produzida por este formatter.
func (StringsFormatter) Extension() string {
	return extensionStrings
}

// Marshal organiza a Collection pronta no texto de saída (todos os idiomas).
// Recebe apenas DTO pronto; não toca no domínio.
func (StringsFormatter) Marshal(c dto.Collection) ([]byte, error) {
	return Marshal(c)
}

// MarshalLangs organiza a Collection pronta contendo só os idiomas pedidos
// (nil/vazio = todos).
func (StringsFormatter) MarshalLangs(c dto.Collection, langs []string) ([]byte, error) {
	return MarshalLangs(c, langs)
}

// Unmarshal parseia o texto de volta para a Collection (DTO),
// com rows ordenados por Index para reconstrução posicional do binário.
func (StringsFormatter) Unmarshal(data []byte) (dto.Collection, error) {
	return Unmarshal(data)
}

// asStringsPath troca a extensão .json do caminho irmão por .strings.
func asStringsPath(jsonPath string) string {
	return stdstrings.TrimSuffix(jsonPath, ".json") + extensionStrings
}

// EventsStringsPath resolve o caminho do Strings de eventos em edits/,
// ao lado do JSON (mesmo basename).
func EventsStringsPath(c dto.Collection, version common.GameVersion) (string, error) {
	p, err := json.EventsJSONPath(c, version)
	if err != nil {
		return "", err
	}
	return asStringsPath(p), nil
}

// WriteEvents serializa a Collection e escreve o arquivo Strings.
// Recebe apenas DTO pronto e devolve o caminho escrito.
// langs nil/vazio = todos os idiomas.
func (f StringsFormatter) WriteEvents(c dto.Collection, version common.GameVersion, langs []string) (string, error) {
	if len(c) == 0 {
		return "", fmt.Errorf("no events with string data to export")
	}
	raw, err := f.MarshalLangs(c, langs)
	if err != nil {
		return "", err
	}
	filePath, err := EventsStringsPath(c, version)
	if err != nil {
		return "", err
	}
	if err := common.WriteBytesToFile(filePath, raw); err != nil {
		return "", fmt.Errorf("error writing strings file %s: %v", filePath, err)
	}
	common.LogVerbose("Exported event strings file: %s", filePath)
	return filePath, nil
}

// ReadEvents lê o arquivo Strings de eventos e devolve a Collection (DTO).
func (f StringsFormatter) ReadEvents(filePath string) (dto.Collection, error) {
	common.LogVerbose("Loading event strings file: %s", filePath)
	raw, err := common.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load event strings file: %v", err)
	}
	c, err := f.Unmarshal(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to load event strings file: %v", err)
	}
	common.LogVerbose("Successfully loaded %d events from strings", len(c))
	return c, nil
}

// ObjectsStringsPath resolve o caminho do Strings de um arquivo de objetos
// em edits/, ao lado do JSON (mesmo basename).
func ObjectsStringsPath(key string, version common.GameVersion) (string, error) {
	p, err := json.ObjectsJSONPath(key, version)
	if err != nil {
		return "", err
	}
	return asStringsPath(p), nil
}

// WriteObjects serializa a Collection e escreve um arquivo Strings por
// entrada. Recebe apenas DTO pronto e devolve os caminhos escritos.
// langs nil/vazio = todos os idiomas.
func (f StringsFormatter) WriteObjects(c dto.Collection, version common.GameVersion, langs []string) ([]string, error) {
	if len(c) == 0 {
		return nil, fmt.Errorf("no objects with text data to export")
	}
	var paths []string
	for _, key := range c.SortedKeys() {
		singleRaw, err := f.MarshalLangs(dto.Collection{key: c[key]}, langs)
		if err != nil {
			return nil, err
		}
		filePath, err := ObjectsStringsPath(key, version)
		if err != nil {
			return nil, err
		}
		if err := common.WriteBytesToFile(filePath, singleRaw); err != nil {
			return nil, fmt.Errorf("error writing strings file %s: %v", filePath, err)
		}
		common.LogVerbose("Exported objects strings file: %s", filePath)
		paths = append(paths, filePath)
	}
	return paths, nil
}

// ReadObjects lê um arquivo Strings de objetos e devolve a Collection (DTO).
func (f StringsFormatter) ReadObjects(filePath string) (dto.Collection, error) {
	common.LogVerbose("Loading objects strings file: %s", filePath)
	raw, err := common.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load objects strings file: %v", err)
	}
	c, err := f.Unmarshal(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to load objects strings file: %v", err)
	}
	common.LogVerbose("Successfully loaded %d object file(s) from strings", len(c))
	return c, nil
}

// MacroStringsPath resolve o caminho do Strings do dicionário de macros,
// ao lado do JSON (mesmo basename).
func MacroStringsPath(version common.GameVersion) (string, error) {
	p, err := json.MacroJSONPath(version)
	if err != nil {
		return "", err
	}
	return asStringsPath(p), nil
}

// WriteMacro serializa a Collection e escreve o arquivo Strings.
// Recebe apenas DTO pronto e devolve o caminho escrito.
// langs nil/vazio = todos os idiomas.
func (f StringsFormatter) WriteMacro(c dto.Collection, version common.GameVersion, langs []string) (string, error) {
	if len(c) == 0 {
		return "", fmt.Errorf("no macro data in DTO to export")
	}
	filePath, err := MacroStringsPath(version)
	if err != nil {
		return "", err
	}
	return f.WriteMacroFile(c, filePath, langs)
}

// WriteMacroFile serializa a Collection e escreve no caminho dado.
// langs nil/vazio = todos os idiomas.
func (f StringsFormatter) WriteMacroFile(c dto.Collection, filePath string, langs []string) (string, error) {
	if len(c) == 0 {
		return "", fmt.Errorf("no macro data in DTO to export")
	}
	raw, err := f.MarshalLangs(c, langs)
	if err != nil {
		return "", err
	}
	if err := common.WriteBytesToFile(filePath, raw); err != nil {
		return "", fmt.Errorf("error writing strings file %s: %v", filePath, err)
	}
	common.LogVerbose("Exported macro dictionary strings file: %s", filePath)
	return filePath, nil
}

// ReadMacro lê o arquivo Strings do dicionário de macros e devolve a
// Collection (DTO).
func (f StringsFormatter) ReadMacro(filePath string) (dto.Collection, error) {
	common.LogVerbose("Loading macro dictionary strings file: %s", filePath)
	raw, err := common.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load macro dictionary strings file: %v", err)
	}
	c, err := f.Unmarshal(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to load macro dictionary strings file: %v", err)
	}
	common.LogVerbose("Successfully loaded %d macro chunk(s) from strings", len(c))
	return c, nil
}
