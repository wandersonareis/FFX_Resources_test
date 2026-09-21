package ffxencoding

import (
	"errors"
	"fmt"
	"sort"

	"ffxresources/backend/common"
)

// Sentinelas da distinção entre mapa ausente (configuração: aborta a
// tarefa) e caractere ausente (dado: pula e continua). Propagam com %w
// para inspeção via errors.Is em cada camada.
var (
	ErrVersionNotFound = errors.New("version não encontrada nos mapas de charset")
	ErrCharsetNotFound = errors.New("charset não encontrado para essa version")
	ErrCharNotFound    = errors.New("caractere não encontrado no charset")
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
// Mapa/charset ausente (configuração) e caractere ausente (dado) são
// distinguidos via ErrVersionNotFound/ErrCharsetNotFound/ErrCharNotFound.
func ByteToChar(hex uint, charset string, version common.GameVersion) (rune, error) {
	byCharset, ok := ByteToCharMaps[version]
	if !ok {
		return 0, fmt.Errorf("%w: %v", ErrVersionNotFound, version)
	}
	charsetMap, ok := byCharset[charset]
	if !ok {
		return 0, fmt.Errorf("%w: %q (version=%v)", ErrCharsetNotFound, charset, version)
	}

	char, ok := charsetMap[hex]
	if !ok {
		return 0, fmt.Errorf("%w: 0x%02X (charset=%q)", ErrCharNotFound, hex, charset)
	}
	return char, nil
}

// CharToByte resolve uma rune para byte no charset da versão indicada.
func CharToByte(chr rune, charset string, version common.GameVersion) (uint, error) {
	byCharset, ok := CharToByteMaps[version]
	if !ok {
		return 0, fmt.Errorf("%w: %v", ErrVersionNotFound, version)
	}
	charsetMap, ok := byCharset[charset]
	if !ok {
		return 0, fmt.Errorf("%w: %q (version=%v)", ErrCharsetNotFound, charset, version)
	}
	b, ok := charsetMap[chr]
	if !ok {
		return 0, fmt.Errorf("%w: %q (U+%04X)", ErrCharNotFound, chr, chr)
	}
	return b, nil
}

// EnsureCharsetLoaded verifica (só leitura, sem criar buckets) que os mapas
// da versão/charset existem. É o fail-fast do load: mapa ausente aborta a
// tarefa antes de produzir milhares de placeholders. A versão passa por
// common.CharsetVersion (lastmiss usa os mapas de ffx2).
func EnsureCharsetLoaded(version common.GameVersion, charset string) error {
	version = common.CharsetVersion(version)
	byCharset, ok := ByteToCharMaps[version]
	if !ok {
		return fmt.Errorf("%w: %v", ErrVersionNotFound, version)
	}
	if _, ok := byCharset[charset]; !ok {
		return fmt.Errorf("%w: %q (version=%v)", ErrCharsetNotFound, charset, version)
	}
	byReverse, ok := CharToByteMaps[version]
	if !ok {
		return fmt.Errorf("%w: %v", ErrVersionNotFound, version)
	}
	if _, ok := byReverse[charset]; !ok {
		return fmt.Errorf("%w: %q (version=%v)", ErrCharsetNotFound, charset, version)
	}
	return nil
}

// EnsureAllCharsetsLoaded fail-fast para todos os charsets das localizações
// suportadas (ordem determinística). Chamado na entrada dos loads
// (events/objects/macro): mapa ausente aborta a tarefa antes de produzir
// milhares de placeholders.
func EnsureAllCharsetsLoaded(version common.GameVersion) error {
	locs := make([]string, 0, len(common.SupportedLanguages))
	for loc := range common.SupportedLanguages {
		locs = append(locs, loc)
	}
	sort.Strings(locs)
	for _, loc := range locs {
		if err := EnsureCharsetLoaded(version, GetCharsetForLanguage(loc)); err != nil {
			return err
		}
	}
	return nil
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
