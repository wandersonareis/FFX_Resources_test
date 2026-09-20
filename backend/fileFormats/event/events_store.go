package event

import (
	"ffxresources/backend/common"
	"ffxresources/backend/datastore"
)

var _ datastore.IEventObject = (*EventFile)(nil)

// SetEvent registra um evento diretamente no datastore (fonte única da verdade)
func SetEvent(gameVersion common.GameVersion, eventID string, eventFile *EventFile) {
	if eventFile != nil {
		datastore.SetEvent(resolveEventVersion(eventFile, gameVersion), eventID, eventFile)
	}
}

// GetEvent recupera um evento do datastore e converte de volta para *EventFile
func GetEvent(gameVersion common.GameVersion, eventID string) *EventFile {
	eventObj := datastore.GetEvent(gameVersion, eventID)
	if eventObj != nil {
		if eventFile, ok := eventObj.(*EventFile); ok {
			return eventFile
		}
	}
	return nil
}

// SetEvents registra múltiplos eventos no datastore de uma vez
func SetEvents(gameVersion common.GameVersion, events map[string]*EventFile) {
	for eventID, eventFile := range events {
		SetEvent(gameVersion, eventID, eventFile)
	}
}

// ClearEvents limpa todos os eventos do datastore para a versão indicada
func ClearEvents(gameVersion common.GameVersion) {
	datastore.Instance.ClearEvents(gameVersion)
}

// ClearAllEvents limpa os eventos de todas as versões
func ClearAllEvents() {
	datastore.Instance.ClearAllEvents()
}

// GetAllEventIDs recupera todos os IDs dos eventos registrados no datastore para a versão indicada
func GetAllEventIDs(gameVersion common.GameVersion) []string {
	return datastore.Instance.GetAllEventIDs(gameVersion)
}

// HasEvents verifica se há eventos carregados no datastore (útil para testes)
func HasEvents(gameVersion common.GameVersion) bool {
	eventIDs := GetAllEventIDs(gameVersion)
	return len(eventIDs) > 0
}

// resolveEventVersion usa a versão do EventFile quando disponível, senão o fallback.
func resolveEventVersion(eventFile *EventFile, fallback common.GameVersion) common.GameVersion {
	if eventFile != nil && eventFile.Version.String() != "" {
		return eventFile.Version
	}
	return fallback
}
