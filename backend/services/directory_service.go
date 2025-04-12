package services

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/notifications"
	"io/fs"
	"path/filepath"
)

type (
	IDirectoryService interface {
		ProcessDirectory(targetPath string, pathMap *NodeStore) error
	}

	directoryExtractService  struct{}
	directoryCompressService struct{}
)

func NewDirectoryExtractService() IDirectoryService {
	return &directoryExtractService{}
}

func NewDirectoryCompressService() IDirectoryService {
	return &directoryCompressService{}
}

func (e *directoryExtractService) ProcessDirectory(targetPath string, pathMap *NodeStore) error {
	return ProcessDirectoryCommon(targetPath, pathMap, func(processor interfaces.IFileProcessor) error {
		return processor.Extract()
	})
}

func (e *directoryCompressService) ProcessDirectory(targetPath string, pathMap *NodeStore) error {
	return ProcessDirectoryCommon(targetPath, pathMap, func(processor interfaces.IFileProcessor) error {
		return processor.Compress()
	})
}

func ProcessDirectoryCommon(
	targetPath string,
	pathMap *NodeStore,
	operation func(interfaces.IFileProcessor) error,
) error {
	common.CheckArgumentNil(targetPath, "targetPath")
	common.CheckArgumentNil(pathMap, "pathMap")

	cb := func() error {
		return processDirectory(targetPath, pathMap, operation)
	}

	if err := common.RecoverFn(cb); err != nil {
		return err
	}

	return nil
}

func processDirectory(targetPath string, pathMap *NodeStore, operation func(interfaces.IFileProcessor) error) error {
	filesProcessorList := components.NewEmptyList[interfaces.IFileProcessor]()
	defer filesProcessorList.Clear()

	err := filepath.WalkDir(targetPath, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if node, ok := pathMap.Get(path); ok {
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

	filesProcessorList.ParallelForEach(func(processor interfaces.IFileProcessor) {
		if e := operation(processor); e != nil {
			notifications.NotifyError(e)
			return
		}
		progress.StepFile(processor.GetSource().Get().Name)
	})

	return nil
}
