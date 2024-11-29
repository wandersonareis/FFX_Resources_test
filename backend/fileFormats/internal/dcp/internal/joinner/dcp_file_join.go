package joinner

import (
	"bytes"
	"ffxresources/backend/fileFormats/internal/dcp/internal/file"
	"ffxresources/backend/fileFormats/internal/dcp/internal/parts"
	"ffxresources/backend/interactions"
	"fmt"
	"os"
)

func DcpFileJoiner(dataInfo interactions.IGameDataInfo, xplitedFiles *[]parts.DcpFileParts, targetReimportFile string) error {
	originalDcpFile := dataInfo.GetGameData().FullFilePath

	importLocation := dataInfo.GetImportLocation()

	if err := importLocation.ProvideTargetPath(); err != nil {
		return fmt.Errorf("error when providing target path: %w", err)
	}

	err := dcpWriter(originalDcpFile, xplitedFiles, targetReimportFile)
	if err != nil {
		return err
	}

	return nil
}

func dcpWriter(inputFilePath string, parts *[]parts.DcpFileParts, newContainerPath string) error {
	inputFile, err := os.Open(inputFilePath)
	if err != nil {
		return fmt.Errorf("error when opening the original file: %w", err)
	}

	defer inputFile.Close()

	header := file.NewHeader()
	header.FromFile(inputFilePath)
	header.Update(*parts)

	var buffer bytes.Buffer

	if err := header.Write(&buffer); err != nil {
		return fmt.Errorf("error when writing the header: %w", err)
	}

	content := file.NewContentWithBuffer(&buffer)
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
