package tags

import "fmt"

type FFXTextTagColor struct {
	colorByte byte
	colorTag  byte
}

func NewTextTagColor() *FFXTextTagColor {
	colorByte := byte(0x0A)
	colorTag := byte(0x01)

	return &FFXTextTagColor{
		colorByte: colorByte,
		colorTag:  colorTag,
	}
}

func (c *FFXTextTagColor) FFXColorsPage() []string {
	colorsMap := c.getColorsMap()
	colors := make([]string, 0, len(colorsMap)+1)

	colors = append(colors, c.colorCommand())

	for key, value := range colorsMap {
		colors = append(colors, fmt.Sprintf("\\x%02X\\x%02X={%s}", c.colorByte, key, value))
	}

	return colors
}

func (c *FFXTextTagColor) colorCommand() string {
	return fmt.Sprintf("\\x%02X\\c%02X={Color:\\h%02X}", c.colorByte, c.colorTag, c.colorTag)

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
