package formats

import "ffxresources/backend/lib"

func dialogsUnpacker(dialogsFileInfo lib.FileInfo) error {
	handler, err := getDialogsHandler()
	if err != nil {
		return err
	}

	defer lib.RemoveFile(handler)

	targetFile := dialogsFileInfo.AbsolutePath
	outputFile := dialogsFileInfo.ExtractLocation.TargetFile
	outputPath := dialogsFileInfo.ExtractLocation.TargetPath

	err = lib.EnsurePathExists(outputPath)
	if err != nil {
		return err
	}

	args, codeTable, err := decoderArgs()
	if err != nil {
		return err
	}

	defer lib.RemoveFile(codeTable)

	args = append(args, targetFile, outputFile)

	err = lib.RunCommand(handler, args)
	if err != nil {
		return err
	}

	return nil
}
