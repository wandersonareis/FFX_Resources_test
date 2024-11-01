package common

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func PathJoin(parts ...string) string {
	return filepath.Join(parts...)
}

func sanitizationPath(path string) string {
	return filepath.Clean(path)
}

func GetAbsolutePath(path string) (string, error) {
	return filepath.Abs(path)
}

func GetDir(path string) string {
	cPath := sanitizationPath(path)
	return filepath.Dir(cPath)
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

func EnumerateFilesDev(s string) ([]string, error) {
	fullpath, err := GetAbsolutePath(s)
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

func EnumerateFilesByPattern(files *[]string, path, pattern string) error {
	fullpath, err := GetAbsolutePath(path)
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

	//path := fileInfo.AbsolutePath
	index := strings.Index(path, marker)
	if index == -1 {
		log.Println("unable to find marker in path:", path)
		return ""
	}

	return path[index:]
}
