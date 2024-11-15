package internal

import (
	"ffxresources/backend/interactions"
	"fmt"
	"os"
)

func DcpFileXpliter(dataInfo interactions.IGameDataInfo) error {
	targetFile := dataInfo.GetGameData().FullFilePath

	extractLocation := dataInfo.GetExtractLocation()

	if err := extractLocation.ProvideTargetPath(); err != nil {
		return fmt.Errorf("error when creating the extraction directory: %w", err)
	}

	if err := dcpReader(targetFile, extractLocation.TargetPath); err != nil {
		return err
	}

	return nil
}

func dcpReader(dcpFilePath, outputDir string) error {
	file, err := os.Open(dcpFilePath)
	if err != nil {
		return fmt.Errorf("error when opening the file %s: %w", dcpFilePath, err)
	}

	defer file.Close()

	header := NewHeader()
	header.FromFile(dcpFilePath)

	if err := header.DataLengths(header, file); err != nil {
		return fmt.Errorf("error when calculating the data intervals: %w", err)
	}

	content := NewContent(header, outputDir)

	if err := content.Read(file); err != nil {
		return fmt.Errorf("error reading the data: %w", err)
	}

	return nil
}
