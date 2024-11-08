package internal

import (
	"ffxresources/backend/common"
	"ffxresources/backend/interactions"
)

func FindLockitParts(parts *[]LockitFileParts, targetPath, pattern string) error {
	fileParts := make([]string, 0, 16)

	common.EnsurePathExists(targetPath)

	if err := common.ListFilesMatchingPattern(&fileParts, targetPath, pattern); err != nil {
		return err
	}

	for _, file := range fileParts {
		info := interactions.NewGameDataInfo(file)
		if info.GameData.Size == 0 {
			continue
		}

		lockitPart := NewLockitFileParts(info)
		if lockitPart == nil {
			continue
		}

		*parts = append(*parts, *lockitPart)
	}

	return nil
}