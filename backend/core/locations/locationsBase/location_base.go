package locationsBase

import (
	"ffxresources/backend/common"
	"ffxresources/backend/loggingService"
	"os"
	"path/filepath"
)

type (
	ILocationBase interface {
		ITargetDirectoryBase
		ITargetFileBase
		IBuildTargets

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
			l := loggingService.Get()
			l.Error().Msgf("error when removing file: %s", err)
		}
	}
}
