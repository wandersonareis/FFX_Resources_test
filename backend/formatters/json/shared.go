// Package json implementa o formato JSON usado na exportação/importação
// dos arquivos de localização. É o único ponto público do formato JSON:
// os pacotes de domínio (event, objectsfile, macrodic) desconhecem JSON.
//
// Os formatters recebem apenas DTO pronto (backend/dto) e devolvem DTO:
// quem monta o DTO é backend/builders e quem aplica o DTO de volta no
// binário é o applier do domínio.
package json

import (
	"bytes"
	"encoding/json"

	"ffxresources/backend/dto"
)

// extensionJSON é a extensão produzida pelos formatters JSON.
const extensionJSON = ".json"

// ExtensionJSON retorna a extensão produzida pelos formatters JSON.
func ExtensionJSON() string {
	return extensionJSON
}

// marshalNoEscape serializa v com indentação de 2 espaços e sem escapar HTML
// (`&`, `<`, `>` saem literais).
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

// marshalCollection serializa a Collection com chaves e rows ordenadas
// (saída determinística). Recebe apenas DTO pronto.
func marshalCollection(c dto.Collection) ([]byte, error) {
	ordered := make(map[string]dto.FileEntry, len(c))
	for _, k := range c.SortedKeys() {
		entry := c[k]
		rows := make([]dto.TextRow, len(entry.Rows))
		copy(rows, entry.Rows)
		dto.SortRows(rows)
		entry.Rows = rows
		ordered[k] = entry
	}
	return marshalNoEscape(ordered)
}

// unmarshalCollection parseia o JSON de volta para a Collection (DTO),
// com rows ordenados por Index para reconstrução posicional do binário.
func unmarshalCollection(data []byte) (dto.Collection, error) {
	var c dto.Collection
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	for k, entry := range c {
		dto.SortRows(entry.Rows)
		c[k] = entry
	}
	return c, nil
}
