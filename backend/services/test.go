package services

import (
	"ffxresources/backend/core"
	"ffxresources/backend/formats"
	"ffxresources/backend/interactions"
	"ffxresources/backend/spira"
	"fmt"
)

func TestExtractDir(path string, testExtract, testCompress bool) {
	source, err := core.NewSource(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	tree := make([]interactions.TreeNode, 0, 1)

	err = spira.BuildFileTree(&tree, source)
	if err != nil {
		fmt.Println(err)
		return
	}

	dataInfo := &tree[0].Data

	if testExtract {
		extractService := NewExtractService()
		extractService.Extract(dataInfo)
	}

	if testCompress {
		compressService := NewCompressService()
		compressService.Compress(dataInfo)
	}
}

func TestExtractFile(path string, testExtract, testCompress bool) {
	dataInfo := interactions.NewGameDataInfo(path)

	if testExtract {
		fileProcessor := formatsDev.NewFileExtractor(dataInfo)
		if fileProcessor == nil {
			fmt.Println("invalid file type")
			return
		}
		fileProcessor.Extract()
	}

	if testCompress {
		fileProcessor := formatsDev.NewFileCompressor(dataInfo)
		if fileProcessor == nil {
			fmt.Println("invalid file type")
			return
		}

		fileProcessor.Compress()
	}
}
