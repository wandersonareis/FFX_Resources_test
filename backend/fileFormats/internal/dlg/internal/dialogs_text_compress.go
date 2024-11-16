package internal

import (
	"ffxresources/backend/fileFormats/util"
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
	"fmt"
)

func DialogsFileCompressor(gameData interactions.IGameDataInfo) error {
	handler := newDialogsHandler(gameData.GetGameData().Type)
	defer handler.dispose()

	executable, err := handler.getDialogsHandler()
	if err != nil {
		return err
	}

	translateLocation := gameData.GetTranslateLocation()
	importLocation := gameData.GetImportLocation()

	if err := translateLocation.Validate(); err != nil {
		return err
	}

	if err := importLocation.ProvideTargetPath(); err != nil {
		return err
	}

	codeTableHandler := new(util.CharacterTable)
	defer codeTableHandler.Dispose()

	codeTable, err := codeTableHandler.GetFfx2CharacterTable()
	if err != nil {
		return fmt.Errorf("failed to get code table: %w", err)
	}

	targetFile := gameData.GetGameData().FullFilePath

	args := []string{"-i", "-t", codeTable, targetFile, translateLocation.TargetFile, importLocation.TargetFile}

	//args = append(args, codeTable, targetFile, translateLocation.TargetFile, importLocation.TargetFile)

	if err := lib.RunCommand(executable, args); err != nil {
		return err
	}

	return nil
}
