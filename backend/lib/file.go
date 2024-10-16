package lib

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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

func WriteFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

func WriteBytesToFile(file string, data []byte) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write(data)

	return err
}

func guessBySpiraFileType(path string) NodeType {
	extension := strings.ToLower(filepath.Ext(path))

	switch extension {
	case ".bin":
		if strings.Contains(path, "kernel") {
			return Kernel
		}
		return Dialogs
	case ".msb":
		return Tutorial
	case ".dcp":
		return Dcp
	case ".000", ".001", ".002", ".003", ".004", ".005", ".006":
		return DcpParts
	default:
		return File
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

func CountSeparators(fileInfo *FileInfo) int {
	separator := FFX_TEXT_FORMAT_SEPARATOR
	input, err := ReadFile(fileInfo.ExtractLocation.TargetFile)
	if err != nil {
		return 0
	}

	return strings.Count(input, separator)
}

func EnsureWindowsFormat(fileInfo *FileInfo) error {
	if fileInfo.Type == Dcp {
		return nil
	}

	file, err := os.Open(fileInfo.TranslateLocation.TargetFile)
	if err != nil {
		return fmt.Errorf("error when opening the file: %s", err)
	}
	defer file.Close()

	text, err := changeLineBreaksToWindows(file)
	if err != nil {
		return fmt.Errorf("error when reading the file: %s", err)
	}

	ensureStartEndLineBreaks(&text)

	err = WriteFile(fileInfo.TranslateLocation.TargetFile, text)
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
