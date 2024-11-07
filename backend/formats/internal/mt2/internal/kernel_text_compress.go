package internal

import (
	"ffxresources/backend/common"
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
)

func KernelTextPacker(kernelFileInfo *interactions.GameDataInfo) error {
	handler, err := GetKernelFileHandler()
	if err != nil {
		return err
	}

	defer common.RemoveFile(handler)

	if err := kernelFileInfo.TranslateLocation.Validate(); err != nil {
		lib.NotifyWarn(err.Error())
		return nil
	}

	targetFile := kernelFileInfo.GameData.AbsolutePath
	extractedFile := kernelFileInfo.ExtractLocation.TargetFile
	translateLocation := kernelFileInfo.TranslateLocation

	if err := translateLocation.ProvideTargetPath(); err != nil {
		return err
	}

	args, codeTable, err := encoderArgs()
	if err != nil {
		return err
	}

	defer common.RemoveFile(codeTable)

	args = append(args, targetFile, extractedFile, translateLocation.TargetFile)

	if err := lib.RunCommand(handler, args); err != nil {
		return err
	}

	return nil
}
