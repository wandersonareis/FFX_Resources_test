package dcpCore

import (
	"bytes"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/dcp/internal/dcpParts"
	"ffxresources/backend/interfaces"
	"fmt"
	"os"
)

type (
	IDcpFileJoiner interface {
		DcpFileJoiner(source interfaces.ISource, destination locations.IDestination, xplitedFiles components.IList[dcpParts.DcpFileParts], targetReimportFile string) error
	}

	dcpFileJoiner struct {
	}
)

func NewDcpFileJoiner() IDcpFileJoiner {
	return &dcpFileJoiner{}
}

func (dfj *dcpFileJoiner) DcpFileJoiner(source interfaces.ISource, destination locations.IDestination, xplitedFiles components.IList[dcpParts.DcpFileParts], targetReimportFile string) error {
	originalDcpFile := source.Get().Path

	importLocation := destination.Import().Get()

	if err := importLocation.ProvideTargetPath(); err != nil {
		return fmt.Errorf("error when providing target path: %w", err)
	}

	if err := dcpWriter(originalDcpFile, xplitedFiles, targetReimportFile); err != nil {
		return err
	}

	return nil
}

func dcpWriter(inputFilePath string, parts components.IList[dcpParts.DcpFileParts], newContainerPath string) error {
	inputFile, err := os.Open(inputFilePath)
	if err != nil {
		return fmt.Errorf("error when opening the original file: %w", err)
	}

	defer inputFile.Close()

	header := newHeader()
	if err := header.FromFile(inputFilePath); err != nil {
		return fmt.Errorf("error when reading the header: %w", err)
	}

	if err := header.Update(parts); err != nil {
		return err
	}

	var buffer bytes.Buffer

	if err := header.Write(&buffer); err != nil {
		return fmt.Errorf("error when writing the header: %w", err)
	}

	content := NewContentWithBuffer(&buffer)
	if err := content.Write(header, parts); err != nil {
		return fmt.Errorf("error when writing the content: %w", err)
	}

	newFile, err := os.Create(newContainerPath)
	if err != nil {
		return fmt.Errorf("error when creating the new container file: %w", err)
	}

	defer newFile.Close()

	if _, err := buffer.WriteTo(newFile); err != nil {
		return fmt.Errorf("error when writing buffer to file: %w", err)
	}

	/* originalData, err := os.ReadFile(inputFilePath)
	if err != nil {
		return fmt.Errorf("error reading the original file: %v", err)
	}

	newData, _ := os.ReadFile(newContainerPath)

	isExactMatch := bytes.Equal(originalData, newData)
	if !isExactMatch {
		return fmt.Errorf("arquivos não correspondem")
	} else {
		fmt.Println("Arquivos dcp correspondem")
	} */

	return nil
}
