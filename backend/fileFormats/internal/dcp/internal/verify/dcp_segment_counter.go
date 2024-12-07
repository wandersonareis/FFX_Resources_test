package verify

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/fileFormats/internal/dcp/internal/parts"
	"ffxresources/backend/interactions"
	"ffxresources/backend/logger"
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
		if dcpFilePart.GetGameData().Size == 0 {
			if err := os.Remove(dcpFilePart.GetGameData().FullFilePath); err != nil {
				sc.log.Error().
					Err(err).
					Str("file", dcpFilePart.GetGameData().FullFilePath).
					Msg("error when removing part")

				return fmt.Errorf("error when removing part")
			}

			return fmt.Errorf("invalid size for part: %s", dcpFilePart.GetGameData().Name)
		}
	}

	return nil
}

func (sc *segmentCounter) CountTextParts(partsList components.IList[parts.DcpFileParts], options interactions.DcpFileOptions) error {
	for _, part := range partsList.GetItems() {
		if common.CountSegments(part.GetExtractLocation().TargetFile) <= 0 {
			sc.log.Error().
				Str("part", part.GetExtractLocation().TargetFile).
				Msg("error when counting segments in part")

			return fmt.Errorf("error when counting segments in part")
		}
	}

	return nil
}
