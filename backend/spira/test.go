package spira

import (
	"ffxresources/backend/lib"
	"fmt"
)

func TestExtractFile(path string) {
	source, err := lib.NewSource(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	node, err := ListFilesAndDirectories(source, "")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(node)
	fileInfo := node[0].Data

	fileProcessor := NewFileProcessor(fileInfo)
	if fileProcessor == nil {
		lib.NotifyError(fmt.Errorf("invalid file type: %s", fileInfo.Name))
		return
	}

	//fileProcessor.Extract()
	//fileProcessor.Compress()
}
