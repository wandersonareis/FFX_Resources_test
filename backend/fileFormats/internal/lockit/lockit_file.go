package lockit

import (
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/base"
	"ffxresources/backend/fileFormats/internal/lockit/internal/lib"
	"ffxresources/backend/fileFormats/internal/lockit/internal/lockitFileParts"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
)

type LockitFile struct {
	*base.FormatsBase
	*lockitFileCompressor
	*LockitFileExtractor
}

func NewLockitFile(source interfaces.ISource, destination locations.IDestination) interfaces.IFileProcessor {
	partsList := components.NewEmptyList[lockitFileParts.LockitFileParts]()

	destination.CreateRelativePath(source, interactions.NewInteractionService().GameLocation.GetTargetDirectory())

	destination.InitializeLocations(source, formatters.NewTxtFormatter())

	err := components.GenerateGameFileParts(
		partsList,
		destination.Extract().Get().GetTargetPath(),
		lib.LOCKIT_FILE_PARTS_PATTERN,
		lockitFileParts.NewLockitFileParts)

	if err != nil {
		l := logger.Get()
		l.Error().Err(err).Msg("error when finding lockit parts")
		return nil
	}

	return &LockitFile{
		FormatsBase:          base.NewFormatsBase(source, destination),
		lockitFileCompressor: newLockitFileCompressor(source, destination, partsList),
		LockitFileExtractor:  newLockitFileExtractor(source, destination, partsList),
	}
}
