package internal

import (
	"ffxresources/backend/common"
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
)

func KernelUnpacker(kernelFileInfo *interactions.GameDataInfo) error {
	handler, err := GetKernelFileHandler()
	if err != nil {
		return err
	}

	defer common.RemoveFile(handler)

	targetFile := kernelFileInfo.GameData.FullFilePath
	extractLocation := kernelFileInfo.ExtractLocation

	if err := extractLocation.ProvideTargetPath(); err != nil {
		return err
	}

	args, codeTable, err := decoderArgs()
	if err != nil {
		return err
	}

	defer common.RemoveFile(codeTable)

	args = append(args, targetFile, extractLocation.TargetFile)

	if err := lib.RunCommand(handler, args); err != nil {
		return err
	}

	return nil
}
