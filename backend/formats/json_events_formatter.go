package formats

import (
	"sort"

	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/event"
	"ffxresources/backend/models"
)

// JSONEventsFormatter formata o array original de eventos para JSON,
// organizando a saída como hoje: mapa por ID com metadata por evento,
// envelope {"data": ...}, indentação de 2 espaços e sem escapar HTML.
type JSONEventsFormatter struct{}

// NewJSONEventsFormatter constrói o formatter JSON de eventos.
func NewJSONEventsFormatter() JSONEventsFormatter {
	return JSONEventsFormatter{}
}

// Extension retorna a extensão produzida por este formatter.
func (JSONEventsFormatter) Extension() string {
	return extensionJSON
}

// Marshal organiza os eventos originais no formato JSON de saída.
func (JSONEventsFormatter) Marshal(events []event.EventFileData, version common.GameVersion) ([]byte, error) {
	stringsMap := make(map[string]models.EventFileExport, len(events))
	for _, e := range events {
		strings := make([]models.EventStringDataExport, 0, len(e.Strings))
		for _, s := range e.Strings {
			strings = append(strings, models.EventStringDataExport{Index: s.Index, Text: s.Text})
		}
		stringsMap[e.ID] = models.EventFileExport{
			Metadata: models.NewEventFileInfo(e.ID, version),
			ID:       e.ID,
			Strings:  strings,
		}
	}
	return marshalNoEscape(models.DataWrapper[models.EventsFileExport]{Data: models.EventsFileExport{Strings: stringsMap}})
}

// Unmarshal parseia o JSON de volta para o array original de eventos,
// ordenado por ID para entrega determinística.
func (JSONEventsFormatter) Unmarshal(data []byte) ([]event.EventFileData, error) {
	loaded, err := unmarshalWrapped[models.EventsFileExport](data)
	if err != nil {
		return nil, err
	}
	out := make([]event.EventFileData, 0, len(loaded.Strings))
	for eventID, export := range loaded.Strings {
		strings := make([]event.EventStringData, 0, len(export.Strings))
		for _, s := range export.Strings {
			strings = append(strings, event.EventStringData{Index: s.Index, Text: s.Text})
		}
		out = append(out, event.EventFileData{ID: eventID, Strings: strings})
	}
	sort.Slice(out, func(a, b int) bool { return out[a].ID < out[b].ID })
	return out, nil
}

var _ event.IEventsFormatter = JSONEventsFormatter{}
