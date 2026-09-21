package converter

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"ffxresources/backend/common"
	encoding "ffxresources/backend/core/encoding"
)

var (
	WriteLinebreaksAsCommands = true
)

var (
	reCmd    = regexp.MustCompile(`^CMD:([0-9A-Fa-f]{1,2}):([0-9A-Fa-f]{1,2})`)
	reChoice = regexp.MustCompile(`\{CHOICE:([0-9A-Fa-f]{2})\}`)
	reMCR    = regexp.MustCompile(`^MCR:s([0-9A-Fa-f]{1,2}):l([0-9A-Fa-f]{1,2}):`)
	reHEX    = regexp.MustCompile(`^HEX:([0-9A-Fa-f]{2}(?::[0-9A-Fa-f]{2})*)$`)
)

// Faixa PUA espelhando o esquema de reader/charset.go (puaRune(code) com
// puaBase 0xE000): runes PUA são dados esperados nos textos e voltam ao
// código do jogo pela bijeção, mesmo fora dos mapas.
const (
	puaBase = 0xE000
	puaLast = 0xF8FF
)

// puaCode extrai o código do jogo de uma rune PUA (U+E000+código).
func puaCode(chr rune) (uint, bool) {
	if chr >= puaBase && chr <= puaLast {
		return uint(chr - puaBase), true
	}
	return 0, false
}

func CharToBytes(chr rune, charset string, version common.GameVersion) ([]uint, error) {
	if chr == '\n' {
		return []uint{0x03}, nil
	}

	indexValue, err := encoding.CharToByte(chr, charset, version)
	if err != nil {
		if errors.Is(err, encoding.ErrVersionNotFound) || errors.Is(err, encoding.ErrCharsetNotFound) {
			// erro de configuração: fatal
			return nil, fmt.Errorf("CharToBytes: %w", err)
		}
		if !errors.Is(err, encoding.ErrCharNotFound) {
			return nil, fmt.Errorf("CharToBytes: %w", err)
		}
		// Rune PUA fora dos mapas: dado esperado, deriva o código pela
		// bijeção em vez de descartar (round-trip com o slot fixo).
		code, ok := puaCode(chr)
		if !ok {
			// caractere ausente: retorna erro específico, chamador decide
			return nil, fmt.Errorf("CharToBytes: %w", err)
		}
		indexValue = code
	}

	if indexValue < 0x100 {
		return []uint{indexValue}, nil
	}

	section := (indexValue - 0x30) / 0xD0
	byte1 := section + 0x2B
	byte2 := indexValue - (section * 0xD0)

	if byte1 <= 0x2F {
		return []uint{uint(byte1), uint(byte2)}, nil
	}

	adjustedValue := indexValue - 0x410

	if adjustedValue < 0x100 {
		return []uint{0x04, adjustedValue}, nil
	}
	adjustedSection := (adjustedValue - 0x30) / 0xD0
	adjustedByte1 := adjustedSection + 0x2B
	adjustedByte2 := adjustedValue - (adjustedSection * 0xD0)
	return []uint{0x04, uint(adjustedByte1), uint(adjustedByte2)}, nil
}

func GetChoicesInString(s string) int {
	choices := 0
	for {
		choiceTag := fmt.Sprintf("{CHOICE:%02X}", choices)
		if !strings.Contains(s, choiceTag) {
			break
		}
		choices++
	}
	return choices
}

func GetFirstChoiceInString(s string) (uint16, bool) {
	match := reChoice.FindStringSubmatch(s)
	if len(match) > 1 {
		if val, err := strconv.ParseUint(match[1], 16, 16); err == nil {
			return uint16(val), true
		}
	}
	return 0, false
}

func FillByteList(s string, buf *bytes.Buffer, charset string, version common.GameVersion) {
	version = common.CharsetVersion(version)
	runes := []rune(s)

	for i := 0; i < len(runes); i++ {
		chr := runes[i]
		var cmdBytes []uint

		if chr == '{' {
			cmdBytes = ParseCommand(runes, i)
		}

		if cmdBytes == nil {
			charBytes, err := CharToBytes(chr, charset, version)
			if err != nil {
				if errors.Is(err, encoding.ErrVersionNotFound) || errors.Is(err, encoding.ErrCharsetNotFound) {
					common.LogError("FillByteList: configuração inválida: %v", err)
				} else {
					fmt.Fprintf(os.Stderr, "Unknown character %c at index %d in string %s\n", chr, i, s)
				}
			} else {
				for _, b := range charBytes {
					buf.WriteByte(byte(b))
				}
			}
		} else {
			for _, b := range cmdBytes {
				buf.WriteByte(byte(b))
			}
			i = getRunePosition(runes, '}', i)
		}
	}
	buf.WriteByte(0x00)
}

func StringToByteList(runes []rune, charset string, version common.GameVersion) ([]byte, error) {
	version = common.CharsetVersion(version)
	var buf bytes.Buffer
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		var cmdBytes []uint
		if r == '{' {
			cmdBytes = ParseCommand(runes, i)
		}
		if cmdBytes == nil {
			charBytes, err := CharToBytes(r, charset, version)
			if err != nil {
				if errors.Is(err, encoding.ErrVersionNotFound) || errors.Is(err, encoding.ErrCharsetNotFound) {
					return nil, fmt.Errorf("StringToByteList: configuração inválida: %w", err)
				}
				fmt.Fprintf(os.Stderr, "Unknown character %c at index %d in %q: %v\n", r, i, string(runes), err)
				continue
			}
			for _, b := range charBytes {
				buf.WriteByte(byte(b))
			}
		} else {
			for _, b := range cmdBytes {
				buf.WriteByte(byte(b))
			}
			i = getRunePosition(runes, '}', i)
		}
	}
	return buf.Bytes(), nil
}

func StringToBytes(s, charset string, version common.GameVersion) ([]byte, error) {
	runes := []rune(s)
	gameVersion := common.CharsetVersion(version)
	return StringToByteList(runes, charset, gameVersion)
}

func GetStringBytesAtLookupOffset(table []byte, offset int) []byte {
	if offset < 0 || offset >= len(table) {
		return nil
	}

	end := bytes.IndexByte(table[offset:], 0x00)
	if end == -1 {
		return table[offset:]
	}
	return table[offset : offset+end]
}

func ParseCommand(runes []rune, startIndex int) []uint {
	if startIndex >= len(runes) {
		return nil
	}

	end := getRunePosition(runes, '}', startIndex)
	if end < 0 {
		return nil
	}

	cmd := string(runes[startIndex+1 : end])
	if strings.Contains(cmd, "??}") {
		fmt.Printf("Invalid command format: %s\n", cmd)
		return nil
	}

	switch {
	case cmd == "PAUSE":
		return []uint{0x01}
	case cmd == "BREAK":
		return []uint{0x02}
	case cmd == "\\n" || cmd == "TEXT_NEWLINE":
		return []uint{0x03}
	case cmd == "TEXT_ITALIC":
		return []uint{0x0E, 0x40}
	case cmd == "TEXT_NORMAL":
		return []uint{0x0E, 0x41}
	case cmd == "BLANK05":
		return []uint{0x05}
	case cmd == "BLANK0C":
		return []uint{0x0C}
	case cmd == "BLANK0F":
		return []uint{0x0F}
	case cmd == "BLANK11":
		return []uint{0x11}
	case strings.HasPrefix(cmd, "SPACE:"):
		val, err := strconv.ParseUint(cmd[6:], 16, 8)
		if err != nil {
			return nil
		}
		pixels := val + 0x30
		return []uint{0x07, uint(pixels)}
	case strings.HasPrefix(cmd, "TIME:"):
		val, err := strconv.ParseUint(cmd[5:], 16, 8)
		if err != nil {
			return nil
		}
		boxType := val + 0x30
		return []uint{0x09, uint(boxType)}
	case strings.HasPrefix(cmd, "CLR:"):
		clr := encoding.ColorToByte(cmd[4:])
		return []uint{0x0A, uint(clr)}
	case strings.HasPrefix(cmd, "COLOR:"):
		clr := encoding.ColorToByte(cmd[6:])
		return []uint{0x0A, uint(clr)}
	case strings.HasPrefix(cmd, "ICON:") || strings.HasPrefix(cmd, "BUTTON:"):
		parts := strings.SplitN(cmd, ":", 3)
		if len(parts) != 3 {
			fmt.Printf("Invalid ICON/BUTTON format: %s\n", cmd)
			return nil
		}
		iconIdx, err := strconv.ParseUint(parts[1], 16, 8)
		if err != nil {
			fmt.Printf("Invalid hex value: %s\n", parts[1])
			return nil
		}
		return []uint{0x0B, uint(iconIdx)}
	case cmd == "CHOICE-END":
		return []uint{0x10, 0xFF}
	case strings.HasPrefix(cmd, "CHOICE:"):
		choiceIdx, err := strconv.ParseUint(cmd[7:], 16, 8)
		if err != nil {
			return nil
		}
		choiceIdx += 0x30
		return []uint{0x10, uint(choiceIdx)}
	case strings.HasPrefix(cmd, "VAR:"):
		val, err := strconv.ParseUint(cmd[4:], 16, 8)
		if err != nil {
			return nil
		}
		varIdx := val + 0x30
		return []uint{0x12, uint(varIdx)}
	case strings.HasPrefix(cmd, "PC:"):
		parts := strings.SplitN(cmd, ":", 3)
		if len(parts) != 3 {
			fmt.Printf("Invalid PC format: %s\n", cmd)
			return nil
		}
		val, err := strconv.ParseUint(parts[1], 16, 8)
		if err != nil {
			fmt.Printf("Invalid hex value: %s\n", parts[1])
			return nil
		}
		pc := val + 0x30
		return []uint{0x13, uint(pc)}
	case strings.HasPrefix(cmd, "MCR:"):
		matches := reMCR.FindStringSubmatch(cmd)
		if len(matches) != 3 {
			fmt.Printf("Invalid MCR format: %s\n", cmd)
			return nil
		}
		secVal, err1 := strconv.ParseUint(matches[1], 16, 8)
		lineVal, err2 := strconv.ParseUint(matches[2], 16, 8)
		if err1 != nil || err2 != nil {
			return nil
		}
		section := secVal + 0x13
		line := lineVal + 0x30
		return []uint{uint(section), uint(line)}
	case strings.HasPrefix(cmd, "MACRO:"):
		if len(cmd) < 12 {
			return nil
		}
		secVal, err1 := strconv.ParseUint(cmd[7:9], 16, 8)
		lineVal, err2 := strconv.ParseUint(cmd[10:12], 16, 8)
		if err1 != nil || err2 != nil {
			return nil
		}
		section := secVal + 0x13
		line := lineVal + 0x30
		return []uint{uint(section), uint(line)}
	case strings.HasPrefix(cmd, "KEY:"):
		val, err := strconv.ParseUint(cmd[4:6], 16, 8)
		if err != nil {
			return nil
		}
		keyItemIdx := val + 0x30
		return []uint{0x23, uint(keyItemIdx)}
	case strings.HasPrefix(cmd, "CMD:"):
		matches := reCmd.FindStringSubmatch(cmd)
		if len(matches) != 3 {
			fmt.Printf("Invalid CMD format: %s\n", cmd)
			return nil
		}
		cmdIdxVal, err1 := strconv.ParseUint(matches[1], 16, 8)
		argVal, err2 := strconv.ParseUint(matches[2], 16, 8)
		if err1 != nil || err2 != nil {
			return nil
		}
		arg := argVal + 0x30
		return []uint{uint(cmdIdxVal), uint(arg)}
	case strings.HasPrefix(cmd, "UNKCHR:"):
		val, err := strconv.ParseUint(cmd[7:9], 16, 8)
		if err != nil {
			return nil
		}
		return []uint{uint(val)}
	case strings.HasPrefix(cmd, "UNKDBLCHR:"):
		secVal, err1 := strconv.ParseUint(cmd[10:12], 16, 8)
		idxVal, err2 := strconv.ParseUint(cmd[13:15], 16, 8)
		if err1 != nil || err2 != nil {
			return nil
		}
		return []uint{uint(secVal), uint(idxVal)}
	case strings.HasPrefix(cmd, "HEX:"):
		matches := reHEX.FindStringSubmatch(cmd)
		if len(matches) != 2 {
			fmt.Printf("Invalid HEX format: %s\n", cmd)
			return nil
		}
		hexParts := strings.Split(matches[1], ":")
		result := make([]uint, len(hexParts))
		for i, part := range hexParts {
			if len(part) != 2 {
				fmt.Printf("Invalid HEX part: %s in command %s\n", part, cmd)
				return nil
			}
			val, _ := strconv.ParseUint(part, 16, 8)
			result[i] = uint(val)
		}
		return result
	default:
		return nil
	}
}

func getRunePosition(runes []rune, target rune, start int) int {
	for i := start; i < len(runes); i++ {
		if runes[i] == target {
			return i
		}
	}
	return -1
}
