package internal

import (
	"ffxresources/backend/common"
	"ffxresources/backend/interactions"
)

func FindDcpParts(parts *[]DcpFileParts, targetPath, pattern string) error {
	fileParts := make([]string, 0, 7)

	common.EnsurePathExists(targetPath)

	if err := common.EnumerateFilesByPattern(&fileParts, targetPath, pattern); err != nil {
		return err
	}

	for _, file := range fileParts {
		info := interactions.NewGameDataInfo(file)
		if info.GameData.Size == 0 {
			continue
		}

		dcpPart := NewDcpFileParts(info)
		if dcpPart == nil {
			continue
		}

		*parts = append(*parts, *dcpPart)
	}

	return nil
}