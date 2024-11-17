package internal

import (
	"ffxresources/backend/fileFormats/util"
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
)

func lockitDecoderFfx(lockitFileInfo interactions.IGameDataInfo) error {
	characterTable := util.NewCharacterTable()
	characterTable.Dispose()

	codeTable, err := characterTable.GetCharacterOnlyTable()
	if err != nil {
		return err
	}

	return decoderBase(lockitFileInfo, codeTable)
}

func lockitDecoderLoc(lockitFileInfo interactions.IGameDataInfo) error {
	characterTable := util.NewCharacterTable()
	characterTable.Dispose()

	codeTable, err := characterTable.GetCharacterLocTable()
	if err != nil {
		return err
	}

	return decoderBase(lockitFileInfo, codeTable)
}

func decoderBase(lockitFileInfo interactions.IGameDataInfo, codeTable string) error {
	handler := newLockitHandler()
	defer handler.dispose()

	executable, err := handler.getLockitFileHandler()
	if err != nil {
		return err
	}

	targetFile := lockitFileInfo.GetGameData().FullFilePath

	extractLocation := lockitFileInfo.GetExtractLocation()

	if err := extractLocation.ProvideTargetPath(); err != nil {
		return err
	}

	args := []string{"-t", codeTable, targetFile, extractLocation.TargetFile}

	if err := lib.RunCommand(executable, args); err != nil {
		return err
	}

	return nil
}
