package formats

import (
	"ffxresources/backend/common"
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
	"fmt"
)

func dialogsTextCompress(dialogsFileInfo *interactions.GameDataInfo) error {
	handler, err := getDialogsHandler()
	if err != nil {
		return err
	}

	defer common.RemoveFile(handler)

	if !dialogsFileInfo.TranslateLocation.TargetFileExists() {
		msg := "Dialogs translated file does not exist"
		lib.NotifyWarn(msg)
		return nil
	}


	fmt.Println("dialogsTextCompress")
	targetFile := dialogsFileInfo.GameData.AbsolutePath
	targetTranslatedFile := dialogsFileInfo.TranslateLocation.TargetFile
	targetReimportFile := dialogsFileInfo.ImportLocation.TargetFile
	targetReimportPath := dialogsFileInfo.ImportLocation.TargetPath
	err = common.EnsurePathExists(targetReimportPath)
	if err != nil {
		return err
	}

	args, codeTable, err := encoderArgs()
	if err != nil {
		return err
	}

	defer common.RemoveFile(codeTable)

	args = append(args, targetFile, targetTranslatedFile, targetReimportFile)

	err = lib.RunCommand(handler, args)
	if err != nil {
		return err
	}

	return nil
}
