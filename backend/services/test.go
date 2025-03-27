package services

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/spira"
	"fmt"
)

func TestExtractDir(path string, testExtract, testCompress bool) {
	if !testExtract && !testCompress {
		return
	}

	source, err := locations.NewSource(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	
	gamaLocationAux := interactions.NewInteractionService().GameLocation.GetTargetDirectory()
	defer interactions.NewInteractionService().GameLocation.SetTargetDirectory(gamaLocationAux)
	
	tree := components.NewEmptyList[spira.TreeNode]()
	formatter := formatters.NewTxtFormatter()
	
	interactions.NewInteractionService().GameLocation.SetTargetDirectory(path)
	
	gameVersion := interactions.NewInteractionService().FFXGameVersion().GetGameVersion()
	NodeMap = spira.CreateNodeMap(gameVersion, formatter)

	testRun := func(n spira.TreeNode) {
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
			compressService.Compress(source.Get().Path)
		}
	}

	tree.ParallelForEach(testRun)
}

func TestExtractFile(path string, testExtract, testCompress bool) {
	common.CheckArgumentNil(path, "path")

	if !testExtract && !testCompress {
		return
	}

	source, err := locations.NewSource(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	destination := locations.NewDestination()
	destination.InitializeLocations(source, formatters.NewTxtFormatter())

	if testExtract {
		fileProcessor := fileFormats.NewFileExtractor(source, destination)
		if fileProcessor == nil {
			fmt.Println("invalid file type")
			return
		}
		if err := fileProcessor.Extract(); err != nil {
			fmt.Println(err)
			return
		}
	}

	if testCompress {
		fileProcessor := fileFormats.NewFileCompressor(source, destination)
		if fileProcessor == nil {
			fmt.Println("invalid file type")
			return
		}

		if err := fileProcessor.Compress(); err != nil {
			fmt.Println(err)
			return
		}
	}
}

/* func TestExtractFileDev(path string, testExtract, testCompress bool) {
	dataInfo := interactions.NewGameDataInfo(path)
	gamePart := interactions.NewInteraction().FFXGameVersion.FFXGameVersion()
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
