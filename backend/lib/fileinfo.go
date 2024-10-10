package lib

func CreateFileInfo(source *Source) (FileInfo, error) {
	var nodeData = FileInfo{
		Name:         source.Name,
		Size:         source.Size,
		Type:         source.Type,
		IsDir:        source.IsDir,
		Parent:       source.Parent,
		Extension:    source.Extension,
		AbsolutePath: source.FullPath,
	}

	return nodeData, nil
}
