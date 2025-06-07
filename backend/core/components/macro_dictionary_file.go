package components

type MacroDictionaryFile struct {
	AllStrings   [][]*MacroString
	Localization string
}

func NewMacroDictionaryFile(bytes []byte, localization string) *MacroDictionaryFile {
	mdf := &MacroDictionaryFile{
		AllStrings:   make([][]*MacroString, 0),
		Localization: localization,
	}
	chunks := BytesToChunks(bytes, 16, 0)
	for _, chunk := range chunks {
		mdf.AllStrings = append(mdf.AllStrings, mdf.mapStringsForChunk(chunk))
	}
	MACRODICTFILE[mdf.Localization] = mdf.AllStrings
	return mdf
}

func (mdf *MacroDictionaryFile) mapStringsForChunk(chunk Chunk) []*MacroString {
	if chunk.Offset == 0 {
		return []*MacroString{}
	}
	return FromStringData(chunk.Bytes, LocalizationToCharset(mdf.Localization))
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
		macroObject, exists := MacroLookup[key]
		if !exists {
			macroObject = NewLocalizedMacroStringObject()
			MacroLookup[key] = macroObject
		}
		macroObject.SetLocalizedContent(mdf.Localization, macroStr)
	}
}
