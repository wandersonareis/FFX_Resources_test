package tags

import "fmt"

type FFXTextItems struct {
	ffxTagsBase
}

func NewTextItems() *FFXTextItems {
	return &FFXTextItems{
		ffxTagsBase: ffxTagsBase{},
	}
}

func (i *FFXTextItems) FFXTextItemsCodePage() []string {
	return i.processCodePage(&ffxItems{itemByte: 0x23})
}

type ffxItems struct {
	itemByte byte
}

func (i *ffxItems) generateCode(key byte, value string) string {
	return fmt.Sprintf("\\x%02X\\x%02X={%s}", i.itemByte, key, value)
}

func (i *ffxItems) getMap() map[byte]string {
	return map[byte]string{
		0x32: "Cloudy Mirror",
		0x33: "Celestial Mirror",
		0x50: "Jecht's Sphere",
		0x51: "Rusty Sword",
		0x53: "Sun Crest",
		0x54: "Sun Sigil",
		0x55: "Moon Crest",
		0x56: "Moon Sigil",
		0x57: "Saturn Crest",
		0x58: "Saturn Sigil",
		0x59: "Mark of Conquest",
		0x5A: "Jupiter Crest",
		0x5B: "Jupiter Sigil",
		0x5C: "Mercury Crest",
		0x5D: "Mercury Sigil",
		0x5E: "Mars Crest",
		0x5F: "Mars Sigil",
		0x60: "Venus Crest",
		0x61: "Venus Sigil",
		0x62: "Blossom Crown",
		0x63: "Flower Scepter",
	}
}
