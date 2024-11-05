package dcp_internal

import (
	"bytes"
	"ffxresources/backend/common"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type DataRanges struct {
	Start int64
	End   int64
}

type Content struct {
	container *bytes.Buffer
	ranges    []DataRanges
	outputDir string
}

func NewContent(outputDir string) *Content {
	return &Content{
		outputDir: outputDir,
	}
}

func NewContentWithBuffer(container *bytes.Buffer) *Content {
	return &Content{
		container: container,
	}
}

func (c *Content) CalculateRanges(header *Header, file *os.File) error {
	//func calculateFileDataRange(start, end *int64, pointers []Pointer, file *os.File, index int) error {
	for i := 0; i < len(header.Pointers); i++ {
		ranges := DataRanges{}
		ranges.Start = int64(header.Pointers[i].Value)

		if i+1 < len(header.Pointers) {
			ranges.End = int64(header.Pointers[i+1].Value)
		} else {
			fileInfo, err := file.Stat()
			if err != nil {
				return fmt.Errorf("erro ao obter informações do arquivo: %w", err)
			}
			ranges.End = fileInfo.Size()
		}

		c.ranges = append(c.ranges, ranges)
	}

	return nil
	//}
}

func (c Content) Read(file *os.File) error {
	//func readDataBlock(file *os.File, start, end *int64) ([]byte, error) {
	//length := *end - *start
	// 	for _, range := range c.ranges {

	// 	if _, err := file.Seek(*start, io.SeekStart); err != nil {
	// 		return nil, fmt.Errorf("error when positioning in the file: %w", err)
	// 	}

	// 	data := make([]byte, length)

	// 	if _, err := io.ReadFull(file, data); err != nil {
	// 		return nil, fmt.Errorf("Error reading data: %w", err)
	// 	}

	// 	//return data, nil
	// }

	for i, dataRange := range c.ranges {
		dataLentgh := dataRange.End - dataRange.Start

		if _, err := file.Seek(dataRange.Start, io.SeekStart); err != nil {
			return fmt.Errorf("error when positioning in the file: %w", err)
		}

		data := make([]byte, dataLentgh)

		outputFileName := fmt.Sprintf("%s.%03d", "macrodic", i)

		outputFilePartsPath := common.PathJoin(c.outputDir, outputFileName)

		if _, err := io.ReadFull(file, data); err != nil {
			return fmt.Errorf("error reading data: %w", err)
		}

		if err := common.WriteBytesToFile(outputFilePartsPath, data); err != nil {
			return fmt.Errorf("erro ao salvar o arquivo %03d: %w", i, err)
		}

	}

	return nil
}
func (c Content) Write(header *Header, parts *[]DcpFileParts) error {
	for _, part := range *parts {
		filePath := part.gameDataInfo.GameData.AbsolutePath
		fileName := filepath.Base(filePath)

		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("error when opening the file %s: %w", fileName, err)
		}

		defer file.Close()

		if _, err := io.Copy(c.container, file); err != nil {
			return fmt.Errorf("error recording the data from the file %s to the container: %w", fileName, err)
		}
	}

	return nil
}
