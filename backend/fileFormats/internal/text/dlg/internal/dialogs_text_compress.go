package internal

import (
	"ffxresources/backend/core/encoding"
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
)

func DialogsFileCompressor(dialogsFileInfo interactions.IGameDataInfo) error {
	encoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextDlgEncoding(dialogsFileInfo.GetGameData().Type)
	defer encoding.Dispose()
	
	/* handler := newDialogsHandler(dialogsFileInfo.GetGameData().Type)
	defer handler.dispose() */

	//executable, err := handler.getDialogsHandler()
	executable, err := encoding.FetchDlgHandler().FetchDlgTextsHandler()
	if err != nil {
		return err
	}

	translateLocation := dialogsFileInfo.GetTranslateLocation()
	importLocation := dialogsFileInfo.GetImportLocation()

	if err := translateLocation.Validate(); err != nil {
		return err
	}

	if err := importLocation.ProvideTargetPath(); err != nil {
		return err
	}

	/* codeTableHandler := new(util.CharacterTable)
	
	codeTable, err := codeTableHandler.GetFfx2CharacterTable()
	if err != nil {
		return fmt.Errorf("failed to get code table: %w", err)
	}
	
	defer codeTableHandler.Dispose(codeTable) */
	
	targetFile := dialogsFileInfo.GetGameData().FullFilePath

	args := []string{"-i", "-t", encoding.FetchEncoding(), targetFile, translateLocation.TargetFile, importLocation.TargetFile}


	if err := lib.RunCommand(executable, args); err != nil {
		return err
	}

	return nil
}
