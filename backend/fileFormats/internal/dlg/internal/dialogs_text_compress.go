package internal

import (
	"ffxresources/backend/common"
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
)

func DialogsTextCompress(dialogsFileInfo *interactions.GameDataInfo) error {
	handler, err := getDialogsHandler()
	if err != nil {
		return err
	}

	defer common.RemoveFile(handler)

	translateLocation := dialogsFileInfo.TranslateLocation
	importLocation := dialogsFileInfo.ImportLocation

	if err := translateLocation.Validate(); err != nil {
		return err
	}

	if err := importLocation.ProvideTargetPath(); err != nil {
		return err
	}

	args, codeTable, err := encoderArgs()
	if err != nil {
		return err
	}

	defer common.RemoveFile(codeTable)

	targetFile := dialogsFileInfo.GameData.FullFilePath

	args = append(args, targetFile, translateLocation.TargetFile, importLocation.TargetFile)

	if err := lib.RunCommand(handler, args); err != nil {
		return err
	}

	return nil
}
