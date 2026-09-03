package sharedutils

var (
	ByteToCharMaps            = make(map[string]map[uint]rune)
	CharToByteMaps            = make(map[string]map[rune]uint)
	WriteLinebreaksAsCommands = true
)

func CharToBytes(chr rune, charset string) []uint {
	if chr == '\n' {
		return []uint{0x03}
	}

	indexValue, exists := CharToByteMaps[charset][chr]
	if !exists {
		return nil
	}

	if indexValue < 0x100 {
		return []uint{indexValue}
	}

	section := (indexValue - 0x30) / 0xD0
	byte1 := section + 0x2B
	byte2 := indexValue - (section * 0xD0)

	if byte1 <= 0x2F {
		return []uint{uint(byte1), uint(byte2)}
	}

	adjustedValue := indexValue - 0x410

	if adjustedValue < 0x100 {
		return []uint{0x04, adjustedValue}
	} else {
		adjustedSection := (adjustedValue - 0x30) / 0xD0
		adjustedByte1 := adjustedSection + 0x2B
		adjustedByte2 := adjustedValue - (adjustedSection * 0xD0)
		return []uint{0x04, uint(adjustedByte1), uint(adjustedByte2)}
	}
}

func ByteToChar(hex uint, charset string) (rune, bool) {
	charsetMap, exists := ByteToCharMaps[charset]
	if !exists {
		return 0, false
	}

	char, exists := charsetMap[hex]
	return char, exists
}

func SetCharMap(charset string, byteToCharMap map[uint]rune, charToByteMap map[rune]uint) {
	ByteToCharMaps[charset] = byteToCharMap
	CharToByteMaps[charset] = charToByteMap
}
