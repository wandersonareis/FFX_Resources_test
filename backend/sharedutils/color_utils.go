package sharedutils

import (
	"fmt"
	"strconv"
	"strings"
)

var (
	colorToByteMap = map[string]byte{
		"WHITE":     0x41,
		"YELLOW":    0x43,
		"GREY":      0x52,
		"BLUE":      0x88,
		"RED":       0x94,
		"PINK":      0x97,
		"OL_PURPLE": 0xA1,
		"OL_CYAN":   0xB1,
	}
	byteToColorMap = map[byte]string{
		0x41: "WHITE",
		0x43: "YELLOW",
		0x52: "GREY",
		0x88: "BLUE",
		0x94: "RED",
		0x97: "PINK",
		0xA1: "OL_PURPLE",
		0xB1: "OL_CYAN",
	}
)

func ByteToColor(hex byte) string {
	if color, exists := byteToColorMap[hex]; exists {
		return color
	}
	return fmt.Sprintf("%02X", hex)
}

func ColorToByte(color string) byte {
	if val, exists := colorToByteMap[strings.ToUpper(color)]; exists {
		return val
	}
	if parsed, err := strconv.ParseUint(color, 16, 16); err == nil {
		return byte(parsed)
	}
	return 0
}

func GetColorString(hex uint8) string {
	return fmt.Sprintf("{CLR:%s}", ByteToColor(hex))
}
