package internal

import (
	"bufio"
	"io"

	"ffxresources/backend/core"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/interfaces"
	"fmt"
	"os"
	"path/filepath"
)

type ILockitFileSplitter interface {
	FileSplitter(source interfaces.ISource, extractLocation locations.IExtractLocation, options core.ILockitFileOptions) error
}

type lockitFileSplitter struct{}

func NewLockitFileSplitter() ILockitFileSplitter {
	return &lockitFileSplitter{}
}

func (ls *lockitFileSplitter) FileSplitter(source interfaces.ISource, extractLocation locations.IExtractLocation, options core.ILockitFileOptions) error {
	if err := extractLocation.ProvideTargetPath(); err != nil {
		return fmt.Errorf("error when providing the target path: %w", err)
	}

	file, err := os.Open(source.GetPath())
	if err != nil {
		return fmt.Errorf("error when opening the file: %v", err)
	}

	defer file.Close()

	reader := bufio.NewReader(file)

	if err := ls.splitFileByLineCount(
		reader,
		extractLocation.GetTargetPath(),
		options.GetNameBase(),
		options.GetPartsSizes(),
		options.GetPartsLength(),
		options.GetLineBreaksCount()); err != nil {
		return fmt.Errorf("error when dividing the file: %w", err)
	}

	return nil
}

func (lfs *lockitFileSplitter) splitFileByLineCount(reader *bufio.Reader, outputDir string, segmentPartsName string, segmentPartsSizes []int, segmentPartsLength, segmentMaxLineBreaks int) error {
	partIndex := 0
	lineCount := 0

	var buffer []byte

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error reading the file: %v", err)
		}

		buffer = append(buffer, line...)

		// maxLineCount calculates the cumulative sum of segment sizes from index 0 to partIndex (inclusive).
		maxLineCount := sum(segmentPartsSizes[:partIndex+1])

		lineCount++

		if partIndex < segmentPartsLength && lineCount == maxLineCount {
			outputFileName := filepath.Join(outputDir, fmt.Sprintf("%s.part%02d", segmentPartsName, partIndex))

			if err := os.WriteFile(outputFileName, buffer, 0644); err != nil {
				return fmt.Errorf("error when writing the file: %v", err)
			}

			buffer = nil
			partIndex++
		}
	}

	if lineCount != segmentMaxLineBreaks {
		return fmt.Errorf("error: expected %d line breaks, but found %d", segmentMaxLineBreaks, lineCount)
	}

	return nil
}

func sum(slice []int) int {
	total := 0
	for _, v := range slice {
		total += v
	}
	return total
}
