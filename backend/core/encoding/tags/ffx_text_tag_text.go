package tags

import "fmt"

type FFXTextTagText struct {
	textFontByte byte
	textLBMap    map[byte]string
	textFontMap  map[byte]string
}

func NewText() *FFXTextTagText {
	textFontByte := byte(0x0E)

	textLBMap := map[byte]string{
		0x00: "{END}",
		0x01: fmt.Sprintf("%s{NEWPAGE}%s", LineBreakString(), LineBreakString()),
		0x03: LineBreakString(),
	}

	textFontMap := map[byte]string{
		0x40: "Font Italic",
		0x41: "Font Normal",
	}

	return &FFXTextTagText{
		textFontByte: textFontByte,
		textLBMap:    textLBMap,
		textFontMap:  textFontMap,
	}
}

func (t *FFXTextTagText) FFXTextTextPage() []string {
	texts := make([]string, 0, len(t.textLBMap)+len(t.textFontMap))

	t.generateLineBreaks(&texts)
	t.generateFonts(&texts)

	return texts
}

func (t *FFXTextTagText) generateLineBreaks(codePage *[]string) {
	for key, value := range t.textLBMap {
		*codePage = append(*codePage, fmt.Sprintf("\\x%02X=%s", key, value))
	}
}

func (t *FFXTextTagText) generateFonts(codePage *[]string) {
	for key, value := range t.textFontMap {
		*codePage = append(*codePage, fmt.Sprintf("\\x%02X\\x%02X={%s}", t.textFontByte, key, value))
	}
}
