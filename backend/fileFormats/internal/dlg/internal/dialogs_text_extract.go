package internal

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

	targetFile := dialogsFileInfo.GameData.FullFilePath

	extractLocation := dialogsFileInfo.ExtractLocation

	err = extractLocation.ProvideTargetPath()
	if err != nil {
		return err
	}

	args, codeTable, err := decoderArgs()
	if err != nil {
		return err
	}

	defer common.RemoveFile(codeTable)

	args = append(args, targetFile, extractLocation.TargetFile)

	err = lib.RunCommand(handler, args)
	if err != nil {
		return err
	}

	return nil
}
