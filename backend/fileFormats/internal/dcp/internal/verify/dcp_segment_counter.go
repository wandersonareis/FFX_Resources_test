package verify

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core"
	"ffxresources/backend/core/components"
	"ffxresources/backend/fileFormats/internal/dcp/internal/parts"
	"ffxresources/backend/logger"
	"fmt"
	"os"
)

type ISegmentCounter interface {
	CountBinaryParts(dcpFileParts components.IList[parts.DcpFileParts], options core.IDcpFileOptions) error
	CountTextParts(partsList components.IList[parts.DcpFileParts]) error
}

type segmentCounter struct {
	log logger.ILoggerHandler
}

func NewSegmentCounter(logger logger.ILoggerHandler) ISegmentCounter {
	return &segmentCounter{
		log: logger,
	}
}

func (sc *segmentCounter) CountBinaryParts(dcpFileParts components.IList[parts.DcpFileParts], options core.IDcpFileOptions) error {
	if dcpFileParts.GetLength() != options.GetPartsLength() {
		sc.log.LogError(nil, "error when ensuring splited macrodic parts", "expected parts", options.GetPartsLength(), "current parts", dcpFileParts.GetLength())

		return fmt.Errorf("error when ensuring splited macrodic parts")
	}

	for _, dcpFilePart := range dcpFileParts.GetItems() {
		sourceFile := dcpFilePart.Source().Get()
		if sourceFile.Size == 0 {
			if err := os.Remove(sourceFile.Path); err != nil {
				sc.log.LogError(err, "error when removing part", "file", sourceFile.Path)

				return fmt.Errorf("error when removing part")
			}

			return fmt.Errorf("invalid size for part: %s", dcpFilePart.Source().Get().Name)
		}
	}

	return nil
}

func (sc *segmentCounter) CountTextParts(partsList components.IList[parts.DcpFileParts]) error {
	errChan := make(chan error, partsList.GetLength())
	defer close(errChan)

	partsList.ForEach(func(part parts.DcpFileParts) {
		targetFile := part.Destination().Extract().Get().GetTargetFile()
		if common.CountSegments(targetFile) <= 0 {
			sc.log.LogError(nil, "error when counting segments in part: %s", targetFile)

			errChan <- fmt.Errorf("error when counting segments in part: %s", targetFile)
		}
	})

	if err := <-errChan; err != nil {
		return err
	}

	return nil
}
