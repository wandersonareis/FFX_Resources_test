package splitter

import (
	"bufio"
	"bytes"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/encoding"
	"ffxresources/backend/fileFormats/internal/lockit/internal/parts"
	"ffxresources/backend/interactions"
	"fmt"
	"os"
	"path/filepath"
)

type IFileSplitter interface {
	DecoderPartsFiles(partsList components.IList[parts.LockitFileParts])
	FileSplitter(dataInfo interactions.IGameDataInfo, options interactions.LockitFileOptions) error
}

type LockitFileSplitter struct {}

func NewLockitFileSplitter() IFileSplitter {
	return &LockitFileSplitter{}
}

func (ls *LockitFileSplitter) DecoderPartsFiles(partsList components.IList[parts.LockitFileParts]) {
	encoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextLocalizationEncoding()
	defer encoding.Dispose()

	extractorFunc := func(index int, part parts.LockitFileParts) {
		if index > 0 && index % 2 == 0 {
			part.Extract(parts.LocEnc, encoding)
		} else {
			part.Extract(parts.FfxEnc, encoding)
		}
	}

	partsList.ParallelForEach(extractorFunc)
}

func (ls *LockitFileSplitter) FileSplitter(dataInfo interactions.IGameDataInfo, options interactions.LockitFileOptions) error {
	extractLocation := dataInfo.GetExtractLocation()

	if err := extractLocation.ProvideTargetPath(); err != nil {
		return fmt.Errorf("error when providing the target path: %w", err)
	}

	file, err := os.Open(dataInfo.GetGameData().FullFilePath)
	if err != nil {
		return fmt.Errorf("error when opening the file: %v", err)
	}

	defer file.Close()

	segmentPartsSizes := options.PartsSizes

	if err := checkSizesAscending(segmentPartsSizes); err != nil {
		return fmt.Errorf("error when checking the sizes: %w", err)
	}

	reader := bufio.NewReader(file)

	if err := splitFileByLineCount(reader, extractLocation.TargetPath, options); err != nil {
		return fmt.Errorf("error when dividing the file: %w", err)
	}

	return nil
}

func countLineEndings(data []byte) int {
	return bytes.Count(data, []byte{0x0d, 0x0a})
}

func checkSizesAscending(sizes []int) error {
	for i := 1; i < len(sizes); i++ {
		if sizes[i] <= sizes[i-1] {
			return fmt.Errorf("sizes must be in ascending order")
		}
	}
	return nil
}

func splitFileByLineCount(reader *bufio.Reader, outputDir string, options interactions.LockitFileOptions) error {
	occurrences := 0
	partIndex := 0

	segmentPartsSizes := options.PartsSizes

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
		occurrences += countLineEndings(line)

		if partIndex < len(segmentPartsSizes) && occurrences >= segmentPartsSizes[partIndex] {
			outputFileName := filepath.Join(outputDir, fmt.Sprintf("%s.part%02d", options.NameBase, partIndex))

			if err := os.WriteFile(outputFileName, buffer, 0644); err != nil {
				return fmt.Errorf("error when writing the file: %v", err)
			}

			buffer = nil
			partIndex++
		}
	}

	if len(buffer) > 0 {
		outputFileName := filepath.Join(outputDir, fmt.Sprintf("%s.part%02d", options.NameBase, partIndex))

		if err := os.WriteFile(outputFileName, buffer, 0644); err != nil {
			return fmt.Errorf("error when writing the file: %v", err)
		}
	}

	return nil
}
