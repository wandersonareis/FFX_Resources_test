package lib

import (
	"ffxresources/backend/fileFormats/util"
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
)

func LockitDecoderFfx(lockitFileInfo interactions.IGameDataInfo) error {
	characterTable := util.NewCharacterTable()
	
	codeTable, err := characterTable.GetCharacterOnlyTable()
	if err != nil {
		return err
	}

	defer characterTable.Dispose(codeTable)

	return decoderBase(lockitFileInfo, codeTable)
}

func LockitDecoderLoc(lockitFileInfo interactions.IGameDataInfo) error {
	characterTable := util.NewCharacterTable()
	
	codeTable, err := characterTable.GetCharacterLocTable()
	if err != nil {
		return err
	}
	
	defer characterTable.Dispose(codeTable)
	
	return decoderBase(lockitFileInfo, codeTable)
}

func decoderBase(lockitFileInfo interactions.IGameDataInfo, codeTable string) error {
	handler := NewLockitHandler()
	defer handler.Dispose()

	executable, err := handler.GetLockitFileHandler()
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
