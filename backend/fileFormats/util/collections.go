package util

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/interfaces"
	"fmt"
)

func FindFileParts[T any](partsList components.IList[T], targetPath, pattern string, formatter interfaces.ITextFormatter,
	partsInstance func(source interfaces.ISource, destination locations.IDestination) *T) error {
	fileParts := components.NewList[string](partsList.GetLength())

	if err := common.EnsurePathExists(targetPath); err != nil {
		return err
	}

	if err := components.ListFilesByRegex(fileParts, targetPath, pattern); err != nil {
		return err
	}

	errChan := make(chan error, fileParts.GetLength())
	defer close(errChan)

	generatePartInstanceFunc := func(item string) {
		source, err := locations.NewSource(item)
		if err != nil {
			errChan <- fmt.Errorf("error creating source: %w", err)
			return
		}

		destination := locations.NewDestination()
		destination.InitializeLocations(source, formatter)

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

	if err := <-errChan; err != nil {
		return err
	}

	return nil
}
