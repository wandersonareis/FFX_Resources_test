package textVerifier

import (
	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/internal/text/internal/lib"
	"ffxresources/backend/models"
	"fmt"
)

type (
	ISegmentCounter interface {
		CompareTextSegmentsCount(binaryFile, textFile string, binaryType models.NodeType, gameVersion models.GameVersion) error
	}

	segmentCounter struct{}
)

func newSegmentCounter() ISegmentCounter {
	return &segmentCounter{}
}

func (sc *segmentCounter) CompareTextSegmentsCount(binaryFile, textFile string, binaryType models.NodeType, gameVersion models.GameVersion) error {
	binarySegmentCount, err := lib.TextSegmentsCounter(binaryFile, binaryType, gameVersion)
	if err != nil {
		return err
	}

	textSegmentCount, err := sc.countTextSegments(textFile)
	if err != nil {
		return err
	}

	if textSegmentCount != binarySegmentCount {
		return fmt.Errorf("source and target segments count mismatch: %s: %d, %s: %d", binaryFile, textSegmentCount, textFile, binarySegmentCount)
	}

	return nil
}

func (sc *segmentCounter) countTextSegments(targetFile string) (int, error) {
	segments := common.CountSegments(targetFile)
	if segments == 0 {
		return 0, fmt.Errorf("error when counting segments: this file is empty %s", targetFile)
	}

	return segments, nil
}
