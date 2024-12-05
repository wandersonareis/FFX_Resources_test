package tags

import (
	"fmt"
	"slices"
)

type FFXTextTagText struct {
	textFontByte byte
}

func NewText() *FFXTextTagText {
	textFontByte := byte(0x0E)

	return &FFXTextTagText{
		textFontByte: textFontByte,
	}
}

func (t *FFXTextTagText) FFXTextTextPage() []string {
	return slices.Concat(
		t.generateLineBreaks(),
		t.generateFonts(),
	)
}

func (t *FFXTextTagText) generateLineBreaks() []string {
	texts := make([]string, 0, len(t.getTextLBMap()))

	for key, value := range t.getTextLBMap() {
		texts = append(texts, fmt.Sprintf("\\x%02X=%s", key, value))
	}

	return texts
}

func (t *FFXTextTagText) generateFonts() []string {
	texts := make([]string, 0, len(t.getTextFontMap()))

	for key, value := range t.getTextFontMap() {
		texts = append(texts, fmt.Sprintf("\\x%02X\\x%02X={%s}", t.textFontByte, key, value))
	}

	return texts
}

func (t *FFXTextTagText) getTextFontMap() map[byte]string {
	return map[byte]string{
		0x40: "Font Italic",
		0x41: "Font Normal",
	}
}

func (t *FFXTextTagText) getTextLBMap() map[byte]string {
	return map[byte]string{
		0x00: "{END}",
		0x01: fmt.Sprintf("%s{NEWPAGE}%s", LineBreakString(), LineBreakString()),
		0x03: LineBreakString(),
	}
}
