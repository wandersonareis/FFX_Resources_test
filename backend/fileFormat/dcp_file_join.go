package fileFormat

import "ffxresources/backend/lib"

func dcpFileJoiner(fileInfo lib.FileInfo, translatedPartsDir string) error {
	xpliterHandler, err := getDcpFileXpliterDev()
	if err != nil {
		return err
	}

	targetFile := fileInfo.AbsolutePath
	translatedFile := fileInfo.TranslatedFile
	translatedPartsDirectory := translatedPartsDir

	translatedFilePath := fileInfo.TranslatedPath
	lib.EnsurePathExists(translatedFilePath)

	args, err := dcpJoinerArgs()
	if err != nil {
		return err
	}

	args = append(args, targetFile, translatedPartsDirectory, translatedFile)

	err = lib.RunCommand(xpliterHandler, args)
	if err != nil {
		return err
	}

	return nil
}
