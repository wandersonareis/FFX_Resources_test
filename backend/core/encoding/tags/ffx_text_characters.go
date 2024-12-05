package tags

import "fmt"

type FFXTextTagCharacter struct {
	ffxTagsBase
}

func NewTextTagCharacter() *FFXTextTagCharacter {
	return &FFXTextTagCharacter{ffxTagsBase: ffxTagsBase{}}
}

func (c *FFXTextTagCharacter) FFXTextCharacterCodePage() []string {
	return c.processCodePage(&ffxCharacter{characterByte: 0x13})
}

type ffxCharacter struct {
	characterByte byte
}

func (c *ffxCharacter) getMap() map[byte]string {
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

func (c *ffxCharacter) generateCode(key byte, value string) string {
	return fmt.Sprintf("\\x%02X\\x%02X={%s}", c.characterByte, key, value)
}
