package objectsfile

import (
	"sort"
	"strconv"

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

// ExportFieldTexts flatteniza todos os campos textuais do objeto em ordem
// determinística: campos estáticos, abilities (ability1..N) e weapons
// (chave + "s"+chave por personagem). Campos vazios/ausentes são pulados,
// como no export legado.
func ExportFieldTexts(obj datastore.IGlobalLocalizedTextObject) []FieldText {
	if obj == nil {
		return nil
	}
	langs := sortedLangs()
	var out []FieldText
	for _, f := range staticFields {
		if texts := collectTexts(obj.GetKeyedString(f.key), langs); len(texts) > 0 {
			out = append(out, FieldText{Key: f.key, Texts: texts})
		}
	}
	for i := 1; ; i++ {
		key := "ability" + strconv.Itoa(i)
		seg := obj.GetKeyedString(key)
		if seg == nil {
			break
		}
		if texts := collectTexts(seg, langs); len(texts) > 0 {
			out = append(out, FieldText{Key: key, Texts: texts})
		}
	}
	if w, ok := obj.(*WeaponsNameTextObject); ok {
		for i, ref := range weaponRefs {
			if texts := collectTexts(w.Names[i], langs); len(texts) > 0 {
				out = append(out, FieldText{Key: ref.key, Texts: texts})
			}
			if texts := collectTexts(w.SimplifiedNames[i], langs); len(texts) > 0 {
				out = append(out, FieldText{Key: "s" + ref.key, Texts: texts})
			}
		}
	}
	return out
}

// ApplyFieldTexts aplica os campos de volta no objeto via GetKeyedString
// (chaves ausentes são no-op, como no import legado). Não toca em disco:
// quem persiste é SaveToBinary do dono da lista.
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
