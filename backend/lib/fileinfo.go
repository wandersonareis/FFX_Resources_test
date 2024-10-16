package lib

func CreateFileInfo(source *Source) (FileInfo, error) {
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
}
