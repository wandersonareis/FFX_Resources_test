package internal

import (
	"ffxresources/backend/fileFormats/util"
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
)

func lockitEncoderFfx(lockitFileInfo interactions.IGameDataInfo) error {
	characterTable := util.NewCharacterTable()
	characterTable.Dispose()

	codeTable, err := characterTable.GetCharacterOnlyTable()
	if err != nil {
		return err
	}

	return encoderBase(lockitFileInfo, codeTable)
}

func lockitEncoderLoc(lockitFileInfo *interactions.GameDataInfo) error {
	characterTable := util.NewCharacterTable()
	characterTable.Dispose()

	codeTable, err := characterTable.GetCharacterLocTable()
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

func encoderBase(lockitFileInfo interactions.IGameDataInfo, codeTable string) error {
	handler := newLockitHandler()
	defer handler.dispose()

	executable, err := handler.getLockitFileHandler()
	if err != nil {
		return err
	}

	targetFile := lockitFileInfo.GetTranslateLocation().TargetFile
	
	importLocation := lockitFileInfo.GetImportLocation()

	if err := importLocation.ProvideTargetPath(); err != nil {
		return err
	}
	
	args := []string{"-tr", codeTable, targetFile, importLocation.TargetFile}

	/* args := make([]string, 0, 4)
	args = append(args, "-tr", codeTable, targetFile, importLocation.TargetFile) */

	if err := lib.RunCommand(executable, args); err != nil {
		return err
	}

	return nil
}

func ensureUtf8Bom(target string) error {
	handler := newLockitHandler()
	defer handler.dispose()

	executable, err := handler.getLockitFileUtf8BomNormalizer()
	if err != nil {
		return err
	}

	args := []string{"-r", target}

	/* args := make([]string, 0, 2)
	args = append(args, "-r", target) */

	if err := lib.RunCommand(executable, args); err != nil {
		return err
	}

	return nil
}
