package common

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/denisvmedia/asar"
)

func getResourcesAsarFile() string {
	return filepath.Join(GetBasePath(), "resources.asar")
}

func readFileFromAsar(asarPackage string, asarIndexString []string) ([]byte, error) {
	asarFile, err := os.Open(asarPackage)
	if err != nil {
		return nil, err
	}
	defer asarFile.Close()

	archive, err := asar.Decode(asarFile)
	if err != nil {
		return nil, err
	}

	fileEntry := archive.Find(asarIndexString...)
	if fileEntry == nil {
		return nil, fmt.Errorf("file not found")
	}

	return fileEntry.Bytes(), nil
}

func createHandlerFile(targetFile string, bytes []byte) error {
	if err := WriteBytesToFile(targetFile, bytes); err != nil {
		return err
	}

	if !IsFileExists(targetFile) {
		return fmt.Errorf("error creating handler file: %s", filepath.Base(targetFile))
	}

	return nil
}

func GetFileFromResources(targetLocation []string, targetFile string) error {
	if err := EnsurePathExists(targetFile); err != nil {
		return err
	}

	resourcesFile := getResourcesAsarFile()
	if err := CheckPathExists(resourcesFile); err != nil {
		return fmt.Errorf("resources.asar not found: %w", err)
	}

	bytes, err := readFileFromAsar(resourcesFile, targetLocation)
	if err != nil {
		return err
	}

	if err = createHandlerFile(targetFile, bytes); err != nil {
		return err
	}

	return nil
}
