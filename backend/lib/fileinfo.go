package lib

func CreateFileInfoFromPath(path string) (*FileInfo, error) {
	fileInfo := &FileInfo{}
	source, err := NewSource(path)
	if err != nil {
		return fileInfo, err
	}

	UpdateFileInfoFromSource(fileInfo, source)

	return fileInfo, nil
}

func UpdateFileInfoFromSource(fileInfo *FileInfo, source *Source) {
	fileInfo.Name = source.Name
	fileInfo.NamePrefix = source.NamePrefix
	fileInfo.Size = source.Size
	fileInfo.Type = source.Type
	fileInfo.IsDir = source.IsDir
	fileInfo.Parent = source.Parent
	fileInfo.Extension = source.Extension
	fileInfo.AbsolutePath = source.FullPath
	fileInfo.ExtractLocation = *NewExtractLocation()
	fileInfo.TranslateLocation = *NewTranslateLocation()
	fileInfo.ImportLocation = *NewImportLocation()
}