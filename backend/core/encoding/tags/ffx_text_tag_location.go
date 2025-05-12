package tags

import "fmt"

type FFXTextTagLocation struct {
	ffxTagsBase
}

func NewTextTagLocation() *FFXTextTagLocation {
	return &FFXTextTagLocation{
		ffxTagsBase: ffxTagsBase{},
	}
}

func (l *FFXTextTagLocation) FFXTextUTF8CodePage() []string {
	return l.generateUTF8CodePage()
}

func (l *FFXTextTagLocation) generateUTF8CodePage() []string {
	return l.processCodePage(&ffxLocationEncoding{})
}

type ffxLocationEncoding struct{}

func (l *ffxLocationEncoding) generateCode(key byte, value string) string {
	return fmt.Sprintf("%s=%s", value, value)
}

func (l *ffxLocationEncoding) getMap() map[byte]string {
	return map[byte]string{
		0xC0: "À",
		0xC1: "Á",
		0xC2: "Â",
		0xC4: "Ä",
		0xC7: "Ç",
		0xC8: "È",
		0xC9: "É",
		0xCA: "Ê",
		0xCB: "Ë",
		0xCC: "Ì",
		0xCD: "Í",
		0xCE: "Î",
		0xCF: "Ï",
		0xD1: "Ñ",
		0xD2: "Ò",
		0xE0: "à",
		0xE1: "á",
		0xE2: "â",
		0xE4: "ä",
		0xE7: "ç",
		0xE8: "è",
		0xE9: "é",
		0xEA: "ê",
		0xEB: "ë",
		0xEC: "ì",
		0xED: "í",
		0xEE: "î",
		0xEF: "ï",
		0xF1: "ñ",
		0xF2: "ò",
	}
}
