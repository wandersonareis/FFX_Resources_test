package lib

/* func CreateFileInfo(source *Source) (FileInfo, error) {
	var nodeData = FileInfo{
		Name:              source.Name,
		NamePrefix:        source.NamePrefix,
		Size:              source.Size,
		Type:              source.Type,
		IsDir:             source.IsDir,
		Parent:            source.Parent,
		Extension:         source.Extension,
		AbsolutePath:      source.FullPath,
		ExtractLocation:   *NewInteraction().ExtractLocation,
		TranslateLocation: *NewInteraction().TranslateLocation,
		ImportLocation:    *NewInteraction().ImportLocation,
	}

	return nodeData, nil
} */

func UpdateFileInfoFromSource(fileInfo *FileInfo, source *Source) {
	fileInfo.Name = source.Name
	fileInfo.NamePrefix = source.NamePrefix
	fileInfo.Size = source.Size
	fileInfo.Type = source.Type
	fileInfo.IsDir = source.IsDir
	fileInfo.Parent = source.Parent
	fileInfo.Extension = source.Extension
	fileInfo.AbsolutePath = source.FullPath
	fileInfo.ExtractLocation = *NewInteraction().ExtractLocation
	fileInfo.TranslateLocation = *NewInteraction().TranslateLocation
	fileInfo.ImportLocation = *NewInteraction().ImportLocation
}