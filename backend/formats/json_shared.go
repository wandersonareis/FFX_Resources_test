// Package formats implementa os formatos de texto (handlers) usados na
// exportação/importação dos arquivos de localização. Cada formato é uma
// struct que implementa o contrato tipado do seu domínio:
//
//   - JSONEventsFormatter  -> event.IEventsFormatter
//   - JSONObjectFormatter  -> datastore.IObjectsFormatter
//   - JSONMacroFormatter   -> macrodic.IMacroFormatter
//
// Um formato novo (ex: YAML) recebe as mesmas representações intermediárias
// e implementa os mesmos contratos; nenhum código de domínio precisa mudar.
package formats

import (
	"bytes"
	"encoding/json"

	"ffxresources/backend/models"
)

// extensionJSON é a extensão produzida pelos formatters JSON.
const extensionJSON = ".json"

// marshalNoEscape serializa v com indentação de 2 espaços e sem escapar HTML
// (`&`, `<`, `>` saem literais). É o comportamento do antigo saveNoEscape de
// models, preservado byte a byte nos arquivos gerados.
func marshalNoEscape(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// unmarshalWrapped desserializa um documento {"data": payload}, com fallback
// para payload direto (formato legado). É o comportamento do antigo
// LoadDataFile de models, agora parametrizado pelo tipo do payload.
func unmarshalWrapped[T any](data []byte) (T, error) {
	var zero T
	var probe map[string]any
	if err := json.Unmarshal(data, &probe); err == nil {
		if _, ok := probe["data"]; ok {
			var wrapped models.DataWrapper[T]
			if err := json.Unmarshal(data, &wrapped); err == nil {
				return wrapped.Data, nil
			}
		}
	}
	if err := json.Unmarshal(data, &zero); err != nil {
		return zero, err
	}
	return zero, nil
}
