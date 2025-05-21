package locationsBase

import (
	"ffxresources/backend/common"
	"ffxresources/backend/loggingService"
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
		if err := common.RemoveFileWithRetries(lb.TargetFile, 3); err != nil {
			l := loggingService.Get()
			l.Error().Msgf("error when removing file: %s with error: %v", lb.TargetFile, err)
		}
	}
}
