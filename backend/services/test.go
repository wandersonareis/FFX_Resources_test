package services

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core"
	"ffxresources/backend/fileFormats"
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

	worker := common.NewWorker[interactions.TreeNode]()
	worker.ParallelForEach(&tree, func(_ int, n interactions.TreeNode) {
		dataInfo := &n.Data

		if testExtract {
			extractService := NewExtractService()
			extractService.Extract(dataInfo)
		}

		if testCompress {
			compressService := NewCompressService()
			compressService.Compress(dataInfo)
		}
	})
}

func TestExtractFile(path string, testExtract, testCompress bool) {
	dataInfo := interactions.NewGameDataInfo(path)

	if testExtract {
		fileProcessor := fileFormats.NewFileExtractor(dataInfo)
		if fileProcessor == nil {
			fmt.Println("invalid file type")
			return
		}
		fileProcessor.Extract()
	}

	if testCompress {
		fileProcessor := fileFormats.NewFileCompressor(dataInfo)
		if fileProcessor == nil {
			fmt.Println("invalid file type")
			return
		}

		fileProcessor.Compress()
	}
}
