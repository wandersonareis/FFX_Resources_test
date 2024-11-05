package dlg_internal

import (
	"ffxresources/backend/common"
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
)

func DialogsUnpacker(dialogsFileInfo *interactions.GameDataInfo) error {
	handler, err := getDialogsHandler()
	if err != nil {
		return err
	}

	defer common.RemoveFile(handler)

	targetFile := dialogsFileInfo.GameData.AbsolutePath
	outputFile := dialogsFileInfo.ExtractLocation.TargetFile
	outputPath := dialogsFileInfo.ExtractLocation.TargetPath

	err = common.EnsurePathExists(outputPath)
	if err != nil {
		return err
	}

	args, codeTable, err := decoderArgs()
	if err != nil {
		return err
	}

	defer common.RemoveFile(codeTable)

	args = append(args, targetFile, outputFile)

	err = lib.RunCommand(handler, args)
	if err != nil {
		return err
	}

	return nil
}
