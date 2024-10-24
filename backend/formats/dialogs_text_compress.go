package formats

import (
	"ffxresources/backend/lib"
	"fmt"
)

func dialogsTextCompress(dialogsFileInfo *lib.FileInfo) error {
	handler, err := getDialogsHandler()
	if err != nil {
		return err
	}

	defer lib.RemoveFile(handler)

	if !dialogsFileInfo.TranslateLocation.TargetFileExists() {
		msg := "Dialogs translated file does not exist"
		lib.NotifyWarn(msg)
		return nil
	}


	fmt.Println("dialogsTextCompress")
	targetFile := dialogsFileInfo.AbsolutePath
	targetTranslatedFile := dialogsFileInfo.TranslateLocation.TargetFile
	targetReimportFile := dialogsFileInfo.ImportLocation.TargetFile
	targetReimportPath := dialogsFileInfo.ImportLocation.TargetPath
	err = lib.EnsurePathExists(targetReimportPath)
	if err != nil {
		return err
	}

	args, codeTable, err := encoderArgs()
	if err != nil {
		return err
	}

	defer lib.RemoveFile(codeTable)

	args = append(args, targetFile, targetTranslatedFile, targetReimportFile)

	err = lib.RunCommand(handler, args)
	if err != nil {
		return err
	}

	return nil
}
