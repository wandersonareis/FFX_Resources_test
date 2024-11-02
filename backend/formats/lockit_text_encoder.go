package formats

import (
	"ffxresources/backend/common"
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
)

func lockitEncoderFfx(lockitFileInfo *interactions.GameDataInfo) error {
	codeTable, err := getCharacterOnlyTable()
	if err != nil {
		return err
	}

	return encoderBase(lockitFileInfo, codeTable)
}

func lockitEncoderLoc(lockitFileInfo *interactions.GameDataInfo) error {
	codeTable, err := getCharacterLocTable()
	if err != nil {
		return err
	}

	err = encoderBase(lockitFileInfo, codeTable)
	if err != nil {
		return err
	}

	encodedFile := lockitFileInfo.ImportLocation.TargetFile

	return ensureUtf8Bom(encodedFile)
}

func encoderBase(lockitFileInfo *interactions.GameDataInfo, codeTable string) error {
	handler, err := getLockitFileHandler()
	if err != nil {
		return err
	}

	defer common.RemoveFile(handler)

	defer common.RemoveFile(codeTable)

	targetFile := lockitFileInfo.TranslateLocation.TargetFile
	outputFile := lockitFileInfo.ImportLocation.TargetFile
	outputPath := lockitFileInfo.ImportLocation.TargetPath

	err = common.EnsurePathExists(outputPath)
	if err != nil {
		return err
	}

	args := make([]string, 0, 4)
	args = append(args, "-tr", codeTable, targetFile, outputFile)

	err = lib.RunCommand(handler, args)
	if err != nil {
		return err
	}

	return nil
}

func ensureUtf8Bom(target string) error {
	handler, err := getLockitFileUtf8BomHandler()
	if err != nil {
		return err
	}

	defer common.RemoveFile(handler)

	args := make([]string, 0, 2)
		args = append(args, "-r", target)

		err = lib.RunCommand(handler, args)
		if err != nil {
			return err
		}

	return nil
}