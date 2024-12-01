package tags

import "fmt"

type FFXTextItems struct {
	itemByte byte
}

func NewTextItems() *FFXTextItems {
	itemByte := byte(0x23)

	return &FFXTextItems{
		itemByte: itemByte,
	}
}

func (i *FFXTextItems) FFXTextItemsCodePage() []string {
	return i.generateItems()
}

func (i *FFXTextItems) generateItems() []string {
	itemsMap := i.getItemsMap()
	codePage := make([]string, 0, len(itemsMap))

	generateItemCode := func(key byte, value string) string {
		return fmt.Sprintf("\\x%02X\\x%02X={%s}", i.itemByte, key, value)
	}

	for key, value := range itemsMap {
		codePage = append(codePage, generateItemCode(key, value))
	}

	return codePage
}

func (i *FFXTextItems) getItemsMap() map[byte]string {
	itemsMap := map[byte]string{
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

	return itemsMap
}
