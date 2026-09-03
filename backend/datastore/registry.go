package datastore

import (
	"ffxresources/backend/core/components"
	"sort"
	"sync"
)

// GlobalDataStore mantém o estado global de todos os dados
type GlobalDataStore struct {
	mu sync.RWMutex
	//KeyItems components.IList[IGlobalLocalizedTextObject]
	macros components.IMap[int, IGlobalLocalizedMacroStringObject]
	events map[string]IEventObject
	lists  map[string]interface{} // Para armazenar IList[T] genéricos
}

var Instance = &GlobalDataStore{
	//KeyItems: components.NewEmptyList[IGlobalLocalizedTextObject](),
	macros: components.NewEmptyMap[int, IGlobalLocalizedMacroStringObject](),
	events: make(map[string]IEventObject),
	lists:  make(map[string]any),
}

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

func (ds *GlobalDataStore) GetMacro(idx int) (IGlobalLocalizedMacroStringObject, bool) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	return ds.macros.Get(idx)
}

func (ds *GlobalDataStore) GetMacros() components.IMap[int, IGlobalLocalizedMacroStringObject] {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	return ds.macros
}

func (ds *GlobalDataStore) SetMacro(idx int, macro IGlobalLocalizedMacroStringObject) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if macro != nil {
		ds.macros.Add(idx, macro)
	}
}

func (ds *GlobalDataStore) SetMacros(macros components.IMap[int, IGlobalLocalizedMacroStringObject]) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.macros.Clear()
	macros.ForEach(func(idx int, macro IGlobalLocalizedMacroStringObject) {
		if macro != nil {
			ds.macros.Add(idx, macro)
		}
	})
}

func (ds *GlobalDataStore) IsEmpty() bool {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	return ds.macros.Count() > 0
}

func (ds *GlobalDataStore) ClearMacros() {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.macros.Clear()
}

// ============= EVENTS =============

func (ds *GlobalDataStore) GetEvent(id string) IEventObject {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	return ds.events[id]
}

func (ds *GlobalDataStore) SetEvent(id string, event IEventObject) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if event != nil {
		ds.events[id] = event
	}
}

func (ds *GlobalDataStore) ClearEvents() {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.events = make(map[string]IEventObject)
}

// GetAllEventIDs returns all event IDs registered in the datastore
func (ds *GlobalDataStore) GetAllEventIDs() []string {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	var eventIDs []string
	for id := range ds.events {
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

func GetMacro(idx int) (IGlobalLocalizedMacroStringObject, bool) {
	return Instance.GetMacro(idx)
}

func GetMacros() components.IMap[int, IGlobalLocalizedMacroStringObject] {
	Instance.mu.RLock()
	defer Instance.mu.RUnlock()

	return Instance.macros
}

func SetMacro(idx int, macro IGlobalLocalizedMacroStringObject) {
	Instance.SetMacro(idx, macro)
}

func HasMacros() bool {
	return Instance.IsEmpty()
}

func GetEvent(id string) IEventObject {
	return Instance.GetEvent(id)
}

func SetEvent(id string, event IEventObject) {
	Instance.SetEvent(id, event)
}

// ============= BULK OPERATIONS =============

/* func SetKeyItems(items components.IList[IGlobalLocalizedTextObject]) {
	Instance.mu.Lock()
	defer Instance.mu.Unlock()

	Instance.KeyItems.Clear()
	Instance.KeyItems = items
} */

func SetMacros(macros components.IMap[int, IGlobalLocalizedMacroStringObject]) {
	Instance.mu.Lock()
	defer Instance.mu.Unlock()

	Instance.macros.Clear()
	macros.ForEach(func(idx int, macro IGlobalLocalizedMacroStringObject) {
		//for idx, macro := range macros {
		if macro != nil {
			Instance.macros.Add(idx, macro)
		}
	})
}

func SetEvents(events map[string]IEventObject) {
	Instance.mu.Lock()
	defer Instance.mu.Unlock()

	for id, event := range events {
		if event != nil {
			Instance.events[id] = event
		}
	}
}
