package tags

import (
	"fmt"
	"slices"
)

type FFXTextTagColor struct {
	ffxTagsBase

	textColorID   byte
	textColorByte byte
}

func NewTextTagColor() *FFXTextTagColor {
	return &FFXTextTagColor{
		ffxTagsBase:   ffxTagsBase{},
		textColorID:   0x01,
		textColorByte: 0x0A,
	}
}

func (c *FFXTextTagColor) FFXColorsCodePage() []string {
	return slices.Concat(
		c.generateColorCommand(),
		c.ffxTagsBase.processCodePage(&ffxTextColor{colorByte: c.textColorByte}),
	)
}

func (c *FFXTextTagColor) generateColorCommand() []string {
	return []string{
		fmt.Sprintf("\\x%02X\\c%02X={Color:\\h%02X}", c.textColorByte, c.textColorID, c.textColorID),
	}
}

type ffxTextColor struct {
	colorByte byte
}

func (c *ffxTextColor) generateCode(colorByte byte, colorName string) string {
	return fmt.Sprintf("\\x%02X\\x%02X={%s}", c.colorByte, colorByte, colorName)
}

func (c *ffxTextColor) getMap() map[byte]string {
	return map[byte]string{
		0x41: "White",
		0x43: "Yellow",
		0x52: "Gray",
		0x88: "Blue",
		0x94: "Red",
		0x97: "Pink",
		0xA1: "OPurple",
		0xB1: "OCyan",
	}
}
