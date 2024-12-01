package tags

import "fmt"

type FFXTextTagCharacter struct {
	characterByte byte
}

func NewTextTagCharacter() *FFXTextTagCharacter {
	return &FFXTextTagCharacter{characterByte: 0x13}
}

func (c *FFXTextTagCharacter) FFXTextCharacterCodePage() []string {
	codePage := make([]string, 0, len(c.getCharactersMap()))
    for key, value := range c.getCharactersMap() {
        codePage = append(codePage, fmt.Sprintf("\\x%02X\\x%02X={%s}", c.characterByte, key, value))
    }
    return codePage
}

/* func (c *FFXTextTagCharacter) generateCharactersCodePage(codePage *[]string) {
	charactersMap := c.getCharactersMap()

	for key, value := range charactersMap {
		*codePage = append(*codePage, fmt.Sprintf("\\x%02X\\x%02X={%s}", c.characterByte, key, value))
	}
} */

func (c *FFXTextTagCharacter) getCharactersMap() map[byte]string {
	charactersMap := map[byte]string{
		0x30: "Tidus",
		0x31: "Yuna",
		0x32: "Auron",
		0x33: "Kimahri",
		0x34: "Wakka",
		0x35: "Lulu",
		0x36: "Rikku",
		0x37: "Seymour",
		0x38: "Valefor",
		0x39: "Ifrit",
		0x3A: "Ixion",
		0x3B: "Shiva",
		0x3C: "Bahamut",
		0x3D: "Anima",
		0x3E: "Yojimbo",
		0x3F: "Cindy",
		0x40: "Sandy",
		0x41: "Mindy",
	}

	return charactersMap
}
