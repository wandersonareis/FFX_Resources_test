package common

import (
	"bufio"
	"ffxresources/backend/models"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func CountSegments(targetFile string) int {
	separator := FFX_TEXT_FORMAT_SEPARATOR

	input, err := ReadFileAsString(targetFile)
	if err != nil {
		return 0
	}

	return strings.Count(input, separator)
}

func EnsureWindowsLineBreaks(targetFile string, nodeType models.NodeType) error {
	if nodeType == models.Dcp {
		return nil
	}

	file, err := os.Open(targetFile)
	if err != nil {
		return fmt.Errorf("error when opening the file: %s", err)
	}
	defer file.Close()

	text, err := convertLineBreaksToWindowsFormat(file)
	if err != nil {
		return fmt.Errorf("error when reading the file: %s", err)
	}

	ensureLineBreaksAtStartAndEnd(&text)

	err = WriteStringToFile(targetFile, text)
	if err != nil {
		return fmt.Errorf("error saving the modified file: %s", err)
	}

	return nil
}

func convertLineBreaksToWindowsFormat(file *os.File) (string, error) {
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

func ensureLineBreaksAtStartAndEnd(text *string) {
	startRegex := regexp.MustCompile(`^\r\n`)
	endRegex := regexp.MustCompile(`\r\n$`)

	if !startRegex.MatchString(*text) {
		*text = "\r\n" + *text
	}

	if !endRegex.MatchString(*text) {
		*text = *text + "\r\n"
	}
}
