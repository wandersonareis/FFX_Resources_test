package services

import (
	"ffxresources/backend/lib"
	"ffxresources/backend/spira"
	"fmt"
)

func TestExtractFile(path string, testExtract, testCompress bool) {
	source, err := lib.NewSource(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	tree := make([]lib.TreeNode, 0, 1)

	err = spira.BuildFileTree(&tree, source)
	if err != nil {
		fmt.Println(err)
		return
	}

	fileInfo := tree[0].Data

	if testExtract {
		extractService := NewExtractService()
		extractService.Extract(fileInfo)
	}

	if testCompress {
		compressService := NewCompressService()
		compressService.Compress(fileInfo)
	}
}
