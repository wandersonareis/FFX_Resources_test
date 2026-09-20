package ffxencoding

import (
	"ffxresources/backend/common"
)

// Mapas de charset versionados por jogo.
//
// Antes havia um único bucket global por charset, então carregar FFX-2
// sobrescrevia os mapas de FFX. Agora cada versão (FFX/FFX-2/LastMiss) tem seu
// próprio bucket: ByteToCharMaps[versão][charset].
var (
	ByteToCharMaps = make(map[common.GameVersion]map[string]map[uint]rune)
	CharToByteMaps = make(map[common.GameVersion]map[string]map[rune]uint)
)

func ensureVersionBucket(version common.GameVersion) common.GameVersion {
	if ByteToCharMaps == nil {
		ByteToCharMaps = make(map[common.GameVersion]map[string]map[uint]rune)
	}
	if CharToByteMaps == nil {
		CharToByteMaps = make(map[common.GameVersion]map[string]map[rune]uint)
	}
	if ByteToCharMaps[version] == nil {
		ByteToCharMaps[version] = make(map[string]map[uint]rune)
	}
	if CharToByteMaps[version] == nil {
		CharToByteMaps[version] = make(map[string]map[rune]uint)
	}
	return version
}

// ByteToChar resolve um byte para rune no charset da versão indicada.
func ByteToChar(hex uint, charset string, version common.GameVersion) (rune, bool) {
	byCharset, exists := ByteToCharMaps[version]
	if !exists {
		return 0, false
	}
	charsetMap, exists := byCharset[charset]
	if !exists {
		return 0, false
	}

	char, exists := charsetMap[hex]
	return char, exists
}

// CharToByte resolve uma rune para byte no charset da versão indicada.
func CharToByte(chr rune, charset string, version common.GameVersion) (uint, bool) {
	byCharset, exists := CharToByteMaps[version]
	if !exists {
		return 0, false
	}
	charsetMap, exists := byCharset[charset]
	if !exists {
		return 0, false
	}

	b, exists := charsetMap[chr]
	return b, exists
}

// SetCharMap publica os mapas de um charset na versão indicada.
func SetCharMap(version common.GameVersion, charset string, byteToCharMap map[uint]rune, charToByteMap map[rune]uint) {
	v := ensureVersionBucket(version)
	ByteToCharMaps[v][charset] = byteToCharMap
	CharToByteMaps[v][charset] = charToByteMap
}

// GetByteToCharMap retorna o mapa byte->rune do charset na versão indicada (pode ser nil).
func GetByteToCharMap(version common.GameVersion, charset string) map[uint]rune {
	if byCharset, ok := ByteToCharMaps[version]; ok {
		return byCharset[charset]
	}
	return nil
}

// GetCharToByteMap retorna o mapa rune->byte do charset na versão indicada (pode ser nil).
func GetCharToByteMap(version common.GameVersion, charset string) map[rune]uint {
	if byCharset, ok := CharToByteMaps[version]; ok {
		return byCharset[charset]
	}
	return nil
}

// ClearCharMaps limpa os charsets da versão indicada.
func ClearCharMaps(version common.GameVersion) {
	delete(ByteToCharMaps, version)
	delete(CharToByteMaps, version)
}

// ClearAllCharMaps limpa todas as versões (útil em testes).
func ClearAllCharMaps() {
	ByteToCharMaps = make(map[common.GameVersion]map[string]map[uint]rune)
	CharToByteMaps = make(map[common.GameVersion]map[string]map[rune]uint)
}
