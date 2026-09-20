package formats

import (
	"encoding/json"

	"ffxresources/backend/datastore"
	"ffxresources/backend/models"
)

// JSONObjectFormatter formata os dados originais de objectsfile para JSON,
// organizando a saída como hoje: metadata do binário + strings embutidas,
// envelope {"data": ...}, indentação de 2 espaços e sem escapar HTML.
type JSONObjectFormatter struct{}

// NewJSONObjectFormatter constrói o formatter JSON de objectsfile.
func NewJSONObjectFormatter() JSONObjectFormatter {
	return JSONObjectFormatter{}
}

// Extension retorna a extensão produzida por este formatter.
func (JSONObjectFormatter) Extension() string {
	return extensionJSON
}

// Marshal organiza os dados originais no formato JSON de saída.
func (JSONObjectFormatter) Marshal(data datastore.ObjectTextData) ([]byte, error) {
	stringsBytes, err := json.Marshal(data.Entries)
	if err != nil {
		return nil, err
	}
	export := models.ObjectsFileExport{
		Metadata: data.Metadata,
		Strings:  stringsBytes,
	}
	return marshalNoEscape(models.DataWrapper[models.ObjectsFileExport]{Data: export})
}

// Unmarshal parseia o JSON de volta para os dados originais de objectsfile.
func (JSONObjectFormatter) Unmarshal(data []byte) (datastore.ObjectTextData, error) {
	var out datastore.ObjectTextData
	loaded, err := unmarshalWrapped[models.ObjectsFileExport](data)
	if err != nil {
		return out, err
	}
	var entries []*datastore.ObjectTextEntry
	if err := json.Unmarshal(loaded.Strings, &entries); err != nil {
		return out, err
	}
	out.Entries = entries
	out.Metadata = loaded.Metadata
	return out, nil
}

var _ datastore.IObjectsFormatter = JSONObjectFormatter{}
