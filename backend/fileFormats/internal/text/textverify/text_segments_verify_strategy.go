package textverify

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/text/internal/lib"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/models"
	"fmt"
	"time"
)

type textSegmentsVerificationStrategy struct {
	TargetFile string
}

func NewTextSegmentsVerificationStrategy(targetFile string) ITextVerificationStrategy {
	return &textSegmentsVerificationStrategy{
		TargetFile: targetFile,
	}
}

func (tsv *textSegmentsVerificationStrategy) Verify(source interfaces.ISource, destination locations.IDestination) error {
	sourceFile := source.GetPath()
	sourceFileType := source.GetType()
	sourceFileVersion := source.GetVersion()

	if err := common.CheckPathExists(sourceFile); err != nil {
		return fmt.Errorf("failed to check source file path: %s", err)
	}

	if err := common.CheckPathExists(tsv.TargetFile); err != nil {
		return fmt.Errorf("failed to check target file path: %s", err)
	}

	if err := tsv.compareTextSegmentsCount(sourceFile, tsv.TargetFile, sourceFileType, sourceFileVersion); err != nil {
		if err := common.RemoveFileWithRetries(tsv.TargetFile, 5, time.Second); err != nil {
			return fmt.Errorf("failed to remove broken text file: %s", tsv.TargetFile)
		}
		return err
	}

	return nil
}

func (tsv *textSegmentsVerificationStrategy) compareTextSegmentsCount(binaryFile, textFile string, binaryType models.NodeType, gameVersion models.GameVersion) error {
	binarySegmentCount, err := lib.TextSegmentsCounter(binaryFile, binaryType, gameVersion)
	if err != nil {
		return err
	}

	textSegmentCount, err := tsv.countTextSegments(textFile)
	if err != nil {
		return err
	}

	if textSegmentCount != binarySegmentCount {
		return fmt.Errorf("segment count mismatch: binary file has %d segments, text file has %d segments", binarySegmentCount, textSegmentCount)
	}

	return nil
}

func (tsv *textSegmentsVerificationStrategy) countTextSegments(targetFile string) (int, error) {
	segments := common.CountSegments(targetFile)
	if segments == 0 {
		return 0, fmt.Errorf("error when counting segments: this file is empty %s", targetFile)
	}

	return segments, nil
}
