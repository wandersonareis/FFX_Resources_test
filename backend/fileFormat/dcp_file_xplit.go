package fileFormat

import "ffxresources/backend/lib"

func dcpFileXpliter(fileInfo lib.FileInfo) error {
	xpliterHandler, err := getDcpFileXpliterDev()
	if err != nil {
		return err
	}

	defer lib.RemoveFile(xpliterHandler)

	targetPath := fileInfo.ExtractLocation.TargetFile
	outputPath := fileInfo.ExtractLocation.TargetPath
	lib.EnsurePathExists(outputPath)

	args, err := dcpXpliterArgs()
	if err != nil {
		return err
	}

	args = append(args, fileInfo.AbsolutePath, targetPath)

	err = lib.RunCommand(xpliterHandler, args)
	if err != nil {
		return err
	}

	return nil
}
