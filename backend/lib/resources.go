package lib

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/denisvmedia/asar"
)

func getResourcesAsarFile() string {
	return filepath.Join(GetExecDir(), "resources.asar")
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
	err := WriteBytesToFile(targetFile, bytes)
	if err != nil {
		return err
	}

	if !FileExists(targetFile) {
		return fmt.Errorf("error creating handler file: %s", filepath.Base(targetFile))
	}

	return nil
}

/* func GetDcpXplitHandler() (string, error) {
	targetHandler := []string{
		interaction.WorkingLocation.WorkDirectoryName,
		DCP_FILE_XPLITTER_APPLICATION,
	}

	targetFile := filepath.Join(interaction.WorkingLocation.WorkDirectory, DCP_FILE_XPLITTER_APPLICATION)
	err := GetFileFromResources(targetHandler, targetFile)
	if err != nil {
		return "", err
	}

	return targetFile, nil
} */

func GetFileFromResources(targetLocation []string, targetFile string) error {
	EnsurePathExists(targetFile)

	resourcesFile := getResourcesAsarFile()
	if !FileExists(resourcesFile) {
		return fmt.Errorf("resources.asar not found")
	}

	bytes, err := readFileFromAsar(resourcesFile, targetLocation)
	if err != nil {
		return err
	}

	err = createHandlerFile(targetFile, bytes)
	if err != nil {
		return err
	}

	return nil
}
