package formats

import (
	"encoding/binary"
	"ffxresources/backend/lib"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func dcpFileJoiner(fileInfo *lib.FileInfo, xplitedFiles *[]string, targetReimportFile string) error {
	/* xpliterHandler, err := GetDcpXplitHandler()
	if err != nil {
		return err
	} */

	originalDcpFile := fileInfo.AbsolutePath

	//reimportFile := fileInfo.ImportLocation.TargetFile
	reimportFilePath := fileInfo.ImportLocation.TargetPath
	lib.EnsurePathExists(reimportFilePath)

	err := DcpWriter(originalDcpFile, xplitedFiles, targetReimportFile)
	if err != nil {
		return err
	}

	/* args, err := dcpJoinerArgs()
	if err != nil {
		return err
	}

	args = append(args, originalDcpFile, reimportedDcpPartsDirectory, reimportFile)

	err = lib.RunCommand(xpliterHandler, args)
	if err != nil {
		return err
	} */

	return nil
}

func DcpWriter(originalFilePath string, xplitedFiles *[]string, newContainerPath string) error {
	// Abrir o arquivo container original
	originalFile, err := os.Open(originalFilePath)
	if err != nil {
		return fmt.Errorf("erro ao abrir o arquivo original: %w", err)
	}

	defer originalFile.Close()

	const headerSize = 0x40
	header := make([]byte, headerSize)

	_, err = io.ReadFull(originalFile, header)
	if err != nil {
		return fmt.Errorf("erro ao ler o header: %w", err)
	}

	var pointers = make([]lib.Pointer, 0, 7)

	err = ExtractPointers(header, &pointers)
	if err != nil {
		return fmt.Errorf("erro ao extrair os ponteiros: %w", err)
	}

	newContainer, err := os.Create(newContainerPath)
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("erro ao criar o novo arquivo container: %w", err)
	}

	defer newContainer.Close()

	err = recalculatePointers(pointers, header, *xplitedFiles)
	if err != nil {
		return fmt.Errorf("erro ao atualizar os ponteiros: %w", err)
	}

	_, err = newContainer.Write(header)
	if err != nil {
		return fmt.Errorf("erro ao gravar o header: %w", err)
	}

	err = writeFilesToContainer(pointers, *xplitedFiles, newContainer)
	if err != nil {
		return fmt.Errorf("erro ao gravar os arquivos: %w", err)
	}

	fmt.Println("Novo arquivo container criado com sucesso.")
	return nil
}

func recalculatePointers(pointers []lib.Pointer, header []byte, xplitedFiles []string) error {
	var currentOffset uint32 = uint32(pointers[0].Value) // O primeiro ponteiro permanece o mesmo

	for i, pointer := range pointers {
		filePath := xplitedFiles[i]
		fileName := filepath.Base(filePath)
		fileInfo, err := os.Stat(filePath)

		if err != nil {
			return fmt.Errorf("erro ao obter informações do arquivo %s: %w", fileName, err)
		}

		// Para o primeiro ponteiro, não alteramos
		if i == 0 {
			currentOffset = uint32(pointer.Value) + uint32(fileInfo.Size())
			continue
		}

		// Para os ponteiros subsequentes, recalculamos com base no tamanho do arquivo anterior
		newPointer := currentOffset
		binary.LittleEndian.PutUint32(header[pointer.Offset:], newPointer)

		// Atualizar o offset para o próximo arquivo
		currentOffset = newPointer + uint32(fileInfo.Size())
	}

	return nil
}

func writeFilesToContainer(pointers []lib.Pointer, xplitedFiles []string, newContainer *os.File) error {
	for i := 0; i < len(pointers); i++ {
		filePath := xplitedFiles[i]
		fileName := filepath.Base(filePath)

		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("erro ao abrir o arquivo %s: %w", fileName, err)
		}

		defer file.Close()

		_, err = io.Copy(newContainer, file)
		if err != nil {
			return fmt.Errorf("erro ao gravar os dados do arquivo %s no container: %w", fileName, err)
		}
	}
	return nil
}