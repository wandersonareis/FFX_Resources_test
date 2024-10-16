package formats

import "ffxresources/backend/lib"

func FileFormatterWriter(formatter lib.ITextFormatter, fileInfo lib.FileInfo) {

}

/* func (t TxtFormatter) Write(fileInfo lib.FileInfo, targetDirectory string) (string, string) {
	var outputFile, outputPath string

	switch fileInfo.Type {
	case lib.Dcp:
		outputFile, outputPath = t.provideDcpOutput(fileInfo, targetDirectory)
	case lib.DcpParts:
		outputFile, outputPath = t.provideDcpPartsOutput(fileInfo, targetDirectory)
	default:
		outputFile, outputPath = t.provideDefaultOutput(fileInfo, targetDirectory)
	}

	return outputFile, outputPath
}

func (t TxtFormatter) provideDefaultOutput(fileInfo lib.FileInfo, targetDirectory string) (string, string) {
	extractedFile, extractedPath := lib.GenerateExtractedOutput(fileInfo, targetDirectory, "", lib.DEFAULT_TEXT_EXTENSION)

	return extractedFile, extractedPath
}

func (t TxtFormatter) provideDcpOutput(fileInfo lib.FileInfo, targetDirectory string) (string, string) {
	path := lib.PathJoin(targetDirectory, fileInfo.RelativePath)

	outputFile := path
	outputPath := lib.GetDir(path)

	return outputFile, outputPath
}

func (t TxtFormatter) provideDcpPartsOutput(fileInfo lib.FileInfo, targetDirectory string) (string, string) {
	outputFile := lib.PathJoin(targetDirectory, lib.DCP_PARTS_TARGET_DIR_NAME, lib.AddExtension(fileInfo.Name, lib.DEFAULT_TEXT_EXTENSION))

	extractedPath := lib.GetDir(outputFile)

	return outputFile, extractedPath
}

func (t TxtFormatter) Read(fileInfo lib.FileInfo, targetDirectory string) (string, string) {

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
