package formats

import "ffxresources/backend/lib"

func FileFormatReader(formatter lib.ITextFormatter, fileInfo lib.FileInfo) {

}

/* func (t TxtFormatter) Read(fileInfo lib.FileInfo, targetDirectory string) (string, string) {

	var outputFile, outputPath string

	switch fileInfo.Type {
	case lib.Dcp:
		outputFile, outputPath = t.provideDcpWritePath(fileInfo, targetDirectory)
	//case lib.DcpParts:
		//outputFile, outputPath = t.provideDcpPartsOutput(fileInfo, targetDirectory)
	default:
		outputFile, outputPath = t.provideDefaultWritePath(fileInfo, targetDirectory)
	}

	//return outputFile, outputPath

	/* outputFile = lib.PathJoin(targetDirectory, lib.ChangeExtension(fileInfo.RelativePath, fileInfo.Extension))
	outputPath = filepath.Dir(outputFile)

	return outputFile, outputPath
}

func (t TxtFormatter) provideDefaultWritePath(fileInfo lib.FileInfo, targetDirectory string) (string, string) {
	outputFile := lib.PathJoin(targetDirectory, lib.ChangeExtension(fileInfo.RelativePath, fileInfo.Extension))
	outputPath := filepath.Dir(outputFile)

	return outputFile, outputPath
}

func (t TxtFormatter) provideDcpWritePath(fileInfo lib.FileInfo, targetDirectory string) (string, string) {
	outputFile := lib.PathJoin(targetDirectory, fileInfo.RelativePath)
	outputPath := lib.GetDir(outputFile)

	return outputFile, outputPath
} */
