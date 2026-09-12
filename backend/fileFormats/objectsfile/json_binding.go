package objectsfile

import (
	"ffxresources/backend/common"
	"ffxresources/backend/datastore"
)

var staticFields = []struct {
	key  string
	get  func(*JSONEntry) *map[string]string
	label string
}{
	{"name", func(e *JSONEntry) *map[string]string { return &e.Name }, "name"},
	{"simplifiedName", func(e *JSONEntry) *map[string]string { return &e.SimplifiedName }, "simplified name"},
	{"description", func(e *JSONEntry) *map[string]string { return &e.Description }, "description"},
	{"simplifiedDescription", func(e *JSONEntry) *map[string]string { return &e.SimplifiedDescription }, "simplified description"},
	{"effect", func(e *JSONEntry) *map[string]string { return &e.Effect }, "effect"},
	{"effectDescription", func(e *JSONEntry) *map[string]string { return &e.EffectDescription }, "effect description"},
	{"sensorText", func(e *JSONEntry) *map[string]string { return &e.SensorText }, "sensor text"},
	{"simplifiedSensorText", func(e *JSONEntry) *map[string]string { return &e.SimplifiedSensorText }, "simplified sensor text"},
	{"scanText", func(e *JSONEntry) *map[string]string { return &e.ScanText }, "scan text"},
	{"simplifiedScanText", func(e *JSONEntry) *map[string]string { return &e.SimplifiedScanText }, "simplified scan text"},
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

func staticBindings(obj datastore.IGlobalLocalizedTextObject, data *JSONEntry) []binding {
	bs := make([]binding, 0, len(staticFields))
	for _, f := range staticFields {
		bs = append(bs, bind(obj, f.key, f.get(data)))
	}
	return bs
}