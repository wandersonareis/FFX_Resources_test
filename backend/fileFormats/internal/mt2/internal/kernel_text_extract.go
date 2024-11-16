package internal

import (
	"ffxresources/backend/fileFormats/util"
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
	"fmt"
)

func KernelUnpacker(kernelFileInfo interactions.IGameDataInfo) error {
	gamePart := interactions.NewInteraction().GamePart.GetGamePart()

	handler := newKernelHandler(gamePart)
	defer handler.Dispose()

	executable, err := handler.getKernelFileHandler()
	if err != nil {
		return err
	}

	targetFile := kernelFileInfo.GetGameData().FullFilePath
	extractLocation := kernelFileInfo.GetExtractLocation()

	if err := extractLocation.ProvideTargetPath(); err != nil {
		return err
	}

	/* args, err := util.DecoderDlgKrnlArgs()
	if err != nil {
		return err
	} */

	characterTable := util.NewCharacterTable()
	characterTable.Dispose()

	codeTable, err := characterTable.GetFfx2CharacterTable()
	if err != nil {
		return fmt.Errorf("failed to get code table: %w", err)
	}
	
	args := []string{"-e", "-t", codeTable, targetFile, extractLocation.TargetFile}
	//args = append(args, codeTable, targetFile, extractLocation.TargetFile)

	if err := lib.RunCommand(executable, args); err != nil {
		return err
	}

	return nil
}
