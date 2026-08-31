package converter

import (
	"bytes"
	"encoding/binary"
	"ffxresources/backend/datastore"
	"ffxresources/backend/sharedutils"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	reCmd    = regexp.MustCompile(`^CMD:([0-9A-Fa-f]{1,2}):([0-9A-Fa-f]{1,2})`)
	reChoice = regexp.MustCompile(`\{CHOICE:([0-9A-Fa-f]{2})\}`)
	reMCR    = regexp.MustCompile(`^MCR:s([0-9A-Fa-f]{1,2}):l([0-9A-Fa-f]{1,2}):`)
	reHEX    = regexp.MustCompile(`^HEX:([0-9A-Fa-f]{2}(?::[0-9A-Fa-f]{2})*)$`)
	rePC     = regexp.MustCompile(`^PC:([0-9A-Fa-f]{1,2}):`)
	reICON   = regexp.MustCompile(`^ICON:([0-9A-Fa-f]{1,2}):`)
)

func readOneByte(buf *bytes.Reader, out *byte) error {
	return binary.Read(buf, binary.LittleEndian, out)
}

func getRunePosition(runes []rune, target rune, start int) int {
	for i := start; i < len(runes); i++ {
		if runes[i] == target {
			return i
		}
	}
	return -1
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
	case cmd == "\\n":
		return []uint{0x03}
	case cmd == "CMD04":
		return []uint{0x04}
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
		clr := sharedutils.ColorToByte(cmd[4:])
		return []uint{0x0A, uint(clr)}
	case strings.HasPrefix(cmd, "COLOR:"):
		clr := sharedutils.ColorToByte(cmd[6:])
		return []uint{0x0A, uint(clr)}
	case strings.HasPrefix(cmd, "ICON:"):
		matches := reICON.FindStringSubmatch(cmd)
		if len(matches) != 2 {
			fmt.Printf("Invalid ICON format: %s\n", cmd)
			return nil
		}
		ctrlIdx, err := strconv.ParseUint(matches[1], 16, 8)
		if err != nil {
			return nil
		}
		return []uint{0x0B, uint(ctrlIdx)}
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
		matches := rePC.FindStringSubmatch(cmd)
		if len(matches) != 2 {
			fmt.Printf("Invalid PC format: %s\n", cmd)
			return nil
		}
		val, err := strconv.ParseUint(matches[1], 16, 8)
		if err != nil {
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
			charBytes := sharedutils.CharToBytes(r, charset)
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

func getStringAtLookupOffsetBinary(table []byte, offset int, localization string) string {
	if offset < 0 || offset >= len(table) {
		return ""
	}

	var (
		out               strings.Builder
		charset           = sharedutils.GetCharsetForLanguage(localization)
		extraFiveSections bool
		buf               = bytes.NewReader(table[offset:])
	)

	for {
		var idx uint8
		err := readOneByte(buf, &idx)
		if err != nil || idx == 0x00 {
			break
		}

		var extraOffset uint = 0
		if extraFiveSections {
			extraOffset = 0x410
			extraFiveSections = false
		}

		switch {
		case idx >= 0x30:
			if chr, ok := sharedutils.ByteToChar(uint(idx)+extraOffset, charset); ok {
				out.WriteRune(chr)
			}
		case idx == 0x01:
			out.WriteString("{PAUSE}")
		case idx == 0x03:
			if sharedutils.WriteLinebreaksAsCommands {
				out.WriteString("{\\n}")
			} else {
				out.WriteByte('\n')
			}
		case buf.Len() == 0:
			out.WriteString(fmt.Sprintf("{HEX:%02X}", idx))
		case idx == 0x04:
			extraFiveSections = true
		case idx == 0x07:
			var pixels uint8
			if err := readOneByte(buf, &pixels); err != nil {
				out.WriteString("{SPACE:??}")
				break
			}
			out.WriteString(fmt.Sprintf("{SPACE:%02X}", pixels-0x30))
		case idx == 0x09:
			var varIdx uint8
			if err := readOneByte(buf, &varIdx); err != nil {
				out.WriteString("{TIME:??}")
				break
			}
			out.WriteString(fmt.Sprintf("{TIME:%02X}", varIdx-0x30))
		case idx == 0x0A:
			var clr uint8
			if err := readOneByte(buf, &clr); err != nil {
				out.WriteString("{CLR:??}")
				break
			}
			out.WriteString(sharedutils.GetColorString(clr))
		case idx == 0x0B:
			var icon uint8
			if err := readOneByte(buf, &icon); err != nil {
				out.WriteString("{ICON:??}")
				break
			}
			out.WriteString(fmt.Sprintf("{ICON:%02X:%s}", icon, sharedutils.GetIconName(icon)))
		case idx == 0x10:
			var rawValue uint8
			if err := readOneByte(buf, &rawValue); err != nil {
				if err == io.EOF {
					out.WriteString(fmt.Sprintf("{HEX:%02X}", idx))
				} else {
					out.WriteString("{CHOICE:??}")
				}
				break
			}
			if rawValue == 0xFF {
				out.WriteString("{CHOICE-END}")
				break
			}
			choiceIdx := rawValue - 0x30
			out.WriteString(fmt.Sprintf("{CHOICE:%02X}", choiceIdx))
		case idx == 0x12:
			var varIdx uint8
			if err := readOneByte(buf, &varIdx); err != nil {
				out.WriteString("{VAR:??}")
				break
			}
			out.WriteString(fmt.Sprintf("{VAR:%02X}", varIdx-0x30))
		case idx == 0x13 && buf.Len() > 0:
			var rawValue uint8
			if err := readOneByte(buf, &rawValue); err != nil || rawValue > 0x43 {
				out.WriteString(fmt.Sprintf("{PC:%02X:??}", rawValue))
				break
			}
			pcIdx := rawValue - 0x30
			out.WriteString(fmt.Sprintf("{PC:%02X:%s}", pcIdx, sharedutils.GetPlayerChar(pcIdx)))
		case idx >= 0x13 && idx <= 0x22:
			var section uint = uint(idx) - 0x13
			var line byte
			if err := readOneByte(buf, &line); err != nil {
				fmt.Println("Error reading line number for MCR command:", err)
				out.WriteString(fmt.Sprintf("{HEX:%02X}", idx))
				break
			}
			lineAdjusted := line - 0x30
			out.WriteString(fmt.Sprintf("{MCR:s%02X:l%02X", section, lineAdjusted))
			if datastore.HasMacros() {
				out.WriteString(":")
				index := int(section*0x100 + uint(lineAdjusted))
				if macro, exist := datastore.GetMacro(index); exist && macro != nil {
					if macroContent, exists := macro.GetLocalizedContent(localization); exists && macroContent != nil {
						out.WriteString(`"`)
						localizedContent, _ := macro.GetLocalizedContent(localization)
						out.WriteString(localizedContent.String())
						out.WriteString(`"`)
					} else {
						out.WriteString("<Missing>")
					}
				} else {
					out.WriteString("<Missing>")
				}
			}
			out.WriteString("}")
		case idx == 0x23:
			var varIdx uint8
			if err := readOneByte(buf, &varIdx); err != nil {
				out.WriteString("{KEY:??}")
				break
			}
			varIdx -= 0x30
			out.WriteString(fmt.Sprintf("{KEY:%02X", varIdx))
			if keyItem := datastore.KeyItems.Get(int(varIdx)); keyItem != nil {
				out.WriteString(fmt.Sprintf(`:"%s"`, keyItem.GetName(localization)))
			}
			out.WriteString("}")
		case idx == 0x28:
			var val uint8
			if err := readOneByte(buf, &val); err != nil {
				out.WriteString("{CMD:28:??}")
				break
			}
			out.WriteString(fmt.Sprintf("{CMD:28:%02X}", val-0x30))
		case idx == 0x2A:
			var val byte
			if err := readOneByte(buf, &val); err != nil {
				out.WriteString("{CMD:2A:??}")
				break
			}
			out.WriteString(fmt.Sprintf("{CMD:2A:%02X}", val-0x30))

		case idx >= 0x2B: // Double-byte character handling
			section := uint(idx) - 0x2B
			var low byte
			if err := readOneByte(buf, &low); err != nil {
				out.WriteString(fmt.Sprintf("{UNKDBLCHR:%02X:??}", idx))
				break
			}

			actualIdx := section*0xD0 + uint(low)
			newVar := actualIdx + extraOffset

			if chr, ok := sharedutils.ByteToChar(newVar, charset); ok {
				out.WriteRune(chr)
			} else {
				out.WriteString(fmt.Sprintf("{UNKDBLCHR:%02X:%02X}", idx, low))
			}
		default:
			var nextByte byte
			if err := readOneByte(buf, &nextByte); err != nil {
				out.WriteString(fmt.Sprintf("{CMD:%02X:??}", idx))
				break
			}
			out.WriteString(fmt.Sprintf("{CMD:%02X:%02X}", idx, nextByte-0x30))
		}
	}

	return out.String()
}

func BytesToString(rawData []byte, localization string) string {
	return getStringAtLookupOffsetBinary(rawData, 0, localization)
}
