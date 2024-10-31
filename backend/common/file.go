package common

import (
	"bufio"
	"ffxresources/backend/models"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	dlgRegex = regexp.MustCompile(`^.*\.bin$`)
	tutorialRegex = regexp.MustCompile(`^.*\.msb$`)
	dcpRegex = regexp.MustCompile(`.*macrodic.*\.dcp$`)
	dcpPartsRegex = regexp.MustCompile(`.*macrodic.*\.00[0-6]$`)
	kernelRegex = regexp.MustCompile(`.*kernel.*\.bin$`)
	lockitRegex = regexp.MustCompile(`.*loc_kit_ps3.*\.bin$`)
	lockitPartsRegex = regexp.MustCompile(`.*loc_kit_ps3.*\.loc_kit_([0-9]{2})$`)
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func RemoveFile(path string) error {
	return os.Remove(path)
}

func ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func guessBySpiraFileType(path string) models.NodeType {
	lowerPath := strings.ToLower(path)

	switch {
	case kernelRegex.MatchString(lowerPath):
		return models.Kernel
	case dcpRegex.MatchString(lowerPath):
		return models.Dcp
	case dcpPartsRegex.MatchString(lowerPath):
		return models.DcpParts
	case lockitRegex.MatchString(lowerPath):
		return models.Lockit
	case lockitPartsRegex.MatchString(lowerPath):
		return models.LockitParts
	case dlgRegex.MatchString(lowerPath):
		return models.Dialogs
	case tutorialRegex.MatchString(lowerPath):
		return models.Tutorial
	default:
		return models.File
	}
}

func ChangeExtension(path, newExt string) string {
	ext := filepath.Ext(path)
	return path[:len(path)-len(ext)] + newExt
}

func AddExtension(path, newExt string) string {
	ext := filepath.Ext(path)
	if ext == newExt {
		return path
	}

	return path + newExt
}

func RemoveFileExtension(filePath string) string {
	ext := filepath.Ext(filePath)
	return filePath[:len(filePath)-len(ext)]
}

func CountSeparators(targetFile string) int {
	separator := FFX_TEXT_FORMAT_SEPARATOR

	input, err := ReadFile(targetFile)
	if err != nil {
		return 0
	}

	return strings.Count(input, separator)
}

func EnsureWindowsFormat(targetFile string, nodeType models.NodeType) error {
	if nodeType == models.Dcp {
		return nil
	}

	file, err := os.Open(targetFile)
	if err != nil {
		return fmt.Errorf("error when opening the file: %s", err)
	}
	defer file.Close()

	text, err := changeLineBreaksToWindows(file)
	if err != nil {
		return fmt.Errorf("error when reading the file: %s", err)
	}

	ensureStartEndLineBreaks(&text)

	//err = WriteFile(fileInfo.TranslateLocation.TargetFile, text)
	err = WriteStringToFile(targetFile, text)
	if err != nil {
		return fmt.Errorf("error saving the modified file: %s", err)
	}

	return nil
}

func changeLineBreaksToWindows(file *os.File) (string, error) {
	var content strings.Builder

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		content.WriteString(scanner.Text() + "\r\n")
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error while reading: %s", err)
	}

	return content.String(), nil
}

func ensureStartEndLineBreaks(text *string) {
	startRegex := regexp.MustCompile(`^\r\n`)
	endRegex := regexp.MustCompile(`\r\n$`)

	if !startRegex.MatchString(*text) {
		*text = "\r\n" + *text
	}

	if !endRegex.MatchString(*text) {
		*text = *text + "\r\n"
	}
}
