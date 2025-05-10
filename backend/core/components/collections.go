package components

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/interfaces"
	"os"
	"path/filepath"
	"regexp"
)

func ListFilesByRegex(list IList[string], path, pattern string) error {
	fullpath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	err = filepath.WalkDir(fullpath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && regex.MatchString(d.Name()) {
			absPath, err := filepath.Abs(path)
			if err != nil {
				return err
			}

			list.Add(absPath)
		}

		return nil
	})

	if err != nil {
		return err
	}

	list.Clip()

	return nil
}

func PopulateGameFilePartsList[T any](
	parts IList[T],
	targetPath,
	pattern string,
	partsInstance func(source interfaces.ISource, destination locations.IDestination) *T) error {
	if err := common.EnsurePathExists(targetPath); err != nil {
		return err
	}

	filesList := NewList[string](parts.GetLength())

	err := ListFilesByRegex(filesList, targetPath, pattern)
	if err != nil {
		return err
	}

	filesList.ForEach(func(item string) {
		s, err := locations.NewSource(item)
		if err != nil {
			return
		}

		if s.GetSize() == 0 {
			return
		}

		t := locations.NewDestination()

		part := partsInstance(s, t)
		if part == nil {
			return
		}

		parts.Add(*part)
	})

	parts.Clip()

	return nil
}

func PopulateFilePartsList[T any](
	parts IList[T],
	targetPath,
	pattern string,
	formatter interfaces.ITextFormatter,
	partsInstance func(source interfaces.ISource, destination locations.IDestination, formatter interfaces.ITextFormatter) *T) error {
	if err := common.EnsurePathExists(targetPath); err != nil {
		return err
	}

	filesList := NewList[string](parts.GetLength())

	err := ListFilesByRegex(filesList, targetPath, pattern)
	if err != nil {
		return err
	}

	filesList.ForEach(func(item string) {
		s, err := locations.NewSource(item)
		if err != nil {
			return
		}

		if s.GetSize() == 0 {
			return
		}

		t := locations.NewDestination()

		part := partsInstance(s, t, formatter)
		if part == nil {
			return
		}

		parts.Add(*part)
	})

	parts.Clip()

	return nil
}
