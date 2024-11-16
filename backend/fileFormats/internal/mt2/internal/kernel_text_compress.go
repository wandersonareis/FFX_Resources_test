package internal

import (
	"ffxresources/backend/fileFormats/util"
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
	"fmt"
)

func KernelTextPacker(gameData interactions.IGameDataInfo) error {
	gamePart := interactions.NewInteraction().GamePart.GetGamePart()

	handler := newKernelHandler(gamePart)
	defer handler.Dispose()

	executable, err := handler.getKernelFileHandler()
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

	characterTable := util.NewCharacterTable()
	characterTable.Dispose()

	codeTable, err := characterTable.GetFfx2CharacterTable()
	if err != nil {
		return fmt.Errorf("failed to get code table: %w", err)
	}

	targetFile := gameData.GetGameData().FullFilePath

	args := []string{"-i", "-t", codeTable, targetFile, translateLocation.TargetFile, importLocation.TargetFile}

	if err := lib.RunCommand(executable, args); err != nil {
		return err
	}

	return nil
}
