package formats

import (
	"ffxresources/backend/common"
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
)

func lockitDecoderFfx(lockitFileInfo *interactions.GameDataInfo) error {
	codeTable, err := getCharacterOnlyTable()
	if err != nil {
		return err
	}

	return decoderBase(lockitFileInfo, codeTable)
}

func lockitDecoderLoc(lockitFileInfo *interactions.GameDataInfo) error {
	codeTable, err := getCharacterLocTable()
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
	outputFile := lockitFileInfo.ExtractLocation.TargetFile
	outputPath := lockitFileInfo.ExtractLocation.TargetPath

	err = common.EnsurePathExists(outputPath)
	if err != nil {
		return err
	}

	args := make([]string, 0, 4)
	args = append(args, "-t", codeTable)
	args = append(args, targetFile)
	args = append(args, outputFile)

	err = lib.RunCommand(handler, args)
	if err != nil {
		return err
	}

	return nil
}
