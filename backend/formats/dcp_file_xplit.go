package formats

import (
	"encoding/binary"
	"ffxresources/backend/common"
	"ffxresources/backend/interactions"
	"ffxresources/backend/models"
	"fmt"
	"io"
	"os"
)

func dcpFileXpliter(dataInfo *interactions.GameDataInfo) error {
	targetFile := dataInfo.GameData.AbsolutePath
	targetNamePrefix := dataInfo.GameData.NamePrefix
	outputPath := dataInfo.ExtractLocation.TargetPath
	
	common.EnsurePathExists(outputPath)

	err := DcpReader(targetFile, targetNamePrefix, outputPath)
	if err != nil {
		return err
	}

	return nil
}

func DcpReader(dcpFilePath, namePrefix, outputDir string) error {
	file, err := os.Open(dcpFilePath)
	if err != nil {
		return fmt.Errorf("erro ao abrir o arquivo %s: %w", dcpFilePath, err)
	}
	defer file.Close()

	const headerSize = 0x40
	header := make([]byte, headerSize)

	_, err = io.ReadFull(file, header)
	if err != nil {
		return fmt.Errorf("erro ao ler o header: %w", err)
	}

	var pointers = make([]models.Pointer, 0, 7)

	err = ExtractPointers(header, &pointers)
	if err != nil {
		return fmt.Errorf("erro ao extrair os ponteiros: %w", err)
	}

	for i := 0; i < len(pointers); i++ {
		var start, end int64

		err := calculateFileDataRange(&start, &end, pointers, file, i)
		if err != nil {
			return fmt.Errorf("erro ao calcular o intervalo de dados: %w", err)
		}

		data, err := readDataBlock(file, &start, &end)
		if err != nil {
			return fmt.Errorf("erro ao ler dados do arquivo: %w", err)
		}

		outputFileName := fmt.Sprintf("%s.%03d", namePrefix, i)
		outputFilePartsPath := common.PathJoin(outputDir, outputFileName)

		err = common.WriteBytesToFile(outputFilePartsPath, data)
		if err != nil {
			return fmt.Errorf("erro ao salvar o arquivo %03d: %w", i, err)
		}

		fmt.Printf("Arquivo salvo: arquivo.%03d\n", i)
	}

	return nil
}

func ExtractPointers(header []byte, pointers *[]models.Pointer) error {
	// Iterar sobre o header em blocos de 4 bytes
	for i := 0; i < len(header); i += 4 {
		// Ler o valor como uint32 (formato Little Endian)
		value := binary.LittleEndian.Uint32(header[i : i+4])

		// Adicionar à lista apenas se o valor for diferente de zero
		if value != 0 {
			*pointers = append(*pointers, models.Pointer{
				Offset: int64(i),
				Value:  value,
			})
		}
	}

	return nil
}

func calculateFileDataRange(start, end *int64, pointers []models.Pointer, file *os.File, index int) error {
	*start = int64(pointers[index].Value)

	if index+1 < len(pointers) {
		*end = int64(pointers[index+1].Value)
	} else {
		fileInfo, err := file.Stat()
		if err != nil {
			return fmt.Errorf("erro ao obter informações do arquivo: %w", err)
		}
		*end = fileInfo.Size()
	}

	return nil
}

func readDataBlock(file *os.File, start, end *int64) ([]byte, error) {
	length := *end - *start

	_, err := file.Seek(*start, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("erro ao posicionar no arquivo: %w", err)
	}

	data := make([]byte, length)
	_, err = io.ReadFull(file, data)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler dados: %w", err)
	}

	return data, nil
}
