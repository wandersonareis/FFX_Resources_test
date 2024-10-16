package formats

import "ffxresources/backend/lib"

func dialogsTextPacker(dialogsFileInfo *lib.FileInfo) error {
	handler, err := getDialogsHandler()
	if err != nil {
		return err
	}

	defer lib.RemoveFile(handler)

	targetFile := dialogsFileInfo.AbsolutePath
	extractedFile := dialogsFileInfo.TranslateLocation.TargetFile
	translatedFile := dialogsFileInfo.ImportLocation.TargetFile
	translatedPath := dialogsFileInfo.ImportLocation.TargetPath
	err = lib.EnsurePathExists(translatedPath)
	if err != nil {
		return err
	}

	args, codeTable, err := encoderArgs()
	if err != nil {
		return err
	}

	defer lib.RemoveFile(codeTable)

	args = append(args, targetFile, extractedFile, translatedFile)

	err = lib.RunCommand(handler, args)
	if err != nil {
		return err
	}

	return nil
}
