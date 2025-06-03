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
	return mdf
}

// mapStringsForChunk converts a chunk to a slice of MacroStrings for the file's localization.
func (mdf *MacroDictionaryFile) mapStringsForChunk(chunk Chunk) []*MacroString {
	if chunk.Offset == 0 {
		return []*MacroString{}
	}
	// FromStringData returns []*MacroString based on bytes and charset
	return FromStringDataDev(chunk.Bytes, LocalizationToCharset(mdf.Localization))
}

// PublishStrings iterates all parsed strings and registers them in MACRO_LOOKUP.
func (mdf *MacroDictionaryFile) PublishStrings() {
	for i := range mdf.AllStrings {
		mdf.publishStringsOfChunk(i)
	}
}

// publishStringsOfChunk registers each MacroString of a given chunk index into MACRO_LOOKUP.
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
