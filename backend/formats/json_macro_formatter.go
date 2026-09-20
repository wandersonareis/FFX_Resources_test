package formats

import (
	"encoding/json"
	"fmt"

	"ffxresources/backend/fileFormats/macrodic"
)

// JSONMacroFormatter formata os containers do dicionário de macros para JSON,
// sem envelope e sem escapar HTML (`&` sai literal, não como \u0026).
type JSONMacroFormatter struct{}

// NewJSONMacroFormatter constrói o formatter JSON do dicionário de macros.
func NewJSONMacroFormatter() JSONMacroFormatter {
	return JSONMacroFormatter{}
}

// Extension retorna a extensão produzida por este formatter.
func (JSONMacroFormatter) Extension() string {
	return extensionJSON
}

// Marshal organiza os containers originais no formato JSON de saída.
func (JSONMacroFormatter) Marshal(containers map[string]*macrodic.MacroDictionaryBinaryFile) ([]byte, error) {
	return marshalNoEscape(macrodic.ExportToJson(containers))
}

// Unmarshal desserializa JSON de volta para a representação de importação do
// dicionário de macros.
func (JSONMacroFormatter) Unmarshal(data []byte) (*macrodic.MacroDictionaryJsonImport, error) {
	var imp macrodic.MacroDictionaryJsonImport
	if err := json.Unmarshal(data, &imp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal macro dictionary JSON: %w", err)
	}
	return &imp, nil
}

var _ macrodic.IMacroFormatter = JSONMacroFormatter{}
