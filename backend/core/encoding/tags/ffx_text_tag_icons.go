package tags

import "fmt"

type FFXTextTagIcons struct{}

func NewTextTagIcons() *FFXTextTagIcons {
	return &FFXTextTagIcons{}
}

func (i *FFXTextTagIcons) FFXTextIconsCodePage() []string {
	return i.generateIconsCodePage()
}

func (i *FFXTextTagIcons) generateIconsCodePage() []string {
	iconsMap := i.getIconsMap()

	codePage := make([]string, 0, len(iconsMap))

	generateIconCode := func(key byte, value string) string {
		return fmt.Sprintf("\\x%02X={%s}", key, value)
	}

	for key, value := range iconsMap {
		codePage = append(codePage, generateIconCode(key, value))
	}

	return codePage
}

// I get this from FFX game font file
func (i *FFXTextTagIcons) getIconsMap() map[byte]string {
	iconsMap := map[byte]string{
		0x8E: "POINT",
		0x8F: "[",
		0x90: "]",
		0x91: "NOTE",
		0x92: "HEART",
		0x98: "!",
		0x99: "UP",
		0x9A: "DOWN",
		0x9B: "LEFT",
		0x9C: "RIGHT",
		0xB9: "BETA",
		0xD0: ",",
		0xD1: "FUNC",
		0xD6: "BIG_POINT",
		0xD8: "~",
		0xD9: "TM",
		0xDB: ">",
		0xDD: "Copyright",
		0xDF: "Registered",
		0xE0: "+-",
		0xE1: "2",
		0xE2: "3",
		0xE3: "1/4",
		0xE4: "1/2",
		0xE5: "3/4",
		0xE6: "*",
		0xE7: "/",
		0xE8: "<",
		0xE9: "...",
	}

	return iconsMap
}
