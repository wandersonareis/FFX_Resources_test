package fileFormat

import "ffxresources/backend/lib"

type TxtFormatter struct{}

func (t TxtFormatter) Write(fileInfo lib.FileInfo, targetDirectory string) (string, string) {
	targetExtension := ".txt"
	extractedFile, extractedPath := lib.GenerateExtractedOutput(fileInfo, targetDirectory, "", targetExtension)

	return extractedFile, extractedPath
}