package lockit_internal

import (
	"bufio"
	"bytes"
	"ffxresources/backend/common"
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
	"fmt"
	"os"
	"path/filepath"
)

type LockitFileXplit struct {
	path string
}

// Função para criar uma instância da struct
func NewLockitFileXplit(dataInfo *interactions.GameDataInfo) *LockitFileXplit {
	return &LockitFileXplit{path: dataInfo.GameData.AbsolutePath}
}

// Função para contar ocorrências de 0d0a
func (fh *LockitFileXplit) CountOccurrences(data []byte) int {
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

// Função para separar as partes e salvar em arquivos
func (fh *LockitFileXplit) XplitFile(sizes []int, outputFileNameBase, outputDir string) error {
	file, err := os.Open(fh.path)
	if err != nil {
		return fmt.Errorf("error when opening the file: %v", err)
	}

	defer file.Close()

	if err := fh.ensureCrescentOrder(sizes); err != nil {
		return err
	}

	lib.LogSeverity(lib.SeverityInfo, "Creating parts of the file ...")

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
		occurrences += fh.CountOccurrences(line)

		if partIndex < len(sizes) && occurrences >= sizes[partIndex] {
			outputFileName := filepath.Join(outputDir, fmt.Sprintf("%s.part%02d", outputFileNameBase, partIndex))

			if err := os.WriteFile(outputFileName, buffer, 0644); err != nil {
				return fmt.Errorf("error when writing the file: %v", err)
			}

			lib.LogSeverity(lib.SeverityInfo, fmt.Sprintf("Part %d created", partIndex))
			buffer = nil
			partIndex++
		}
	}

	if len(buffer) > 0 {
		outputFileName := common.PathJoin(outputDir, fmt.Sprintf("%s.part%02d", outputFileNameBase, partIndex))
		
		if err := os.WriteFile(outputFileName, buffer, 0644); err != nil {
			return fmt.Errorf("error when writing the file: %v", err)
		}

		lib.LogSeverity(lib.SeverityInfo, fmt.Sprintf("Part %d created", partIndex))
	}

	return nil
}
