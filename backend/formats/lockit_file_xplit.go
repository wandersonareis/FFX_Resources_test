package formats

import (
	"bytes"
	"ffxresources/backend/common"
	"ffxresources/backend/interactions"
	"fmt"
	"os"
)

type LockitFileXplit struct {
	path string
}

// Função para criar uma instância da struct
func newLockitFileXplit(dataInfo *interactions.GameDataInfo) *LockitFileXplit {
	return &LockitFileXplit{path: dataInfo.GameData.AbsolutePath}
}

// Função para contar ocorrências de 0d0a
func (fh *LockitFileXplit) CountOccurrences(data []byte) int {
	return bytes.Count(data, []byte{0x0d, 0x0a})
}

func (fh *LockitFileXplit) ensureCrescentOrder(sizes []int) error {
	for i := 1; i < len(sizes); i++ {
		if sizes[i] <= sizes[i-1] {
			return fmt.Errorf("os tamanhos devem estar em ordem crescente")
		}
	}
	return nil
}

// Função para separar as partes e salvar em arquivos
func (fh *LockitFileXplit) XplitFile(sizes []int, outputFileNameBase, outputDir string) error {
	// Lê o arquivo como bytes
	data, err := os.ReadFile(fh.path)
	if err != nil {
		return fmt.Errorf("erro ao ler o arquivo: %v", err)
	}

	// Verifica se o array de tamanhos está em ordem crescente
	err = fh.ensureCrescentOrder(sizes)
	if err != nil {
		return err
	}

	occurrences := 0
	offset := 0
	for i, targetSize := range sizes {
		// Define a quantidade de quebras de linha que cada parte deve ter
		targetLines := targetSize - occurrences

		var part bytes.Buffer
		for j := 0; j < targetLines; j++ {
			index := bytes.Index(data[offset:], []byte{0x0d, 0x0a})
			if index == -1 {
				break // Se não houver mais quebras, para o loop
			}
			part.Write(data[offset : offset+index+2])
			offset += index + 2
			occurrences++
		}

		outputFileName := fmt.Sprintf(outputFileNameBase + ".loc_kit_%02d", i)
		err := os.WriteFile(common.PathJoin(outputDir, outputFileName), part.Bytes(), 0644)
		if err != nil {
			return fmt.Errorf("erro ao criar arquivo de saída: %v", err)
		}
	}

	if offset < len(data) {
		remainingPart := data[offset:]
		outputFileName := fmt.Sprintf(outputFileNameBase + ".loc_kit_%02d", len(sizes))
		err := os.WriteFile(common.PathJoin(outputDir, outputFileName), remainingPart, 0644)
		if err != nil {
			return fmt.Errorf("erro ao criar arquivo de saída: %v", err)
		}
	}

	return nil
}

/* func (fh *LockitFileXplit) ValidateParts(sizes []int) (bool, error) {
	originalData, err := os.ReadFile(fh.path)
	if err != nil {
		return false, fmt.Errorf("erro ao ler o arquivo original: %v", err)
	}

	var combinedBuffer bytes.Buffer

	for i := 0; i < len(sizes)+1; i++ {
		fileName := fmt.Sprintf("locit%03d.bin", i)
		partData, err := os.ReadFile(fileName)
		if err != nil {
			return false, fmt.Errorf("erro ao ler a parte %s: %v", fileName, err)
		}
		combinedBuffer.Write(partData)
	}

	err = os.WriteFile("combinedParts.bin", combinedBuffer.Bytes(), 0644)
	if err != nil {
		return false, fmt.Errorf("erro ao criar arquivo de saída: %v", err)
	}

	return bytes.Equal(originalData, combinedBuffer.Bytes()), nil
}
 */