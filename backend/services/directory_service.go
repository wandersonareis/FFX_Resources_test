package services

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/fileFormats"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/notifications"
	"io/fs"
	"path/filepath"
)

type (
    IDirectoryService interface {
		ProcessDirectory(targetPath string, pathMap fileFormats.TreeMapNode) error
	}

	directoryExtractService struct{}
    directoryCompressService struct{}
)

func (e *directoryExtractService) ProcessDirectory(targetPath string, pathMap fileFormats.TreeMapNode) error {
    return ProcessDirectoryCommon(targetPath, pathMap, func(processor interfaces.IFileProcessor) error {
        return processor.Extract()
    })
}

func (e *directoryCompressService) ProcessDirectory(targetPath string, pathMap fileFormats.TreeMapNode) error {
    return ProcessDirectoryCommon(targetPath, pathMap, func(processor interfaces.IFileProcessor) error {
        return processor.Compress()
    })
}

func ProcessDirectoryCommon(
    targetPath string,
    pathMap fileFormats.TreeMapNode,
    operation func(interfaces.IFileProcessor) error,
) error {
    filesProcessorList := components.NewEmptyList[interfaces.IFileProcessor]()

    err := filepath.WalkDir(targetPath, func(path string, info fs.DirEntry, err error) error {
        if err != nil {
            return err
        }
        if info.IsDir() {
            return nil
        }
        if node, ok := pathMap[path]; ok {
            if node.Data.Source.IsDir {
                return nil
            }
            if node.Data.FileProcessor != nil {
                filesProcessorList.Add(node.Data.FileProcessor)
            }
        }
        return nil
    })
    if err != nil {
        return err
    }

    ctx := interactions.NewInteractionService().Ctx
    progress := common.NewProgress(ctx)
    progress.SetMax(filesProcessorList.GetLength())
    progress.Start()

    filesProcessorList.ParallelForEach(func(_ int, processor interfaces.IFileProcessor) {
        if e := operation(processor); e != nil {
            notifications.NotifyError(e)
            return
        }
        progress.StepFile(processor.Source().Get().Name)
    })
    return nil
}
