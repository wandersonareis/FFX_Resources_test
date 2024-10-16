package formats

import "ffxresources/backend/lib"

func dcpFileJoiner(fileInfo lib.FileInfo, reimportedDcpPartsDirectory string) error {
	xpliterHandler, err := GetDcpXplitHandler()
	if err != nil {
		return err
	}

	originalDcpFile := fileInfo.AbsolutePath

	reimportFile := fileInfo.ImportLocation.TargetFile
	reimportFilePath := fileInfo.ImportLocation.TargetPath
	lib.EnsurePathExists(reimportFilePath)

	args, err := dcpJoinerArgs()
	if err != nil {
		return err
	}

	args = append(args, originalDcpFile, reimportedDcpPartsDirectory, reimportFile)

	err = lib.RunCommand(xpliterHandler, args)
	if err != nil {
		return err
	}

	return nil
}
