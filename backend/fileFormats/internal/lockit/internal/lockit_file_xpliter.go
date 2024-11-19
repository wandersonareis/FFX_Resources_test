package internal

import (
	"bufio"
	"bytes"
	"ffxresources/backend/common"
	"ffxresources/backend/interactions"
	"fmt"
	"os"
	"path/filepath"
)

type lockitFileXplitter struct {
	dataInfo interactions.IGameDataInfo
	options  *interactions.LockitFileOptions
}

func NewLockitFileXplitter(dataInfo interactions.IGameDataInfo) *lockitFileXplitter {
	return &lockitFileXplitter{
		dataInfo: dataInfo,
		options:  interactions.NewInteraction().GamePartOptions.GetLockitFileOptions(),
	}
}

func (lx *lockitFileXplitter) SegmentFile(parts *[]LockitFileParts) {
	worker := common.NewWorker[LockitFileParts]()

	worker.ParallelForEach(*parts,
		func(index int, part LockitFileParts) {
			if index > 0 && index%2 == 0 {
				part.Extract(LocEnc)
			} else {
				part.Extract(FfxEnc)
			}
		})
}

func (lx *lockitFileXplitter) EnsurePartsExists() error {
	if err := lx.xplitter(); err != nil {
		return err
	}

	return nil
}

func (lx *lockitFileXplitter) countLineEndings(data []byte) int {
	return bytes.Count(data, []byte{0x0d, 0x0a})
}

func (lx *lockitFileXplitter) checkSizesAscending(sizes []int) error {
	for i := 1; i < len(sizes); i++ {
		if sizes[i] <= sizes[i-1] {
			return fmt.Errorf("sizes must be in ascending order")
		}
	}
	return nil
}

func (lx *lockitFileXplitter) xplitter() error {
	extractLocation := lx.dataInfo.GetExtractLocation()

	if err := extractLocation.ProvideTargetPath(); err != nil {
		return err
	}

	file, err := os.Open(lx.dataInfo.GetGameData().FullFilePath)
	if err != nil {
		return fmt.Errorf("error when opening the file: %v", err)
	}

	defer file.Close()

	segmentPartsSizes := lx.options.PartsSizes

	if err := lx.checkSizesAscending(segmentPartsSizes); err != nil {
		return err
	}

	reader := bufio.NewReader(file)

	lx.divideFileByLineEndingCount(reader, extractLocation.TargetPath)

	return nil
}

func (lx *lockitFileXplitter) divideFileByLineEndingCount(reader *bufio.Reader, outputDir string) error {
	occurrences := 0
	partIndex := 0

	segmentPartsSizes := lx.options.PartsSizes


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
		occurrences += lx.countLineEndings(line)

		if partIndex < len(segmentPartsSizes) && occurrences >= segmentPartsSizes[partIndex] {
			outputFileName := filepath.Join(outputDir, fmt.Sprintf("%s.part%02d", lx.options.NameBase, partIndex))

			if err := os.WriteFile(outputFileName, buffer, 0644); err != nil {
				return fmt.Errorf("error when writing the file: %v", err)
			}

			buffer = nil
			partIndex++
		}
	}

	if len(buffer) > 0 {
		outputFileName := filepath.Join(outputDir, fmt.Sprintf("%s.part%02d", lx.options.NameBase, partIndex))

		if err := os.WriteFile(outputFileName, buffer, 0644); err != nil {
			return fmt.Errorf("error when writing the file: %v", err)
		}
	}

	return nil
}