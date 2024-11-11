package lib

import (
	"ffxresources/backend/common"
	"ffxresources/backend/interactions"
)

func FindFileParts[T any](parts *[]T, targetPath, pattern string, partsInstance func(info *interactions.GameDataInfo) *T) error {
	fileParts := make([]string, 0, len(*parts))

	common.EnsurePathExists(targetPath)

	if err := common.ListFilesMatchingPattern(&fileParts, targetPath, pattern); err != nil {
		return err
	}

	worker := common.NewWorker[string]()

	worker.ForEach(fileParts, func(_ int, item string) error {
		info := interactions.NewGameDataInfo(item)
		if info.GameData.Size == 0 {
			return nil
		}

		dcpPart := partsInstance(info)
		if dcpPart == nil {
			return nil
		}

		*parts = append(*parts, *dcpPart)
		return nil
	})

	return nil
}