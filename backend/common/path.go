package common

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
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
	dir := cPath

	if filepath.Ext(cPath) != "" {
		dir = filepath.Dir(cPath)
	}

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	return nil
}

func ListFilesInDirectory(s string) ([]string, error) {
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

	return results, nil
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

func GetRelativePath(from, to string) string {
	from = filepath.Clean(from)
	to = filepath.Clean(to)

	index := strings.Index(from, to)
	if index == -1 {
		log.Println("unable to find marker in path:", to)
		return ""
	}

	return to[index:]
}

func GetDifferencePath(fullPath, basePath string) string {
	if strings.HasPrefix(fullPath, basePath) {
		return strings.TrimPrefix(fullPath, basePath + "\\")
	}
	return fullPath
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
