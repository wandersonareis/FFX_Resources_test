package components

var (
	ByteToCharMaps = make(map[string]map[uint]rune)
	CharToByteMaps = make(map[string]map[rune]uint)
)

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
