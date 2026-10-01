package objectsfile

import (
	"ffxresources/backend/common"
	"ffxresources/backend/datastore"
)

// staticFields lista as chaves de segmento estáticas conhecidas, na ordem
// determinística de extração. Usado por ExportFieldTexts (dto_bridge.go).
// Chaves no padrão do fluxo novo (snake_case dos structs C#).
var staticFields = []struct {
	key   string
	label string
}{
	{"name", "name"},
	{"name_simplified", "name simplified"},
	{"desc", "desc"},
	{"desc_simplified", "desc simplified"},
	{"help", "help"},
	{"help_simplified", "help simplified"},
	{"command", "command"},
	{"command_simplified", "command simplified"},
	{"switch_text", "switch text"},
	{"switch_text_simplified", "switch text simplified"},
	{"effect", "effect"},
	{"effect_description", "effect description"},
	{"information", "information"},
	{"creature_data_help", "creature data help"},
	{"messages_0", "message 0"},
	{"messages_1", "message 1"},
	{"messages_2", "message 2"},
	{"messages_3", "message 3"},
	{"sensor_text", "sensor text"},
	{"sensor_text_simplified", "sensor text simplified"},
	{"scan_text", "scan text"},
	{"scan_text_simplified", "scan text simplified"},
}

// staticFieldKeys devolve as chaves de staticFields na ordem canônica.
// É apenas fallback: tipos que conhecem seu layout expõem OrderedFieldKeys.
func staticFieldKeys() []string {
	keys := make([]string, 0, len(staticFields))
	for _, f := range staticFields {
		keys = append(keys, f.key)
	}
	return keys
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
