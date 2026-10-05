package event

import (
	"path/filepath"
	"sync"

	"ffxresources/backend/common"
	"ffxresources/backend/models"
)

// Vigia de frescura dos eventos: carimbo físico (path/size/mtime) do binário
// da localização padrão por evento. Binário modificado no disco (tradução
// copiada/manual para mods/) recarrega o evento na próxima leitura, sem
// reiniciar o app — mesmo vigia do lockit/objectsfile.
//
// Escopo do carimbo: só a localização padrão (a fonte do Traduzido e do
// bulk da descoberta). Os demais idiomas são referência e data/ é imutável
// por design — a tradução que o usuário copia para mods/ entra pela
// localização padrão.

var (
	eventStampsMu sync.Mutex
	eventStamps   = map[common.GameVersion]map[string]common.FileStamp{}
)

// eventStampRelPath devolve o caminho relativo (à raiz de gamefiles) do
// binário da localização padrão do evento.
func eventStampRelPath(version common.GameVersion, eventID string) (string, error) {
	rel, err := EventRelPath(eventID)
	if err != nil {
		return "", err
	}
	return filepath.Join(common.GetLocalizationRootForVersion(version, common.DefaultLocalization), rel), nil
}

// StampEvent registra o carimbo físico do evento (ou o remove quando o
// binário não existe mais).
func StampEvent(version common.GameVersion, eventID string) {
	eventStampsMu.Lock()
	defer eventStampsMu.Unlock()
	m := eventStamps[version]
	if m == nil {
		m = map[string]common.FileStamp{}
		eventStamps[version] = m
	}
	rel, err := eventStampRelPath(version, eventID)
	if err != nil {
		return
	}
	if stamp, ok := common.StampFile(rel); ok {
		m[eventID] = stamp
	} else {
		delete(m, eventID)
	}
}

// stampAllEvents carimba todos os eventos da versão (após a carga bulk).
func stampAllEvents(version common.GameVersion, eventIDs []string) {
	for _, id := range eventIDs {
		StampEvent(version, id)
	}
}

// EnsureAllEventsFresh recarrega os eventos indicados (vazio = todos os
// carimbados da versão) cujo binário mudou no disco. Devolve true quando
// algum store foi atualizado (o chamador invalida caches derivados).
func EnsureAllEventsFresh(version common.GameVersion, eventIDs []string) bool {
	eventStampsMu.Lock()
	var ids []string
	if len(eventIDs) == 0 {
		for id := range eventStamps[version] {
			ids = append(ids, id)
		}
	} else {
		ids = eventIDs
	}
	eventStampsMu.Unlock()

	any := false
	for _, id := range ids {
		if EnsureEventFresh(version, id) {
			any = true
		}
	}
	return any
}

// EnsureEventFresh recarrega o evento do disco quando o binário mudou.
// Devolve true quando o store foi atualizado (o chamador invalida caches
// derivados, como o view dedupado). Evento sem carimbo (nunca carregado):
// sem vigia.
func EnsureEventFresh(version common.GameVersion, eventID string) bool {
	eventStampsMu.Lock()
	stored, watched := eventStamps[version][eventID]
	eventStampsMu.Unlock()
	if !watched {
		return false
	}

	rel, err := eventStampRelPath(version, eventID)
	if err != nil {
		return false
	}
	current, exists := common.StampFile(rel)
	if exists && current == stored {
		return false
	}
	if !exists && stored == (common.FileStamp{}) {
		return false
	}

	common.LogVerbose("[event] %s: binário mudou no disco — recarregando", eventID)
	eventFile, rerr := ReadCompleteEventFile(models.NewEventFileInfo(eventID, version))
	StampEvent(version, eventID) // carimbo novo mesmo em falha: sem loop de recarga
	if rerr != nil {
		common.LogWarning("[event] %s: recarga falhou (store mantém o anterior): %v", eventID, rerr)
		return false
	}
	if eventFile == nil {
		// Sem conteúdo no disco (evento pulado pela régua): o store mantém
		// o anterior — nada a invalidar.
		common.LogVerbose("[event] %s: recarga sem conteúdo — store mantém o anterior", eventID)
		return false
	}
	return true
}
