package macrodic

import (
	"ffxresources/backend/core/components"
	"ffxresources/backend/datastore"
	"ffxresources/backend/models"
)

type MacroDictionaryFile struct {
	AllStrings   [][]*MacroString
	Localization string
}

var MACRODICTFILE = make(map[string][][]*MacroString)
var MacroLookup = make(map[int]*LocalizedMacroStringObject)

func NewMacroDictionaryFile(bytes []byte, localization string) *MacroDictionaryFile {
	mdf := &MacroDictionaryFile{
		AllStrings:   make([][]*MacroString, 0),
		Localization: localization,
	}
	chunks := components.BytesToChunks(bytes, 16, 0)
	for _, chunk := range chunks {
		mdf.AllStrings = append(mdf.AllStrings, mdf.mapStringsForChunk(chunk))
	}
	MACRODICTFILE[mdf.Localization] = mdf.AllStrings
	return mdf
}

func (mdf *MacroDictionaryFile) mapStringsForChunk(chunk models.Chunk) []*MacroString {
	if chunk.Offset == 0 {
		return []*MacroString{}
	}
	return FromStringData(chunk.Bytes, components.GetCharsetForLanguage(mdf.Localization))
}

func (mdf *MacroDictionaryFile) PublishStrings() {
	for i := range mdf.AllStrings {
		mdf.publishStringsOfChunk(i)
	}
}

func (mdf *MacroDictionaryFile) publishStringsOfChunk(i int) {
	list := mdf.AllStrings[i]
	for j, macroStr := range list {
		key := i*0x100 + j
		/* macroObject, exists := MacroLookup[key]
		if !exists {
			macroObject = NewLocalizedMacroStringObject()
			MacroLookup[key] = macroObject
		}
		macroObject.SetLocalizedContent(mdf.Localization, macroStr) */

		macroO, exist := datastore.Instance.GetMacro(key)
		if !exist {
			macroO = NewLocalizedMacroStringObject()
			datastore.Instance.SetMacro(key, macroO)
		}
		macroO.SetLocalizedContent(mdf.Localization, macroStr)
	}
}
