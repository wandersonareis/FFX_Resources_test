package lib

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
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

	info, err := os.Stat(cPath)

	var dir string

	if os.IsNotExist(err) {
		if filepath.Ext(cPath) != "" {
			dir = filepath.Dir(cPath)
		} else {
			dir = cPath
		}
	}

	if info != nil {
		if info.IsDir() {
			dir = cPath
		} else {
			dir = filepath.Dir(cPath)
		}
	}

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	return nil
}

func EnumerateFiles(s string, wg *sync.WaitGroup, results chan<- string, errors chan<- error) {
	fullpath, err := GetAbsolutePath(s)
	if err != nil {
		errors <- err
		return
	}

	defer wg.Done()

	err = filepath.Walk(fullpath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			errors <- err
			return nil
		}

		if !info.IsDir() {
			results <- sanitizationPath(path)
		}

		return nil
	})

	if err != nil {
		errors <- err
	}
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

func EnumerateFilesByPattern(s, pattern string) ([]string, error) {
	fullpath, err := GetAbsolutePath(s)
	if err != nil {
		return nil, err
	}

	var files []string

	var regex = regexp.MustCompile(pattern)

	err = filepath.Walk(fullpath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		fullpath, err := GetAbsolutePath(path)
		if err != nil {
			return err
		}

		if !info.IsDir() && regex.MatchString(info.Name()) {
			files = append(files, fullpath)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

func GuessTypeByPath(path string) NodeType {
	sPath := sanitizationPath(path)
	info, err := os.Stat(sPath)
	if err != nil {
		return None
	}

	/* if !hasSpira(path) {
		return backend.None
	} */

	if info.IsDir() {
		return Folder
	}

	return guessBySpiraFileType(path)
}

func GetRelativePathFromMarker(fileInfo FileInfo) (string, error) {
	var marker = FFX_DIR_MARKER

	path := fileInfo.AbsolutePath
	index := strings.Index(path, marker)
	if index == -1 {
		return "", fmt.Errorf("the path does not contain the marker '%s' -> '%s'", marker, path)
	}

	return path[index:], nil
}

func GenerateExtractedOutput(fileInfo FileInfo, workDirectory, targetDirName, targetExtension string) (string, string) {
	outputFile := PathJoin(workDirectory, targetDirName, ChangeExtension(fileInfo.RelativePath, targetExtension))
	outputPath := filepath.Dir(outputFile)
	return outputFile, outputPath
}

func GeneratedTranslatedOutput(fileInfo FileInfo, workDirectory string) (string, string) {
	outputFile := PathJoin(workDirectory, ChangeExtension(fileInfo.RelativePath, fileInfo.Extension))
	outputPath := filepath.Dir(outputFile)
	return outputFile, outputPath
}
