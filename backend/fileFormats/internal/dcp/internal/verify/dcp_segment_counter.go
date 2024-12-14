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

	"github.com/rs/zerolog"
)

type ISegmentCounter interface {
	CountBinaryParts(dcpFileParts components.IList[parts.DcpFileParts], options interactions.DcpFileOptions) error
	CountTextParts(partsList components.IList[parts.DcpFileParts], options interactions.DcpFileOptions) error
}

type segmentCounter struct {
	log zerolog.Logger
}

func NewSegmentCounter() ISegmentCounter {
	return &segmentCounter{
		log: logger.Get().With().Str("module", "dcp_segment_counter").Logger(),
	}
}

func (sc *segmentCounter) CountBinaryParts(dcpFileParts components.IList[parts.DcpFileParts], options interactions.DcpFileOptions) error {
	if dcpFileParts.GetLength() != options.PartsLength {
		sc.log.Error().
			Int("expected parts", options.PartsLength).
			Int("current parts", dcpFileParts.GetLength()).
			Msg("error when ensuring splited macrodic parts")

		return fmt.Errorf("error when ensuring splited macrodic parts")
	}

	for _, dcpFilePart := range dcpFileParts.GetItems() {
		sourceFile := dcpFilePart.Source().Get()
		if sourceFile.Size == 0 {
			if err := os.Remove(sourceFile.Path); err != nil {
				sc.log.Error().
					Err(err).
					Str("file", sourceFile.Path).
					Msg("error when removing part")

				return fmt.Errorf("error when removing part")
			}

			return fmt.Errorf("invalid size for part: %s", dcpFilePart.Source().Get().Name)
		}
	}

	return nil
}

func (sc *segmentCounter) CountTextParts(partsList components.IList[parts.DcpFileParts], options interactions.DcpFileOptions) error {
	errChan := make(chan error, partsList.GetLength())

	go notifications.ProcessError(errChan, sc.log)

	partsList.ForEach(func(part parts.DcpFileParts) {
		targetFile := part.Destination().Extract().Get().GetTargetFile()
		if common.CountSegments(targetFile) <= 0 {
			sc.log.Error().
				Str("part", targetFile).
				Msg("error when counting segments in part")

			errChan <- fmt.Errorf("error when counting segments in part: %s", targetFile)
		}
	})

	defer close(errChan)

	return nil
}
