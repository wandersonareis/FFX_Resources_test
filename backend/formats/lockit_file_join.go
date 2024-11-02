package formats

import (
	"bytes"
	"ffxresources/backend/common"
	"ffxresources/backend/interactions"
	"fmt"
	"os"
)

type LockitFileJoin struct {
	dataInfo *interactions.GameDataInfo
}

func newLockitFileJoin(dataInfo *interactions.GameDataInfo) *LockitFileJoin {
	return &LockitFileJoin{
		dataInfo: dataInfo,
	}
}

func (l *LockitFileJoin) JoinFile(sizes *[]int) error {
	parts := make([]LockitFileParts, 0, 17)
	partsPath := common.PathJoin(l.dataInfo.ImportLocation.TargetDirectory, common.LOCKIT_TARGET_DIR_NAME)

	err := findLockitParts(&parts, partsPath, common.LOCKIT_FILE_PARTS_PATTERN)
	if err != nil {
		return err
	}

	if len(parts) != len(*sizes)+1 {
		return fmt.Errorf("invalid number of parts: %d expected: %d", len(parts), len(*sizes)+1)

	}

	originalData, err := os.ReadFile(l.dataInfo.GameData.AbsolutePath)
	if err != nil {
		return fmt.Errorf("erro ao ler o arquivo original: %v", err)
	}

	var combinedBuffer bytes.Buffer

	for i := 0; i < len(*sizes)+1; i++ {
		/* if parts[i].dataInfo.GameData.NamePrefix != txtParts[i].dataInfo.GameData.NamePrefix {
			return fmt.Errorf("prefixos de arquivos diferentes")
		} */
		
		fileName := parts[i].dataInfo.ImportLocation.TargetFile
		partData, err := os.ReadFile(fileName)
		if err != nil {
			return fmt.Errorf("erro ao ler a parte %s: %v", fileName, err)
		}
		combinedBuffer.Write(partData)
	}

	targetFile := l.dataInfo.ImportLocation.TargetFile
	err = os.WriteFile(targetFile, combinedBuffer.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo de saída: %v", err)
	}

	isExactMatch := bytes.Equal(originalData, combinedBuffer.Bytes())
	if !isExactMatch {
		return fmt.Errorf("arquivos não correspondem")
	}

	return nil
}
