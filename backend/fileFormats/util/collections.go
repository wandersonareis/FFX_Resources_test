package util

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"ffxresources/backend/notifications"
	"fmt"
)

func FindFileParts[T any](partsList components.IList[T], targetPath, pattern string, partsInstance func(source interfaces.ISource, destination locations.IDestination) *T) error {
	fileParts := components.NewList[string](partsList.GetLength())

	common.EnsurePathExists(targetPath)

	if err := components.ListFilesByRegex(fileParts, targetPath, pattern); err != nil {
		return err
	}

	errChan := make(chan error, fileParts.GetLength())

	go notifications.ProcessError(errChan, logger.Get().With().Str("module", "findFilePartss").Logger())

	generatePartInstanceFunc := func(item string) {
		source, err := locations.NewSource(item, interactions.NewInteraction().FFXGameVersion().GetGameVersion())
		if err != nil {
			errChan <- fmt.Errorf("error creating source: %w", err)
			return
		}

		destination := locations.NewDestination()

		if source.Get().Size == 0 {
			return
		}

		part := partsInstance(source, destination)
		if part == nil {
			return
		}

		partsList.Add(*part)
	}

	fileParts.ForEach(generatePartInstanceFunc)

	return nil
}
