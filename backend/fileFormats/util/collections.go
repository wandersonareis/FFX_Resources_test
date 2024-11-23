package util

import (
	"ffxresources/backend/common"
	"ffxresources/backend/interactions"
	"slices"
)

func FindFileParts[T any](parts *[]T, targetPath, pattern string, partsInstance func(info interactions.IGameDataInfo) *T) error {
	fileParts := make([]string, 0, len(*parts))

	common.EnsurePathExists(targetPath)

	if err := common.ListFilesMatchingPattern(&fileParts, targetPath, pattern); err != nil {
		return err
	}

	worker := common.NewWorker[string]()

	worker.ForEach(&fileParts, func(_ int, item string) error {
		info := interactions.NewGameDataInfo(item)
		if info.GetGameData().Size == 0 {
			return nil
		}

		part := partsInstance(info)
		if part == nil {
			return nil
		}

		*parts = append(*parts, *part)
		return nil
	})

	*parts = slices.Clip(*parts)

	return nil
}
