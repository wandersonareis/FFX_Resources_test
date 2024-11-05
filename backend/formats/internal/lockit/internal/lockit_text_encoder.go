package lockit_internal

import (
	"ffxresources/backend/common"
	tbstables "ffxresources/backend/formats/internal/tbs"
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
)

func lockitEncoderFfx(lockitFileInfo *interactions.GameDataInfo) error {
	codeTable, err := tbstables.NewCharacterTable().GetCharacterOnlyTable()
	if err != nil {
		return err
	}

	return encoderBase(lockitFileInfo, codeTable)
}

func lockitEncoderLoc(lockitFileInfo *interactions.GameDataInfo) error {
	codeTable, err := tbstables.NewCharacterTable().GetCharacterLocTable()
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
	
	importLocation := lockitFileInfo.ImportLocation

	if err := importLocation.ProvideTargetPath(); err != nil {
		return err
	}

	args := make([]string, 0, 4)
	args = append(args, "-tr", codeTable, targetFile, importLocation.TargetFile)

	if err := lib.RunCommand(handler, args); err != nil {
		return err
	}

	return nil
}

func ensureUtf8Bom(target string) error {
	handler, err := getLockitFileUtf8BomNormalizer()
	if err != nil {
		return err
	}

	defer common.RemoveFile(handler)

	args := make([]string, 0, 2)
	args = append(args, "-r", target)

	if err := lib.RunCommand(handler, args); err != nil {
		return err
	}

	return nil
}
