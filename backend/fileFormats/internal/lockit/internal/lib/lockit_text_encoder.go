package lib

import (
	"ffxresources/backend/fileFormats/util"
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
)

func LockitEncoderFfx(lockitFileInfo interactions.IGameDataInfo) error {
	characterTable := util.NewCharacterTable()

	codeTable, err := characterTable.GetCharacterOnlyTable()
	if err != nil {
		return err
	}

	defer characterTable.Dispose(codeTable)

	return encoderBase(lockitFileInfo, codeTable)
}

func LockitEncoderLoc(lockitFileInfo interactions.IGameDataInfo) error {
	characterTable := util.NewCharacterTable()

	codeTable, err := characterTable.GetCharacterLocTable()
	if err != nil {
		return err
	}

	defer characterTable.Dispose(codeTable)

	if err := encoderBase(lockitFileInfo, codeTable); err != nil {
		return err
	}

	encodedFile := lockitFileInfo.GetImportLocation().TargetFile

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

	if err := lib.RunCommand(executable, args); err != nil {
		return err
	}

	return nil
}
