package util

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/interactions"
)

func FindFileParts[T any](partsList components.IList[T], targetPath, pattern string, partsInstance func(info interactions.IGameDataInfo) *T) error {
	fileParts := components.NewList[string](partsList.GetLength())

	common.EnsurePathExists(targetPath)

	if err := components.ListFilesByRegex(fileParts, targetPath, pattern); err != nil {
		return err
	}

	generatePartInstanceFunc := func(item string) {
		info := interactions.NewGameDataInfo(item)
		if info.GetGameData().Size == 0 {
			return
		}

		part := partsInstance(info)
		if part == nil {
			return
		}

		partsList.Add(*part)
	}

	fileParts.ForEach(generatePartInstanceFunc)

	return nil
}
