package tags

import "fmt"

type FFXTextTagLetters struct{}

func NewLetters() *FFXTextTagLetters {
	return &FFXTextTagLetters{}
}

func (l *FFXTextTagLetters) FFXTextLettersCodePage() []string {
	codePage := &[]string{}

	l.generateNumberCode(codePage)
	l.generateSpecialCharacterCodePage(codePage)
	l.generateUpperCaseLetterCode(codePage)
	l.generateLowerCaseLetterCode(codePage)

	return *codePage
}

func (l *FFXTextTagLetters) FFXTextSpecialLettersCodePage() []string {
	codePage := &[]string{}

	l.generateSpecialLetterCode(codePage)

	return *codePage
}

func (l *FFXTextTagLetters) generateUpperCaseLetterCode(codePage *[]string) {
	upperCaseLettersMap := l.getUppercaseLettersMap()

	for key, value := range upperCaseLettersMap {
		*codePage = append(*codePage, l.generateLetterCode(key, value))
	}
}

func (l *FFXTextTagLetters) generateLowerCaseLetterCode(codePage *[]string) {
	lowerCaseLettersMap := l.getLowercaseLettersMap()

	for key, value := range lowerCaseLettersMap {
		*codePage = append(*codePage, l.generateLetterCode(key, value))
	}
}

func (l *FFXTextTagLetters) generateSpecialLetterCode(codePage *[]string) {
	specialLettersMap := l.getSpecialLettersMap()

	for key, value := range specialLettersMap {
		*codePage = append(*codePage, l.generateLetterCode(key, value))
	}
}

func (l *FFXTextTagLetters) generateNumberCode(codePage *[]string) {
	numbersMap := l.getNumbersMap()

	for key, value := range numbersMap {
		*codePage = append(*codePage, l.generateLetterCode(key, value))
	}
}

func (l *FFXTextTagLetters) generateSpecialCharacterCodePage(codePage *[]string) {
	specialCharactersMap := l.getSpecialCharactersMap()

	for key, value := range specialCharactersMap {
		*codePage = append(*codePage, l.generateLetterCode(key, value))
	}
}

func (l *FFXTextTagLetters) generateLetterCode(key byte, value string) string {
	return fmt.Sprintf("\\x%02X=%s", key, value)
}

func (l *FFXTextTagLetters) getNumbersMap() map[byte]string {
	lettersMap := map[byte]string{
		0x30: "0",
		0x31: "1",
		0x32: "2",
		0x33: "3",
		0x34: "4",
		0x35: "5",
		0x36: "6",
		0x37: "7",
		0x38: "8",
		0x39: "9",
	}

	return lettersMap
}

func (l *FFXTextTagLetters) getSpecialCharactersMap() map[byte]string {
	lettersMap := map[byte]string{
		0x3A: " ",
		0x3B: "!",
		0x3C: "\"",
		0x3D: "#",
		0x3E: "$",
		0x3F: "%",
		0x40: "&",
		0x41: "'",
		0x42: "(",
		0x43: ")",
		0x44: "*",
		0x45: "+",
		0x46: ",",
		0x47: "-",
		0x48: ".",
		0x49: "/",
		0x4A: ":",
		0x4B: ";",
		0x4C: "<",
		0x4D: "=",
		0x4E: ">",
		0x4F: "?",
		0x6A: "[",
		0x6B: "\\\\",
		0x6C: "]",
		0x6D: "^",
		0x6E: "_",
		0x6F: "`",
		0x8B: "|",
		0x8D: "~",
		0x94: "“",
		0x95: "”",
		0x96: "—",
		0x9D: "¨",
		0x9E: "«",
		0x9F: "°",
		0xA1: "»",
		0xD2: "„",
		0xD3: "…",
		0xD4: "‘",
		0xD5: "’",
		0xDC: "§",
	}

	return lettersMap
}

func (l *FFXTextTagLetters) getUppercaseLettersMap() map[byte]string {
	lettersMap := map[byte]string{
		0x50: "A",
		0x51: "B",
		0x52: "C",
		0x53: "D",
		0x54: "E",
		0x55: "F",
		0x56: "G",
		0x57: "H",
		0x58: "I",
		0x59: "J",
		0x5A: "K",
		0x5B: "L",
		0x5C: "M",
		0x5D: "N",
		0x5E: "O",
		0x5F: "P",
		0x60: "Q",
		0x61: "R",
		0x62: "S",
		0x63: "T",
		0x64: "U",
		0x65: "V",
		0x66: "W",
		0x67: "X",
		0x68: "Y",
		0x69: "Z",
	}

	return lettersMap
}

func (l *FFXTextTagLetters) getLowercaseLettersMap() map[byte]string {
	lettersMap := map[byte]string{
		0x70: "a",
		0x71: "b",
		0x72: "c",
		0x73: "d",
		0x74: "e",
		0x75: "f",
		0x76: "g",
		0x77: "h",
		0x78: "i",
		0x79: "j",
		0x7A: "k",
		0x7B: "l",
		0x7C: "m",
		0x7D: "n",
		0x7E: "o",
		0x7F: "p",
		0x80: "q",
		0x81: "r",
		0x82: "s",
		0x83: "t",
		0x84: "u",
		0x85: "v",
		0x86: "w",
		0x87: "x",
		0x88: "y",
		0x89: "z",
	}

	return lettersMap
}

func (l *FFXTextTagLetters) getSpecialLettersMap() map[byte]string {
	lettersMap := map[byte]string{
		0xA3: "À",
		0xA4: "Á",
		0xA5: "Â",
		0xA6: "Ã",
		0xA7: "Ç",
		0xA8: "È",
		0xA9: "É",
		0xAA: "Ê",
		0xAB: "Ë",
		0xAC: "Ì",
		0xAD: "Í",
		0xAE: "Î",
		0xAF: "Ï",
		0xB0: "Ñ",
		0xB1: "Ò",
		0xB2: "Ó",
		0xB3: "Ô",
		0xB4: "Õ",
		0xB6: "Ù",
		0xB7: "Ú",
		0xB8: "Û",
		0xBA: "à",
		0xBB: "á",
		0xBC: "â",
		0xBD: "ã",
		0xBE: "ç",
		0xBF: "è",
		0xC0: "é",
		0xC1: "ê",
		0xC2: "ë",
		0xC3: "ì",
		0xC4: "í",
		0xC5: "î",
		0xC6: "ï",
		0xC7: "ñ",
		0xC8: "ò",
		0xC9: "ó",
		0xCA: "ô",
		0xCB: "õ",
		0xCC: "ù",
		0xCD: "ú",
		0xCE: "û",
		0xCF: "ü",
	}

	return lettersMap
}
