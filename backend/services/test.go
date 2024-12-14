package services

import (
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats"
	"ffxresources/backend/interactions"
	"ffxresources/backend/spira"
	"fmt"
)

func TestExtractDir(path string, testExtract, testCompress bool) {
	source, err := locations.NewSource(path, interactions.NewInteraction().GamePart.GetGamePart())
	if err != nil {
		fmt.Println(err)
		return
	}

	tree := components.NewEmptyList[spira.TreeNode]()

	err = spira.BuildFileTree(tree, source)
	if err != nil {
		fmt.Println(err)
		return
	}

	destination := locations.NewDestination()

	testRun := func(_ int, n spira.TreeNode) {
		if testExtract {
			extractService := NewExtractService()
			extractService.Extract(source.Get().Path)

			/*fileProcessor := fileFormats.NewFileExtractor(source, destination)
			if fileProcessor != nil {
				if err := fileProcessor.Extract(); err != nil {
					fmt.Println(err)
				}
			}*/
		}

		if testCompress {
			compressService := NewCompressService()
			compressService.Compress(source, destination)
		}
	}

	tree.ParallelForEach(testRun)
}

func TestExtractFile(path string, testExtract, testCompress bool) {
	source, err := locations.NewSource(path, interactions.NewInteraction().GamePart.GetGamePart())
	if err != nil {
		fmt.Println(err)
		return
	}

	destination := locations.NewDestination()

	if testExtract {
		fileProcessor := fileFormats.NewFileExtractor(source, destination)
		if fileProcessor == nil {
			fmt.Println("invalid file type")
			return
		}
		fileProcessor.Extract()
	}

	if testCompress {
		fileProcessor := fileFormats.NewFileCompressor(source, destination)
		if fileProcessor == nil {
			fmt.Println("invalid file type")
			return
		}

		fileProcessor.Compress()
	}
}

/* func TestExtractFileDev(path string, testExtract, testCompress bool) {
	dataInfo := interactions.NewGameDataInfo(path)
	gamePart := interactions.NewInteraction().GamePart.GetGamePart()
	source, err := core.NewSource(path, gamePart)
	if err != nil {
		fmt.Println(err)
		return
	}

	destination := target.NewDestination(source)


	if testExtract {
		fileProcessor := fileFormats.NewFileProcessorDev(dataInfo, *source, destination)
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
} */
