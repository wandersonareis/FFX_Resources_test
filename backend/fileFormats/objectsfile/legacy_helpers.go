package objectsfile

import (
	"ffxresources/backend/datastore"
	"ffxresources/backend/models"
)

// Shared helpers kept from the legacy concrete types so the generic
// KeyedStringFile path can still use them, and any external package
// that relied on these identifiers continues to compile.

// ----------------------------- Segment helper -----------------------------

func getSegment(keyedObj datastore.IGlobalKeyedString) models.Segment {
	if keyedObj == nil {
		return models.Segment{Offset: 0, Key: 0}
	}
	if ks, ok := keyedObj.(*KeyedString); ok {
		return ks.Segment
	}
	return models.Segment{Offset: 0, Key: 0}
}

// ----------------------- Package-level registries -------------------------

var (
	MONMAGIC1    datastore.IBinaryFile
	MONMAGIC2    datastore.IBinaryFile
	ARMOUR_SKILL datastore.IBinaryFile
	BATTLE_TEXT  datastore.IBinaryFile
	BUILD_TEXT   datastore.IBinaryFile
	CONFIG_TEXT  datastore.IBinaryFile
	ITEM_TEXT    datastore.IBinaryFile
	MONSTER_TEXT datastore.IBinaryFile
	NAME_TEXT    datastore.IBinaryFile
	PLAYER_ROOM  datastore.IBinaryFile
	PLAYER_SAVE  datastore.IBinaryFile
	SAVE_TEXT    datastore.IBinaryFile
	STATS_TEXT   datastore.IBinaryFile
	SUMMON_TEXT  datastore.IBinaryFile
)

// Convenience accessors for code that previously referenced concrete types.

func GetCommand(idx int) datastore.IGlobalLocalizedTextObject {
	if datastore.Commands == nil || datastore.Commands.IsEmpty() || idx < 0 || idx >= datastore.Commands.Len() {
		return nil
	}
	return datastore.Commands.Get(idx)
}

func GetKeyItem(idx int) datastore.IGlobalLocalizedTextObject {
	return datastore.KeyItems.Get(idx)
}

type Nameable interface {
	GetName(string) string
}

func GetNameableObject(typ string, idx int) Nameable {
	switch typ {
	case "command":
		if cmd := GetCommand(idx); cmd != nil {
			return cmd
		}
	case "keyItem":
		if k := GetKeyItem(idx); k != nil {
			return k
		}
	}
	return nil
}