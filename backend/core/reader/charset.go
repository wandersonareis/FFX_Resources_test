package reader

import (
	"encoding/json"
	"ffxresources/backend/common"
	"ffxresources/backend/sharedutils"
	"fmt"
	"path/filepath"
)

func PrepareCharset(charset string) error {
	replacements := loadCharReplacements(charset)

	path := filepath.Join(common.GetPathOriginalsRoot(), common.GetEncodingPath(charset))
	filePath, err := common.NewFileAccessor(path)
	if err != nil {
		return err
	}
	data := filePath.ReadBytes()
	if data == nil {
		return err
	}

	str := string(data)
	runes := []rune(str)
	if len(replacements) > 0 {
		runes = applyReplacements(runes, replacements)
	}

	byteToChar, charToByte := buildMappings(runes)
	sharedutils.SetCharMap(charset, byteToChar, charToByte)
	return nil
}

func loadCharReplacements(charset string) map[rune]rune {
	path := filepath.Join(common.GetPathOriginalsRoot(), "ffx2_encoding", "char_replacements.json")
	resolvedFile, err := common.NewFileAccessor(path)
	if err != nil {
		fmt.Printf("Error resolving char replacements file: %v\n", err)
		return nil
	}
	data, err := common.ReadFile(resolvedFile.ResolvedPath)
	if err != nil {
		return nil
	}

	var config map[string]map[string]string
	if err := json.Unmarshal(data, &config); err != nil {
		return nil
	}
	charsetReplacements, ok := config[charset]
	if !ok {
		return nil
	}

	replacements := make(map[rune]rune)
	for oldChar, newChar := range charsetReplacements {
		oldCharRune := []rune(oldChar)
		newCharRune := []rune(newChar)
		if len(oldCharRune) == 1 && len(newCharRune) == 1 {
			replacements[oldCharRune[0]] = newCharRune[0]
		} else {
			fmt.Printf("Warning: Invalid replacement for charset %s: %s -> %s\n", charset, oldChar, newChar)
		}
	}
	return replacements
}

func applyReplacements(runes []rune, replacements map[rune]rune) []rune {
	if len(replacements) == 0 {
		return runes
	}

	newRunes := make([]rune, len(runes))
	copy(newRunes, runes)

	for i, r := range newRunes {
		if newChar, ok := replacements[r]; ok {
			newRunes[i] = newChar
		}
	}
	return newRunes
}

func buildMappings(runes []rune) (map[uint]rune, map[rune]uint) {
	byteToChar := make(map[uint]rune)
	charToByte := make(map[rune]uint)
	for i, r := range runes {
		idx := uint(i + 0x30) // Start at 0x30
		byteToChar[idx] = r
		if _, exists := charToByte[r]; !exists {
			charToByte[r] = idx
		}
	}
	return byteToChar, charToByte
}