package lockitParts

import (
	"ffxresources/backend/core/components"
	"ffxresources/backend/fileFormats/internal/lockit/internal/lib"
	"fmt"
)

func populateLockitPartsList(list components.IList[LockitFileParts], path, pattern string) error {
	err := components.PopulateGameFilePartsList(
		list,
		path,
		pattern,
		NewLockitFileParts)

	if err != nil {
		return fmt.Errorf(
			"error when finding lockit binary parts: %s in path: %s",
			err.Error(),
			path)
	}

	return nil
}

func PopulateLockitBinaryFileParts(
	binaryPartsList components.IList[LockitFileParts],
	path string) error {
	return populateLockitPartsList(
		binaryPartsList,
		path,
		lib.LOCKIT_FILE_BINARY_PARTS_PATTERN,
	)
}

func PopulateLockitTextFileParts(
	translatedPartsList components.IList[LockitFileParts],
	path string) error {
	return populateLockitPartsList(
		translatedPartsList,
		path,
		lib.LOCKIT_FILE_TXT_PARTS_PATTERN,
	)
}
