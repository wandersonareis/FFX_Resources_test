package objectsfile

import (
	"ffxresources/backend/common"
	"ffxresources/backend/datastore"
)

// staticFields lista as chaves de segmento estáticas conhecidas, na ordem
// determinística de extração. Usado por ExportFieldTexts (dto_bridge.go).
var staticFields = []struct {
	key   string
	label string
}{
	{"name", "name"},
	{"simplifiedName", "simplified name"},
	{"description", "description"},
	{"simplifiedDescription", "simplified description"},
	{"effect", "effect"},
	{"effectDescription", "effect description"},
	{"bonus", "bonus"},
	{"BonusIconA", "bonus icon A"},
	{"BonusIconB", "bonus icon B"},
	{"BonusReserve", "bonus reserve"},
	{"sensorText", "sensor text"},
	{"simplifiedSensorText", "simplified sensor text"},
	{"scanText", "scan text"},
	{"simplifiedScanText", "simplified scan text"},
}

func applyLocalizedText(seg datastore.IGlobalLocalizedKeyedStringObject, texts map[string]string, version common.GameVersion, label string) {
	if seg == nil || len(texts) == 0 {
		return
	}
	for languageCode, newText := range texts {
		if newText == "" {
			common.LogVerbose("Empty %s for language %s, ignoring...", label, languageCode)
			continue
		}
		if !common.IsSupportedLanguage(languageCode) {
			common.LogVerbose("Not recognized location for %s: %s", label, languageCode)
			continue
		}
		if newText == seg.GetLocalizedString(languageCode) {
			continue
		}
		updateOrCreateSegment(seg, newText, languageCode, version)
	}
}
