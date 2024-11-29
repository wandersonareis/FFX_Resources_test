package verify

import (
	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/internal/dcp/internal/parts"
	"ffxresources/backend/interactions"
	"fmt"
	"os"
)

type ISegmentCounter interface {
	CountBinaryParts(dcpFileParts *[]parts.DcpFileParts, options interactions.DcpFileOptions) error
	CountTextParts(partsList *[]parts.DcpFileParts, options interactions.DcpFileOptions) error
}

type segmentCounter struct{}

func (sc *segmentCounter) CountBinaryParts(dcpFileParts *[]parts.DcpFileParts, options interactions.DcpFileOptions) error {
	if len(*dcpFileParts) != options.PartsLength {
		return fmt.Errorf("error when ensuring splited macrodic parts: expected parts: %d Got parts: %d", options.PartsLength, len(*dcpFileParts))
	}

	for _, dcpFilePart := range *dcpFileParts {
		if dcpFilePart.GetGameData().Size == 0 {
			os.Remove(dcpFilePart.GetGameData().FullFilePath)
			return fmt.Errorf("invalid size for part: %s", dcpFilePart.GetGameData().Name)
		}
	}

	return nil
}

func (sc *segmentCounter) CountTextParts(partsList *[]parts.DcpFileParts, options interactions.DcpFileOptions) error {
	list := *partsList

	for index, part := range list {
		if common.CountSegments(part.GetExtractLocation().TargetFile) <= 0 {
			return fmt.Errorf("error when counting segments in part %d", index)
		}
	}

	return nil
}
