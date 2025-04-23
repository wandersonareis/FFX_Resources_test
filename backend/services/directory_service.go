package services

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/interfaces"
	"io/fs"
	"path/filepath"
)

type (
	IDirectoryService interface {
		ProcessDirectory(targetPath string, pathMap *NodeStore) error
	}

	directoryService struct {
		notifierService INotificationService
		progressService IProgressService
	}

	directoryExtractService struct {
		directoryService *directoryService
	}
	directoryCompressService struct {
		directoryService *directoryService
	}
)

func NewDirectoryService(notificationService INotificationService, progressService IProgressService) *directoryService {
	return &directoryService{notifierService: notificationService, progressService: progressService}
}

func NewDirectoryExtractService(notifier INotificationService, progressService IProgressService) IDirectoryService {
	return &directoryExtractService{
		directoryService: NewDirectoryService(notifier, progressService),
	}
}

func NewDirectoryCompressService(notifier INotificationService, progressService IProgressService) IDirectoryService {
	return &directoryCompressService{
		directoryService: NewDirectoryService(notifier, progressService),
	}
}

func (e *directoryExtractService) ProcessDirectory(targetPath string, pathMap *NodeStore) error {
	return e.directoryService.processDirectoryCommon(targetPath, pathMap, func(processor interfaces.IFileProcessor) error {
		return processor.Extract()
	})
}

func (e *directoryCompressService) ProcessDirectory(targetPath string, pathMap *NodeStore) error {
	return e.directoryService.processDirectoryCommon(targetPath, pathMap, func(processor interfaces.IFileProcessor) error {
		return processor.Compress()
	})
}

func (d *directoryService) processDirectoryCommon(
	targetPath string,
	pathMap *NodeStore,
	operation func(interfaces.IFileProcessor) error,
) error {
	if err := common.CheckArgumentNil(targetPath, "targetPath"); err != nil {
		return err
	}

	if err := common.CheckArgumentNil(pathMap, "pathMap"); err != nil {
		return err
	}

	cb := func() error {
		return d.processDirectory(targetPath, pathMap, operation)
	}

	if err := common.RecoverFn(cb); err != nil {
		return err
	}

	return nil
}

func (d *directoryService) processDirectory(targetPath string, pathMap *NodeStore, operation func(interfaces.IFileProcessor) error) error {
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

	d.progressService.Stop()
	d.progressService.SetMax(filesProcessorList.GetLength())
	d.progressService.Start()

	filesProcessorList.ForEach(func(processor interfaces.IFileProcessor) {
		if e := operation(processor); e != nil {
			d.notifierService.NotifyError(e)
			return
		}
		d.progressService.StepFile(processor.GetSource().Get().Name)
	})

	return nil
}
