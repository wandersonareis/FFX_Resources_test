package components

import (
	"bytes"
	"encoding/binary"
	"ffxresources/backend/common"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	localizationMap = map[string]string{
		"ch": "ch",
		"kr": "kr",
		"jp": "jp",
	}
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
	playerCharMap = map[byte]string{
		0x00: "TIDUS",
		0x01: "YUNA",
		0x02: "AURON",
		0x03: "KIMAHRI",
		0x04: "WAKKA",
		0x05: "LULU",
		0x06: "RIKKU",
		0x07: "SEYMOUR",
		0x08: "VALEFOR",
		0x09: "IFRIT",
		0x0A: "IXION",
		0x0B: "SHIVA",
		0x0C: "BAHAMUT",
		0x0D: "ANIMA",
		0x0E: "YOJIMBO",
		0x0F: "CINDY",
		0x10: "SANDY",
		0x11: "MINDY",
		0x12: "DUMMY",
		0x13: "DUMMY2",
	}
	controllerInputMap = map[byte]string{
		0x20: "?L1 (SWITCH)",
		0x30: "TRIANGLE",
		0x31: "X",
		0x32: "CIRCLE",
		0x33: "SQUARE",
		0x34: "L1",
		0x35: "R1",
		0x37: "?R2",
		0x39: "SELECT",
		0x41: "UP",
		0x42: "RIGHT",
		0x44: "DOWN",
		0x48: "LEFT",
	}
	choiceRegex               = regexp.MustCompile(`\{CHOICE:([0-9A-Fa-f]{2})\}`)
	WriteLinebreaksAsCommands = true
)

var (
	ByteToCharMaps = make(map[string]map[uint]rune)
	CharToByteMaps = make(map[string]map[rune]uint)
	MacroLookup    = make(map[int]*LocalizedMacroStringObject)
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

func LocalizationToCharset(localization string) string {
	if charset, ok := localizationMap[localization]; ok {
		return charset
	}
	return "us"
}

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

func GetPlayerChar(pc byte) string {
	if name, ok := playerCharMap[pc]; ok {
		return name
	}
	return "?"
}

func GetControllerInput(ctrlIdx byte) string {
	if input, ok := controllerInputMap[ctrlIdx]; ok {
		return input
	}
	return "?"
}

func GetColorString(hex uint8) string {
	return fmt.Sprintf("{CLR:%s}", ByteToColor(hex))
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
	match := choiceRegex.FindStringSubmatch(s)
	if len(match) > 1 {
		if val, err := strconv.ParseUint(match[1], 16, 16); err == nil {
			return uint16(val), true
		}
	}
	return 0, false
}

func SetCharMap(charset string, byteToCharMap map[uint]rune, charToByteMap map[rune]uint) {
	ByteToCharMaps[charset] = byteToCharMap
	CharToByteMaps[charset] = charToByteMap
}

func BytesToString(bytes []byte, localization string) string {
	return getStringAtLookupOffsetBinary(bytes, 0, localization)
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

			// Skip to closing brace
			i = getRunePosition(runes, '}', i)
		}
	}

	// Add null terminator
	buf.WriteByte(0x00)
}

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
			// Skip to closing brace
			for j := i + 1; j < len(runes); j++ {
				if runes[j] == '}' {
					i = j
					break
				}
			}
		}
	}
	return buf.Bytes()
}

func StringToBytes(s, charset string) []byte {
	runes := []rune(s)
	return StringToByteList(runes, charset)
}

func GetStringBytesAtLookupOffset(table []byte, offset int) []byte {
	if offset < 0 || offset >= len(table) {
		return nil
	}

	var bytes []byte
	for offset < len(table) && table[offset] != 0x00 {
		bytes = append(bytes, table[offset])
		offset++
	}
	return bytes
}

func readOneByte(buf *bytes.Reader, out *byte) error {
	return binary.Read(buf, binary.LittleEndian, out)
}

func getStringAtLookupOffsetBinary(table []byte, offset int, localization string) string {
	if offset < 0 || offset >= len(table) {
		return ""
	}

	var (
		out               strings.Builder
		charset           = LocalizationToCharset(localization)
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
			if chr, ok := ByteToChar(uint(idx)+extraOffset, charset); ok {
				out.WriteRune(chr)
			}
		case idx == 0x01:
			out.WriteString("{PAUSE}")
		case idx == 0x03:
			if WriteLinebreaksAsCommands {
				out.WriteString("{\\n}")
			} else {
				out.WriteByte('\n')
			}
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
			out.WriteString(GetColorString(clr))
		case idx == 0x0B:
			var ctrlIdx uint8
			if err := readOneByte(buf, &ctrlIdx); err != nil {
				out.WriteString("{CTRL:??}")
				break
			}
			out.WriteString(fmt.Sprintf("{CTRL:%02X:%s}", ctrlIdx, GetControllerInput(ctrlIdx)))
		case idx == 0x10:
			var rawValue uint8
			if err := readOneByte(buf, &rawValue); err != nil {
				out.WriteString("{CHOICE:??}")
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
				out.WriteString("{PC:??}")
				break
			}
			pcIdx := rawValue - 0x30
			out.WriteString(fmt.Sprintf("{PC:%02X:%s}", pcIdx, GetPlayerChar(pcIdx)))
		case idx >= 0x13 && idx <= 0x22:
			var section uint = uint(idx) - 0x13
			var line byte
			if err := readOneByte(buf, &line); err != nil {
				fmt.Println("Error reading line number for MCR command:", err)
				fmt.Println("Byte slice length:", buf.Len())
				out.WriteString(fmt.Sprintf("{MCR:s%02X:l??}", section))
				break
			}
			lineAdjusted := line - 0x30
			out.WriteString(fmt.Sprintf("{MCR:s%02X:l%02X", section, lineAdjusted))
			if len(MacroLookup) > 0 {
				out.WriteString(":")
				index := int(section*0x100 + uint(lineAdjusted))
				if macro, ok := MacroLookup[index]; ok {
					out.WriteString(`"`)
					out.WriteString(macro.GetLocalizedContent(localization).String())
					out.WriteString(`"`)
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
			if keyItem := GetKeyItem(int(varIdx) + 0xA000); keyItem != nil {
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

			if chr, ok := ByteToChar(newVar, charset); ok {
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

var (
	reCmd  = regexp.MustCompile(`^CMD:([0-9A-Fa-f]{1,2}):([0-9A-Fa-f]{1,2})`)
	reMCR  = regexp.MustCompile(`^MCR:s([0-9A-Fa-f]{1,2}):l([0-9A-Fa-f]{1,2}):`)
	rePC   = regexp.MustCompile(`^PC:([0-9A-Fa-f]{1,2}):`)
	reCTRL = regexp.MustCompile(`^CTRL:([0-9A-Fa-f]{1,2}):`)
)

func ParseCommand(runes []rune, startIndex int) []uint {
	if startIndex >= len(runes) {
		return nil
	}

	end := getRunePosition(runes, '}', startIndex)
	if end < 0 {
		return nil
	}

	cmd := string(runes[startIndex+1 : end])

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
		clr := ColorToByte(cmd[4:])
		return []uint{0x0A, uint(clr)}
	case strings.HasPrefix(cmd, "COLOR:"):
		clr := ColorToByte(cmd[6:])
		return []uint{0x0A, uint(clr)}
	case strings.HasPrefix(cmd, "CTRL:"):
		matches := reCTRL.FindStringSubmatch(cmd)
		if len(matches) != 2 {
			fmt.Printf("Invalid CTRL format: %s\n", cmd)
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

// ReadStringFile reads string file(s) from the given filename path
// If the path is a directory, it recursively reads all files within it
// Returns a slice of FieldString objects parsed from the file data
//
// Parameters:
//   - filename: Path to file or directory to read
//   - print: If true, prints debug information during processing
//   - localization: Localization code (e.g., "jp", "us", "kr") for charset conversion
//
// Returns:
//   - []*FieldString: Slice of parsed FieldString objects, or nil if directory or error
//
// Behavior:
//   - For directories: Recursively processes all non-hidden files in sorted order
//   - For files: Resolves path, reads bytes, and parses as string data using appropriate charset
func ReadStringFile(filename string, print bool, localization string) []*FieldString {
	resolvedPath, err := ResolveFile(filename, print)
	if err != nil {
		if print {
			fmt.Printf("Error resolving file %s: %v\n", filename, err)
		}
		return nil
	}

	info, err := os.Stat(resolvedPath)
	if err != nil {
		if print {
			fmt.Printf("Error accessing path %s: %v\n", resolvedPath, err)
		}
		return nil
	}

	// If it's a directory, recursively process all files
	if info.IsDir() {
		entries, err := os.ReadDir(resolvedPath)
		if err != nil {
			if print {
				fmt.Printf("Error reading directory %s: %v\n", resolvedPath, err)
			}
			return nil
		}

		var validFiles []string
		for _, entry := range entries {
			if !strings.HasPrefix(entry.Name(), ".") {
				validFiles = append(validFiles, entry.Name())
			}
		}
		sort.Strings(validFiles)

		for _, file := range validFiles {
			fullPath := filepath.Join(filename, file)
			ReadStringFile(fullPath, print, localization)
		}
		return nil
	}

	bytes := FileToBytes(resolvedPath, print)
	if bytes == nil {
		if print {
			fmt.Printf("Failed to read bytes from file %s\n", resolvedPath)
		}
		return nil
	}

	charset := LocalizationToCharset(localization)
	fieldStrings, err := FromFieldStringData(bytes, print, charset)
	if err != nil {
		if print {
			fmt.Printf("Error parsing string data from %s: %v\n", filename, err)
		}
		return nil
	}

	return fieldStrings
}

// ReadLocalizedStringFiles reads localized string files for all available localizations
// This function iterates through all localizations and reads string files for each one
//
// Parameters:
//   - path: Relative path to the string file (e.g., "event/obj_ps3/XX/XXXX/XXXX.bin")
//
// Returns:
//   - []*LocalizedFieldStringObject: Slice of localized string objects with content for each localization
//
// Behavior:
//   - Iterates through all localizations defined in common.Localizations
//   - For each localization, constructs full path using GetLocalizationRoot + path
//   - Reads string files using ReadStringFile
//   - Merges all localized content into LocalizedFieldStringObject instances
//   - Each index in the returned slice contains all localizations for that string
func ReadLocalizedStringFiles(path string) []*LocalizedFieldStringObject {
	localized := make([]*LocalizedFieldStringObject, 0)

	// Iterate through all available localizations
	for key := range common.Localizations {
		// Build full path for this localization
		fullPath := GetLocalizationRoot(key) + path
		// Read string file for this localization
		localizedStrings := ReadStringFile(fullPath, false, key)

		// Process each string in the file
		for i, fieldString := range localizedStrings {
			// Ensure we have enough LocalizedFieldStringObject instances
			for len(localized) <= i {
				localized = append(localized, NewLocalizedFieldStringObject())
			}

			// Set the localized content for this localization
			localized[i].SetLocalizedContent(key, fieldString)
		}
	}

	return localized
}

// GetLocalizationRoot returns the root path for a given localization
// This function needs to be imported from the reader package or implemented here
func GetLocalizationRoot(localization string) string {
	return common.PathFfxRoot + "new_" + localization + "pc/"
}
