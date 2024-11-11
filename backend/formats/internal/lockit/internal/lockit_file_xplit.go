package internal

import (
	"bufio"
	"bytes"
	"ffxresources/backend/events"
	"ffxresources/backend/interactions"
	"fmt"
	"os"
	"path/filepath"
)

type LockitFileXplit struct {
	path string
}

func NewLockitFileXplit(dataInfo *interactions.GameDataInfo) *LockitFileXplit {
	return &LockitFileXplit{path: dataInfo.GameData.AbsolutePath}
}

func (fh *LockitFileXplit) countOccurrences(data []byte) int {
	return bytes.Count(data, []byte{0x0d, 0x0a})
}

func (fh *LockitFileXplit) ensureCrescentOrder(sizes []int) error {
	for i := 1; i < len(sizes); i++ {
		if sizes[i] <= sizes[i-1] {
			return fmt.Errorf("sizes must be in ascending order")
		}
	}
	return nil
}

func (fh *LockitFileXplit) XplitFile(sizes []int, outputFileNameBase, outputDir string) error {
	file, err := os.Open(fh.path)
	if err != nil {
		return fmt.Errorf("error when opening the file: %v", err)
	}

	defer file.Close()

	if err := fh.ensureCrescentOrder(sizes); err != nil {
		return err
	}

	events.LogSeverity(events.SeverityInfo, "Creating parts of the file ...")

	reader := bufio.NewReader(file)
	occurrences := 0
	partIndex := 0

	var buffer []byte

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return fmt.Errorf("error reading the file: %v", err)
		}

		buffer = append(buffer, line...)
		occurrences += fh.countOccurrences(line)

		if partIndex < len(sizes) && occurrences >= sizes[partIndex] {
			outputFileName := filepath.Join(outputDir, fmt.Sprintf("%s.part%02d", outputFileNameBase, partIndex))

			if err := os.WriteFile(outputFileName, buffer, 0644); err != nil {
				return fmt.Errorf("error when writing the file: %v", err)
			}

			events.LogSeverity(events.SeverityInfo, fmt.Sprintf("Part %d created", partIndex))
			buffer = nil
			partIndex++
		}
	}

	if len(buffer) > 0 {
		outputFileName := filepath.Join(outputDir, fmt.Sprintf("%s.part%02d", outputFileNameBase, partIndex))
		
		if err := os.WriteFile(outputFileName, buffer, 0644); err != nil {
			return fmt.Errorf("error when writing the file: %v", err)
		}

		events.LogSeverity(events.SeverityInfo, fmt.Sprintf("Part %d created", partIndex))
	}

	return nil
}
