package datastore

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"sort"
	"sync"
)

// MacroObjects: objetos de macro de uma versão, indexados por ID
// (chunk*0x100 + índice dentro do chunk).
type MacroObjects = components.IMap[int, IGlobalLocalizedMacroStringObject]

// VersionedMacros: objetos de macro indexados por versão do jogo.
type VersionedMacros = components.IMap[common.GameVersion, MacroObjects]

// GlobalDataStore mantém o estado global de todos os dados,
// segregado por versão do jogo (FFX / FFX-2) para macros e events.
type GlobalDataStore struct {
	mu sync.RWMutex
	//KeyItems components.IList[IGlobalLocalizedTextObject]
	macros map[common.GameVersion]MacroObjects
	events map[common.GameVersion]map[string]IEventObject
	lists  map[string]interface{} // Para armazenar IList[T] genéricos
}

func newGlobalDataStore() *GlobalDataStore {
	return &GlobalDataStore{
		//KeyItems: components.NewEmptyList[IGlobalLocalizedTextObject](),
		macros: map[common.GameVersion]MacroObjects{
			common.GameVersionFFX:  components.NewEmptyMap[int, IGlobalLocalizedMacroStringObject](),
			common.GameVersionFFX2: components.NewEmptyMap[int, IGlobalLocalizedMacroStringObject](),
		},
		events: map[common.GameVersion]map[string]IEventObject{
			common.GameVersionFFX:  make(map[string]IEventObject),
			common.GameVersionFFX2: make(map[string]IEventObject),
		},
		lists: make(map[string]any),
	}
}

var Instance = newGlobalDataStore()

var (
	KeyItems      components.IList[IGlobalLocalizedTextObject] = components.NewEmptyList[IGlobalLocalizedTextObject]()
	Commands      components.IList[IGlobalLocalizedTextObject] = components.NewEmptyList[IGlobalLocalizedTextObject]()
	Items         components.IList[IGlobalLocalizedTextObject] = components.NewEmptyList[IGlobalLocalizedTextObject]()
	ArmsTxt       components.IList[IGlobalLocalizedTextObject] = components.NewEmptyList[IGlobalLocalizedTextObject]()
	BattleTxt     components.IList[IGlobalLocalizedTextObject] = components.NewEmptyList[IGlobalLocalizedTextObject]()
	BattleEndTxt  components.IList[IGlobalLocalizedTextObject] = components.NewEmptyList[IGlobalLocalizedTextObject]()
	BuildTxt      components.IList[IGlobalLocalizedTextObject] = components.NewEmptyList[IGlobalLocalizedTextObject]()
	ConfigTxt     components.IList[IGlobalLocalizedTextObject] = components.NewEmptyList[IGlobalLocalizedTextObject]()
	ItemCommands  components.IList[IGlobalLocalizedTextObject] = components.NewEmptyList[IGlobalLocalizedTextObject]()
	MenuTxt       components.IList[IGlobalLocalizedTextObject] = components.NewEmptyList[IGlobalLocalizedTextObject]()
	MMainTxt      components.IList[IGlobalLocalizedTextObject] = components.NewEmptyList[IGlobalLocalizedTextObject]()
	NameTxt       components.IList[IGlobalLocalizedTextObject] = components.NewEmptyList[IGlobalLocalizedTextObject]()
	PlayerRoomTxt components.IList[IGlobalLocalizedTextObject] = components.NewEmptyList[IGlobalLocalizedTextObject]()
	SaveTxt       components.IList[IGlobalLocalizedTextObject] = components.NewEmptyList[IGlobalLocalizedTextObject]()
)

// normalizeGameVersion garante que apenas FFX ou FFX2 sejam usados.
// Qualquer valor desconhecido (ex.: 0) cai para FFX.
func normalizeGameVersion(gameVersion common.GameVersion) common.GameVersion {
	if gameVersion == common.GameVersionFFX2 {
		return common.GameVersionFFX2
	}
	return common.GameVersionFFX
}

// macrosFor retorna o mapa de macros da versão indicada.
// Deve ser chamado com o lock (R ou W) já adquirido.
// Inicializa sob demanda caso a versão ainda não exista.
func (ds *GlobalDataStore) macrosFor(gameVersion common.GameVersion) MacroObjects {
	v := normalizeGameVersion(gameVersion)
	m, ok := ds.macros[v]
	if !ok || m == nil {
		m = components.NewEmptyMap[int, IGlobalLocalizedMacroStringObject]()
		if ds.macros == nil {
			ds.macros = make(map[common.GameVersion]MacroObjects)
		}
		ds.macros[v] = m
	}
	return m
}

// eventsFor retorna o mapa de events da versão indicada.
// Deve ser chamado com o lock (R ou W) já adquirido.
// Inicializa sob demanda caso a versão ainda não exista.
func (ds *GlobalDataStore) eventsFor(gameVersion common.GameVersion) map[string]IEventObject {
	v := normalizeGameVersion(gameVersion)
	e, ok := ds.events[v]
	if !ok || e == nil {
		e = make(map[string]IEventObject)
		if ds.events == nil {
			ds.events = make(map[common.GameVersion]map[string]IEventObject)
		}
		ds.events[v] = e
	}
	return e
}

// ============= KEY ITEMS =============

/* func (ds *GlobalDataStore) GetKeyItem(idx int) IGlobalLocalizedTextObject {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	actual := idx - 0xA000
	if actual >= 0 {
		return ds.KeyItems.Get(actual)
	}
	return nil
} */

/* func (ds *GlobalDataStore) GetKeyItems() components.IList[IGlobalLocalizedTextObject] {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	return ds.KeyItems
} */

/* func (ds *GlobalDataStore) SetKeyItem(idx int, item IGlobalLocalizedTextObject) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if item != nil {
		ds.KeyItems.Add(item)
	}
} */

/* func (ds *GlobalDataStore) SetKeyItems(items components.IList[IGlobalLocalizedTextObject]) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.ClearKeyItems()

	ds.KeyItems = items
} */

/* func (ds *GlobalDataStore) GetKeyItemsLength() int {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	return ds.KeyItems.Len()
} */

/* func (ds *GlobalDataStore) ClearKeyItems() {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.KeyItems.Clear()
} */

// ============= MACROS =============

func (ds *GlobalDataStore) GetMacro(gameVersion common.GameVersion, idx int) (IGlobalLocalizedMacroStringObject, bool) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	return ds.macrosFor(gameVersion).Get(idx)
}

func (ds *GlobalDataStore) GetMacros(gameVersion common.GameVersion) MacroObjects {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	return ds.macrosFor(gameVersion)
}

func (ds *GlobalDataStore) SetMacro(gameVersion common.GameVersion, idx int, macro IGlobalLocalizedMacroStringObject) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if macro != nil {
		ds.macrosFor(gameVersion).Add(idx, macro)
	}
}

// TryAddMacro insere apenas se o ID ainda não existir na versão indicada,
// retornando false caso contrário. A trava de escrita é mantida aqui dentro,
// então o check-and-set é atômico frente a outros acessos ao datastore.
func (ds *GlobalDataStore) TryAddMacro(gameVersion common.GameVersion, idx int, macro IGlobalLocalizedMacroStringObject) bool {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if macro == nil {
		return false
	}
	return ds.macrosFor(gameVersion).TryAdd(idx, macro)
}

func (ds *GlobalDataStore) SetMacros(gameVersion common.GameVersion, macros MacroObjects) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	target := ds.macrosFor(gameVersion)
	target.Clear()
	if macros == nil {
		return
	}
	macros.ForEach(func(idx int, macro IGlobalLocalizedMacroStringObject) {
		if macro != nil {
			target.Add(idx, macro)
		}
	})
}

func (ds *GlobalDataStore) IsEmpty(gameVersion common.GameVersion) bool {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	return ds.macrosFor(gameVersion).Count() > 0
}

func (ds *GlobalDataStore) ClearMacros(gameVersion common.GameVersion) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.macrosFor(gameVersion).Clear()
}

func (ds *GlobalDataStore) ClearAllMacros() {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	for _, m := range ds.macros {
		if m != nil {
			m.Clear()
		}
	}
}

// ============= EVENTS =============

func (ds *GlobalDataStore) GetEvent(gameVersion common.GameVersion, id string) IEventObject {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	return ds.eventsFor(gameVersion)[id]
}

func (ds *GlobalDataStore) SetEvent(gameVersion common.GameVersion, id string, event IEventObject) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if event != nil {
		ds.eventsFor(gameVersion)[id] = event
	}
}

func (ds *GlobalDataStore) ClearEvents(gameVersion common.GameVersion) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.events[normalizeGameVersion(gameVersion)] = make(map[string]IEventObject)
}

func (ds *GlobalDataStore) ClearAllEvents() {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	for v := range ds.events {
		ds.events[v] = make(map[string]IEventObject)
	}
}

// GetAllEventIDs returns all event IDs registered in the datastore for the given game version
func (ds *GlobalDataStore) GetAllEventIDs(gameVersion common.GameVersion) []string {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	var eventIDs []string
	for id := range ds.eventsFor(gameVersion) {
		eventIDs = append(eventIDs, id)
	}
	sort.Strings(eventIDs)
	return eventIDs
}

// ============= LISTS GENÉRICAS =============

/* func (ds *GlobalDataStore) SetList(name string, list interface{}) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.lists[name] = list
} */

/* func (ds *GlobalDataStore) GetList(name string) interface{} {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	return ds.lists[name]
} */

// ============= FUNÇÕES GLOBAIS PARA BACKWARD COMPATIBILITY =============

/* func GetKeyItem(idx int) IGlobalLocalizedTextObject {
	return Instance.GetKeyItem(idx)
} */

/* func GetKeyItems() components.IList[IGlobalLocalizedTextObject] {
	return Instance.KeyItems
} */

/* func SetKeyItem(idx int, item IGlobalLocalizedTextObject) {
	Instance.SetKeyItem(idx, item)
} */

func GetMacro(gameVersion common.GameVersion, idx int) (IGlobalLocalizedMacroStringObject, bool) {
	return Instance.GetMacro(gameVersion, idx)
}

func GetMacros(gameVersion common.GameVersion) MacroObjects {
	Instance.mu.RLock()
	defer Instance.mu.RUnlock()

	return Instance.macrosFor(gameVersion)
}

func SetMacro(gameVersion common.GameVersion, idx int, macro IGlobalLocalizedMacroStringObject) {
	Instance.SetMacro(gameVersion, idx, macro)
}

func TryAddMacro(gameVersion common.GameVersion, idx int, macro IGlobalLocalizedMacroStringObject) bool {
	return Instance.TryAddMacro(gameVersion, idx, macro)
}

func HasMacros(gameVersion common.GameVersion) bool {
	return Instance.IsEmpty(gameVersion)
}

func GetEvent(gameVersion common.GameVersion, id string) IEventObject {
	return Instance.GetEvent(gameVersion, id)
}

func SetEvent(gameVersion common.GameVersion, id string, event IEventObject) {
	Instance.SetEvent(gameVersion, id, event)
}

// ============= BULK OPERATIONS =============

/* func SetKeyItems(items components.IList[IGlobalLocalizedTextObject]) {
	Instance.mu.Lock()
	defer Instance.mu.Unlock()

	Instance.KeyItems.Clear()
	Instance.KeyItems = items
} */

func SetMacros(gameVersion common.GameVersion, macros MacroObjects) {
	Instance.mu.Lock()
	defer Instance.mu.Unlock()

	target := Instance.macrosFor(gameVersion)
	target.Clear()
	if macros == nil {
		return
	}
	macros.ForEach(func(idx int, macro IGlobalLocalizedMacroStringObject) {
		//for idx, macro := range macros {
		if macro != nil {
			target.Add(idx, macro)
		}
	})
}

func SetEvents(gameVersion common.GameVersion, events map[string]IEventObject) {
	Instance.mu.Lock()
	defer Instance.mu.Unlock()

	target := Instance.eventsFor(gameVersion)
	for id, event := range events {
		if event != nil {
			target[id] = event
		}
	}
}
