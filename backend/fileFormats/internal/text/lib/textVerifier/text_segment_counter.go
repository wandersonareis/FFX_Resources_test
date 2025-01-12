package textVerifier

import (
	"ffxresources/backend/common"
	"fmt"
	"os"
)

type ISegmentCounter interface {
	CountBinary(targetFile string) error
	CountText(targetFile string) error
}

type segmentCounter struct{}

func (sc *segmentCounter) CountBinary(targetFile string) error {
	info, err := os.Stat(targetFile)
	if err != nil {
		return fmt.Errorf("error when getting file info: %w", err)
	}

	if info.Size() == 0 {
		/* if err := os.Remove(targetFile); err != nil {
			return fmt.Errorf("error when removing file: %w", err)
		} */

		return fmt.Errorf("invalid size for part: %s size: %d", targetFile, info.Size())
	}

	return nil
}

func (sc *segmentCounter) CountText(targetFile string) error {
	if common.CountSegments(targetFile) == 0 {
		return fmt.Errorf("error when counting segments in: %s", targetFile)
	}

	return nil
}
