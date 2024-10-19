package formats

import "ffxresources/backend/lib"

func dialogsTextCompress(dialogsFileInfo *lib.FileInfo) error {
	handler, err := getDialogsHandler()
	if err != nil {
		return err
	}

	defer lib.RemoveFile(handler)

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
