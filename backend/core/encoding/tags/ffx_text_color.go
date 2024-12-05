package tags

import (
	"fmt"
	"slices"
)

type FFXTextTagColor struct {
	textColorByte byte
	textColorTag  byte
}

func NewTextTagColor() *FFXTextTagColor {
	return &FFXTextTagColor{
		textColorByte: 0x0A,
		textColorTag:  0x01,
	}
}

func (c *FFXTextTagColor) FFXColorsPage() []string {
	return slices.Concat(
		c.generateColorCommand(),
		c.generateColorsCodePage(),
	)
}

func (c *FFXTextTagColor) generateColorCommand() []string {
	return []string{
		fmt.Sprintf("\\x%02X\\c%02X={Color:\\h%02X}", c.textColorByte, c.textColorTag, c.textColorTag),
	}
}

func (c *FFXTextTagColor) generateColorsCodePage() []string {
	colorsMap := c.getColorsMap()
	colors := make([]string, 0, len(colorsMap))

	generateColorsCode := func(colorByte byte, colorName string) string {
		return fmt.Sprintf("\\x%02X\\x%02X={%s}", c.textColorByte, colorByte, colorName)
	}

	for key, value := range colorsMap {
		colors = append(colors, generateColorsCode(key, value))
	}

	return colors
}

func (c *FFXTextTagColor) getColorsMap() map[byte]string {
	colorMap := map[byte]string{
		0x41: "White",
		0x43: "Yellow",
		0x52: "Gray",
		0x88: "Blue",
		0x94: "Red",
		0x97: "Pink",
		0xA1: "OPurple",
		0xB1: "OCyan",
	}

	return colorMap
}
