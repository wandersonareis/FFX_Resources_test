package fileFormat

import "ffxresources/backend/lib"

func dialogsTextPacker(dialogsFileInfo lib.FileInfo) error {
	handler, err := getDialogsFileHandler()
	if err != nil {
		return err
	}

	defer lib.RemoveFile(handler)

	targetFile := dialogsFileInfo.AbsolutePath
	extractedFile := dialogsFileInfo.ExtractLocation.TargetFile
	translatedFile := dialogsFileInfo.TranslatedFile
	translatedPath := dialogsFileInfo.TranslatedPath
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
