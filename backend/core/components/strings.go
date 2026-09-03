package components

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	WriteLinebreaksAsCommands = true
)

var (
	reCmd    = regexp.MustCompile(`^CMD:([0-9A-Fa-f]{1,2}):([0-9A-Fa-f]{1,2})`)
	reChoice = regexp.MustCompile(`\{CHOICE:([0-9A-Fa-f]{2})\}`)
	reMCR    = regexp.MustCompile(`^MCR:s([0-9A-Fa-f]{1,2}):l([0-9A-Fa-f]{1,2}):`)
	reHEX    = regexp.MustCompile(`^HEX:([0-9A-Fa-f]{2}(?::[0-9A-Fa-f]{2})*)$`)
	rePC     = regexp.MustCompile(`^PC:([0-9A-Fa-f]{1,2}):`)
	reICON   = regexp.MustCompile(`^ICON:([0-9A-Fa-f]{1,2}):`)
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

// FillByteList processes a string and fills the provided byte list with the converted bytes
// This function appends to the existing byte list if it has content, or starts fresh if empty
//
// Parameters:
//   - s: The string to process (may contain command tags like {MCR:...}, {CLR:...}, etc.)
//   - byteList: Pointer to the byte slice to fill/append to
//   - charset: The charset to use for character conversion (e.g., "jp", "us", "kr")
//
// Behavior:
//   - Processes each character in the string
//   - Handles command tags by parsing them with ParseCommand
//   - Converts regular characters using CharToBytes
//   - Appends all bytes to the provided byteList
//   - Adds null terminator (0x00) at the end
func FillByteList(s string, buf *bytes.Buffer, charset string) {
	runes := []rune(s)

	for i := 0; i < len(runes); i++ {
		chr := runes[i]
		var cmdBytes []uint

		if chr == '{' {
			cmdBytes = ParseCommand(runes, i)
		}

		if cmdBytes == nil {
			charBytes := CharToBytes(chr, charset)
			if charBytes != nil {
				for _, b := range charBytes {
					buf.WriteByte(byte(b))
				}
			} else {
				fmt.Fprintf(os.Stderr, "Unknown character %c at index %d in string %s\n", chr, i, s)
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

// StringToByteList converts a slice of runes to a byte slice using the specified character encoding.
// It processes each rune in the input slice, handling special command sequences enclosed in curly braces.
//
// Parameters:
//   - runes: slice of runes to be converted to bytes
//   - charset: string specifying the character encoding to use for conversion
//
// Returns:
//   - []byte: the resulting byte slice after conversion
//
// Behavior:
//   - Regular runes are converted to bytes using the specified charset via CharToBytes
//   - Command sequences starting with '{' are parsed using ParseCommand and their resulting bytes are included
//   - When a command is found, the function skips to the closing '}' using getRunePosition
//   - Unknown characters that cannot be converted are reported to stderr with their position and context
//   - All resulting bytes are accumulated in a buffer and returned as a single byte slice
func StringToByteList(runes []rune, charset string) []byte {
	var buf bytes.Buffer
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		var cmdBytes []uint
		if r == '{' {
			cmdBytes = ParseCommand(runes, i)
		}
		if cmdBytes == nil {
			charBytes := CharToBytes(r, charset)
			if charBytes != nil {
				for _, b := range charBytes {
					buf.WriteByte(byte(b))
				}
			} else {
				fmt.Fprintf(os.Stderr, "Unknown character %c at index %d in string %s\n", r, i, string(runes))
			}
		} else {
			for _, b := range cmdBytes {
				buf.WriteByte(byte(b))
			}
			i = getRunePosition(runes, '}', i)
		}
	}
	return buf.Bytes()
}

func StringToBytes(s, charset string) []byte {
	runes := []rune(s)
	return StringToByteList(runes, charset)
}

// GetStringBytesAtLookupOffset retrieves a null-terminated string from a byte table
// starting at the specified offset. The function reads bytes from the offset position
// until it encounters a null byte (0x00) or reaches the end of the table.
//
// Parameters:
//   - table: The byte slice containing the string data
//   - offset: The starting position in the table to begin reading from
//
// Returns:
//   - []byte: The string bytes without the null terminator, or nil if offset is invalid
//   - If no null terminator is found, returns all bytes from offset to end of table
//   - Returns nil if offset is negative or beyond the table length
//   - For directories: Recursively processes all non-hidden files in sorted order
//   - For files: Resolves path, reads bytes, and parses as string data using appropriate charset
func GetStringBytesAtLookupOffset(table []byte, offset int) []byte {
	if offset < 0 || offset >= len(table) {
		return nil
	}

	end := bytes.IndexByte(table[offset:], 0x00)
	if end == -1 {
		return table[offset:]
	}
	subArray := table[offset : offset+end]
	var newArray = make([]byte, len(subArray))
	copy(newArray, subArray)
	return newArray
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
	case cmd == "\\n":
		return []uint{0x03}
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
		clr := ColorToByte(cmd[4:])
		return []uint{0x0A, uint(clr)}
	case strings.HasPrefix(cmd, "COLOR:"):
		clr := ColorToByte(cmd[6:])
		return []uint{0x0A, uint(clr)}
	case strings.HasPrefix(cmd, "ICON:"):
		// Divide em até 3 partes: ["ICON", "hex", "identificador"]
		parts := strings.SplitN(cmd, ":", 3)
		if len(parts) != 3 {
			fmt.Printf("Invalid ICON format: %s\n", cmd)
			return nil
		}
		// parts[0] = "ICON", parts[1] = hex, parts[2] = identificador (ex: "SQUARE")
		iconIdx, err := strconv.ParseUint(parts[1], 16, 8)
		if err != nil {
			fmt.Printf("Invalid hex value: %s\n", parts[1])
			return nil
		}
		// Se quiser usar o identificador (opcional):
		// ident := parts[2]
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
		// parts[0] = "PC", parts[1] = hex, parts[2] = identificador (ex: "IFRIT")
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
