package locationsBase

import (
	"ffxresources/backend/common"
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

func NewLocationBase(targetDirectoryName, targetDirectoryPath, gameVersionDir string) LocationBase {
	targetDirectory := filepath.Join(targetDirectoryPath, gameVersionDir)

	return LocationBase{
		TargetDirectoryBase: TargetDirectoryBase{
			targetDirectory:     targetDirectory,
			targetDirectoryName: targetDirectoryName,
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
