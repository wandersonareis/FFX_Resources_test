package formats

import (
	"ffxresources/backend/common"
	"ffxresources/backend/lib"
)

func kernelTextPacker(kernelFileInfo *lib.FileInfo) error {
	handler, err := getKernelFileHandler()
	if err != nil {
		return err
	}

	defer common.RemoveFile(handler)

	if !kernelFileInfo.TranslateLocation.TargetFileExists() {
		msg := "translated file does not exist"
		lib.NotifyWarn(msg)
		return nil
	}

	targetFile := kernelFileInfo.AbsolutePath
	extractedFile := kernelFileInfo.ExtractLocation.TargetFile
	translatedFile := kernelFileInfo.TranslateLocation.TargetFile
	translatedPath := kernelFileInfo.TranslateLocation.TargetPath

	err = common.EnsurePathExists(translatedPath)
	if err != nil {
		return err
	}

	args, codeTable, err := encoderArgs()
	if err != nil {
		return err
	}

	defer common.RemoveFile(codeTable)

	args = append(args, targetFile, extractedFile, translatedFile)

	err = lib.RunCommand(handler, args)
	if err != nil {
		return err
	}

	return nil
}
