package objectsfile

import (
	"sort"

	"ffxresources/backend/common"
	"ffxresources/backend/core/converter"
	"ffxresources/backend/datastore"
	"ffxresources/backend/models"
)

// FieldText é um campo textual flattenizado de um objeto: a chave do
// segmento (ex: "name", "ability1", "T"/"sT" de weapons) mais os textos
// não-vazios por idioma. É a unidade que os builders convertem em
// dto.TextRow{Index: objectID, Name: Key}.
type FieldText struct {
	Key   string
	Texts map[string]string
}

// sortedLangs devolve as chaves de idioma ordenadas para extração determinística.
func sortedLangs() []string {
	langs := make([]string, 0, len(common.SupportedLanguages))
	for locKey := range common.SupportedLanguages {
		langs = append(langs, locKey)
	}
	sort.Strings(langs)
	return langs
}

// collectTexts lê os textos não-vazios do segmento por idioma.
func collectTexts(seg datastore.IGlobalLocalizedKeyedStringObject, langs []string) map[string]string {
	if seg == nil {
		return nil
	}
	var out map[string]string
	for _, locKey := range langs {
		if text := seg.GetLocalizedString(locKey); text != "" {
			if out == nil {
				out = make(map[string]string)
			}
			out[locKey] = text
		}
	}
	return out
}

// orderedFieldKeys expõe a ordem dos campos textuais do objeto conforme o
// layout/arquivo. Implementado pelos tipos concretos; quando ausente,
// ExportFieldTexts cai na ordem canônica de staticFields.
type orderedFieldKeys interface {
	OrderedFieldKeys() []string
}

// ExportFieldTexts flatteniza todos os campos textuais do objeto na ordem do
// layout (ordem em que os textos aparecem no arquivo), para que o tradutor
// leia os campos na mesma sequência do jogo. Campos vazios/ausentes são
// pulados.
func ExportFieldTexts(obj datastore.IGlobalLocalizedTextObject) []FieldText {
	if obj == nil {
		return nil
	}
	langs := sortedLangs()
	keys := staticFieldKeys()
	if ordered, ok := obj.(orderedFieldKeys); ok {
		keys = ordered.OrderedFieldKeys()
	}
	var out []FieldText
	for _, key := range keys {
		if texts := collectTexts(obj.GetKeyedString(key), langs); len(texts) > 0 {
			out = append(out, FieldText{Key: key, Texts: texts})
		}
	}
	return out
}

// ApplyFieldTexts aplica os campos de volta no objeto via GetKeyedString
// (chaves ausentes são no-op).
func ApplyFieldTexts(obj datastore.IGlobalLocalizedTextObject, fields []FieldText, version common.GameVersion) {
	if obj == nil {
		return
	}
	for _, f := range fields {
		if f.Key == "" || len(f.Texts) == 0 {
			continue
		}
		applyLocalizedText(obj.GetKeyedString(f.Key), f.Texts, version, f.Key)
	}
}

// createNewKeyedString creates a new KeyedString with the given text and charset.
// Encoding failure (configuração inválida) keeps empty Bytes and logs;
// o chamador (void) não pode abortar.
func createNewKeyedString(text string, charset string, version common.GameVersion) *KeyedString {
	encoded, err := converter.StringToBytes(text, charset, version)
	if err != nil {
		common.LogError("createNewKeyedString: %v", err)
	}
	return &KeyedString{
		Charset: charset,
		Version: version,
		Segment: models.Segment{Offset: 0, Key: 0},
		Bytes:   encoded,
	}
}

// updateOrCreateSegment updates or creates the localized content of a single keyed-string
// segment (name or description), shared between the V1 and V2 name+description objects.
func updateOrCreateSegment(segment datastore.IGlobalLocalizedKeyedStringObject, newText, languageCode string, version common.GameVersion) {
	existingContent := segment.GetLocalizedContent(languageCode)
	charset := common.LanguageCodeToCharset(languageCode)

	if existingContent != nil {
		existingContent.SetString(newText, charset)
	} else {
		segment.SetLocalizedContent(languageCode, createNewKeyedString(newText, charset, version))
	}

	common.LogVerbose("Segment updated (%s): %s", languageCode, newText)
}
