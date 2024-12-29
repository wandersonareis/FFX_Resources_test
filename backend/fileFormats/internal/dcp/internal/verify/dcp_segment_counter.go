package verify

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/fileFormats/internal/dcp/internal/parts"
	"ffxresources/backend/interactions"
	"ffxresources/backend/logger"
	"ffxresources/backend/notifications"
	"fmt"
	"os"
)

type ISegmentCounter interface {
	CountBinaryParts(dcpFileParts components.IList[parts.DcpFileParts], options interactions.DcpFileOptions) error
	CountTextParts(partsList components.IList[parts.DcpFileParts], options interactions.DcpFileOptions) error
}

type segmentCounter struct {
	logger.ILoggerHandler
}

func NewSegmentCounter() ISegmentCounter {
	return &segmentCounter{
		ILoggerHandler: &logger.LogHandler{
			Logger: logger.Get().With().Str("module", "dcp_segment_counter").Logger(),
		},
	}
}

func (sc *segmentCounter) CountBinaryParts(dcpFileParts components.IList[parts.DcpFileParts], options interactions.DcpFileOptions) error {
	if dcpFileParts.GetLength() != options.PartsLength {
		sc.LogError(nil, "error when ensuring splited macrodic parts", "expected parts", options.PartsLength, "current parts", dcpFileParts.GetLength())

		return fmt.Errorf("error when ensuring splited macrodic parts")
	}

	for _, dcpFilePart := range dcpFileParts.GetItems() {
		sourceFile := dcpFilePart.Source().Get()
		if sourceFile.Size == 0 {
			if err := os.Remove(sourceFile.Path); err != nil {
				sc.LogError(err, "error when removing part", "file", sourceFile.Path)

				return fmt.Errorf("error when removing part")
			}

			return fmt.Errorf("invalid size for part: %s", dcpFilePart.Source().Get().Name)
		}
	}

	return nil
}

func (sc *segmentCounter) CountTextParts(partsList components.IList[parts.DcpFileParts], options interactions.DcpFileOptions) error {
	errChan := make(chan error, partsList.GetLength())

	go notifications.ProcessError(errChan, sc.GetLogger())

	partsList.ForEach(func(part parts.DcpFileParts) {
		targetFile := part.Destination().Extract().Get().GetTargetFile()
		if common.CountSegments(targetFile) <= 0 {
			sc.LogError(nil, "error when counting segments in part: %s", targetFile)

			errChan <- fmt.Errorf("error when counting segments in part: %s", targetFile)
		}
	})

	defer close(errChan)

	return nil
}
