package internal

import (
	"ffxresources/backend/interactions"
	"fmt"
	"os"
)

func DcpFileXpliter(dataInfo *interactions.GameDataInfo) error {
	targetFile := dataInfo.GameData.AbsolutePath

	extractLocation := dataInfo.ExtractLocation

	if err := extractLocation.ProvideTargetPath(); err != nil {
		return err
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
		return fmt.Errorf("erro ao calcular os intervalos de dados: %w", err)
	}
	
	content := NewContent(header, outputDir)


	if err := content.Read(file); err != nil {
		return fmt.Errorf("erro ao ler os dados: %w", err)
	}

	return nil
}
