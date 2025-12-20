package event

import (
	"ffxresources/backend/datastore"
)

// Adapter para converter EventFile para datastore.EventObject
type eventDatastoreAdapter struct {
	obj *EventFile
}

func (ea *eventDatastoreAdapter) GetName() string {
	if ea.obj != nil {
		return ea.obj.GetName()
	}
	return ""
}

func (ea *eventDatastoreAdapter) GetID() string {
	if ea.obj != nil {
		return ea.obj.ID
	}
	return ""
}

// SetEvent registra um evento diretamente no datastore (fonte única da verdade)
func SetEvent(eventID string, eventFile *EventFile) {
	if eventFile != nil {
		adapter := &eventDatastoreAdapter{obj: eventFile}
		datastore.SetEvent(eventID, adapter)
	}
}

// GetEvent recupera um evento do datastore e converte de volta para *EventFile
func GetEvent(eventID string) *EventFile {
	eventObj := datastore.GetEvent(eventID)
	if eventObj != nil {
		// Type assertion segura para recuperar o EventFile original
		if adapter, ok := eventObj.(*eventDatastoreAdapter); ok {
			return adapter.obj
		}
	}
	return nil
}

// SetEvents registra múltiplos eventos no datastore de uma vez
func SetEvents(events map[string]*EventFile) {
	for eventID, eventFile := range events {
		SetEvent(eventID, eventFile)
	}
}

// ClearEvents limpa todos os eventos do datastore
func ClearEvents() {
	datastore.Instance.ClearEvents()
}

// GetAllEventIDs recupera todos os IDs dos eventos registrados no datastore
func GetAllEventIDs() []string {
	return datastore.Instance.GetAllEventIDs()
}

// HasEvents verifica se há eventos carregados no datastore (útil para testes)
func HasEvents() bool {
	eventIDs := GetAllEventIDs()
	return len(eventIDs) > 0
}
