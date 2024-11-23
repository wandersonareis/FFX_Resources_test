package common

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

func sanitizationPath(path string) string {
	return filepath.Clean(path)
}

func GetDir(path string) string {
	cPath := sanitizationPath(path)
	return filepath.Dir(cPath)
}

func IsPathExists(path string) bool {
	cPath := sanitizationPath(path)
	_, err := os.Stat(cPath)
	return !os.IsNotExist(err)
}

func EnsurePathExists(path string) error {
	cPath := sanitizationPath(path)

	if filepath.Ext(cPath) != "" {
		cPath = filepath.Dir(cPath)
	}

	if err := os.MkdirAll(cPath, os.ModePerm); err != nil {
		return fmt.Errorf("error when creating the destination directory: %w", err)
	}

	return nil
}

func ListFilesInDirectory(s string) (*[]string, error) {
	fullpath, err := filepath.Abs(s)
	if err != nil {
		return nil, err
	}

	var results []string

	err = filepath.Walk(fullpath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			results = append(results, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	results = slices.Clip(results)

	return &results, nil
}

func ListFilesMatchingPattern(files *[]string, path, pattern string) error {
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
			*files = append(*files, absPath)
		}
		return nil
	})

	*files = slices.Clip(*files)

	return err
}

func GetRelativePathFromMarker(path string) string {
	var marker = FFX_DIR_MARKER

	index := strings.Index(path, marker)
	if index == -1 {
		log.Println("unable to find marker in path:", path)
		return ""
	}

	return path[index:]
}

func MakeRelativePath(from, to string) string {
	if strings.HasPrefix(from, to) {
		return strings.TrimPrefix(from, to + "\\")
	}
	return from
}

func ContainsNewUSPCPath(path string) error {
	cPath := filepath.Clean(path)

	requiredSequence := filepath.Join("ffx_ps2", "ffx2", "master", "new_uspc")
	requiredPath := filepath.Join(cPath, requiredSequence)

	if !IsPathExists(requiredPath) {
		return fmt.Errorf("is not a valid spira us path: %s", path)
	}

	return nil
}

func ContainsGameResourcesPath(path string) error {
	cPath := filepath.Clean(path)

	requiredSequence := filepath.Join("ffx-2_data", "gamedata", "ps3data")
	requiredPath := filepath.Join(cPath, requiredSequence)

	if !IsPathExists(requiredPath) {
		return fmt.Errorf("is not a valid spira game resources us path: %s", path)
	}

	return nil
}
