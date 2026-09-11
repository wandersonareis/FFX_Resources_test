package objectsfile

import (
	"ffxresources/backend/datastore"
	"ffxresources/backend/core/components"
)

// Stub: ReadMonsterLocalizations - Not implemented.
func ReadMonsterLocalizations(patternPath string) datastore.IBinaryFile {
	return nil
}

// Stub: ReadWeaponNamesLocalizations - Not implemented.
func ReadWeaponNamesLocalizations(patternPath string) datastore.IBinaryFile {
	return nil
}

// Stub: ReadJobLocalizations - Not implemented.
func ReadJobLocalizations(patternPath string, segmentPosition int) datastore.IBinaryFile {
	return nil
}

// Stub: ReadNameDescriptionEffectAbilitiesLocalizations - Not implemented.
func ReadNameDescriptionEffectAbilitiesLocalizations(patternPath string, plateAbilitiesCount int, plateEffectSegmentDefaultPosition int) datastore.IBinaryFile {
	return nil
}

// ReadMonsterNamesLocalizations placeholder.
func ReadMonsterNamesLocalizations(patternPath string) datastore.IBinaryFile {
	return nil
}

// ReadNameOnlyMonsterLocalizations placeholder.
func ReadNameOnlyMonsterLocalizations(patternPath string, position int) components.IList[datastore.IGlobalLocalizedTextObject] {
	return nil
}
