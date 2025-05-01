package spira

import (
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats"
	"ffxresources/backend/interfaces"
)

type dataSetter interface {
	SetNodeData(data *fileFormats.DataInfo)
}

func addTreeNodeData[T dataSetter](item T, source interfaces.ISource, destination locations.IDestination) {
	fileProcessor := fileFormats.NewFileProcessor(source, destination)

	gameDataInfo := &fileFormats.DataInfo{
		Source:    source.Get(),
		Extract:   destination.Extract(),
		Translate: destination.Translate().Get(),

		FileProcessor: fileProcessor,
	}

	item.SetNodeData(gameDataInfo)
}
