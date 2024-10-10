package fileFormat

import "ffxresources/backend/lib"

func kernelUnpacker(kernelFileInfo lib.FileInfo) error {
	handler, err := getKernelFileHandler()
	if err != nil {
		return err
	}

	defer lib.RemoveFile(handler)

	targetFile := kernelFileInfo.AbsolutePath
	outputFile := kernelFileInfo.ExtractLocation.TargetFile
	outputPath := kernelFileInfo.ExtractLocation.TargetPath

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
