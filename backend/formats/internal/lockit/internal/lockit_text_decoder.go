package internal

import (
	"ffxresources/backend/common"
	"ffxresources/backend/formats/internal/tbs"
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
)

func lockitDecoderFfx(lockitFileInfo *interactions.GameDataInfo) error {
	codeTable, err := tbstables.NewCharacterTable().GetCharacterOnlyTable()
	if err != nil {
		return err
	}

	return decoderBase(lockitFileInfo, codeTable)
}

func lockitDecoderLoc(lockitFileInfo *interactions.GameDataInfo) error {
	codeTable, err := tbstables.NewCharacterTable().GetCharacterLocTable()
	if err != nil {
		return err
	}

	return decoderBase(lockitFileInfo, codeTable)
}

func decoderBase(lockitFileInfo *interactions.GameDataInfo, codeTable string) error {
	handler, err := getLockitFileHandler()
	if err != nil {
		return err
	}

	defer common.RemoveFile(handler)

	defer common.RemoveFile(codeTable)

	targetFile := lockitFileInfo.GameData.AbsolutePath

	extractLocation := lockitFileInfo.ExtractLocation

	if err := extractLocation.ProvideTargetPath(); err != nil {
		return err
	}

	args := make([]string, 0, 4)
	args = append(args, "-t", codeTable)
	args = append(args, targetFile)
	args = append(args, extractLocation.TargetFile)

	if err := lib.RunCommand(handler, args); err != nil {
		return err
	}

	return nil
}
