package internal

import (
	"ffxresources/backend/common"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"os"
	"path/filepath"
)

type (
	ILocationBase interface {
		ITargetDirectoryBase
		ITargetFileBase

		BuildTargetReadOutput(source interfaces.ISource, formatter interfaces.ITextFormatter)
		BuildTargetWriteOutput(source interfaces.ISource, formatter interfaces.ITextFormatter)
		Dispose()
	}
	LocationBase struct {
		TargetDirectoryBase
		TargetFileBase
	}
)

func NewLocationBase(options *LocationBaseOptions) LocationBase {
	gameVersionDirBase := interactions.NewInteractionService().FFXGameVersion().GetGameVersion().String()

	targetDirectory := filepath.Join(options.TargetDirectory, gameVersionDirBase)

	return LocationBase{
		TargetDirectoryBase: TargetDirectoryBase{
			TargetDirectory:     targetDirectory,
			TargetDirectoryName: options.TargetDirectoryName,
		},
		TargetFileBase: TargetFileBase{},
	}
}

func (lb *LocationBase) Dispose() {
	if common.IsFileExists(lb.GetTargetFile()) {
		err := os.Remove(lb.TargetFile)
		if err != nil {
			l := logger.Get()
			l.Error().Msgf("error when removing file: %s", err)
		}
	}
}
