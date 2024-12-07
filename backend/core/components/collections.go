package components

import (
	"ffxresources/backend/common"
	"ffxresources/backend/interactions"
	"os"
	"path/filepath"
	"regexp"
)

func ListFiles(list IList[string], s string) error {
	fullpath, err := filepath.Abs(s)
	if err != nil {
		return err
	}

	err = filepath.Walk(fullpath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			list.Add(path)
		}

		return nil
	})

	if err != nil {
		return err
	}

	list.Clip()

	return nil
}

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

func GenerateGameFileParts[T any](parts IList[T], targetPath, pattern string, partsInstance func(info interactions.IGameDataInfo) *T) error {
	common.EnsurePathExists(targetPath)

	filesList := NewList[string](parts.GetLength())

	err := ListFilesByRegex(filesList, targetPath, pattern)
	if err != nil {
		return err
	}

	filesList.ForEach(func(item string) {
		info := interactions.NewGameDataInfo(item)
		if info.GetGameData().Size == 0 {
			return
		}

		part := partsInstance(info)
		if part == nil {
			return
		}

		parts.Add(*part)
	})

	parts.Clip()

	return nil
}