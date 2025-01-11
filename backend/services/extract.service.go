package services

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"ffxresources/backend/models"
	"ffxresources/backend/notifications"
	"fmt"
	"io/fs"
	"path/filepath"
)

type ExtractService struct{}

func NewExtractService() *ExtractService {
	return &ExtractService{}
}

func (e *ExtractService) Extract(path string) {
	defer func() {
		l := logger.Get()
		if r := recover(); r != nil {
			l.Error().
				Interface("recover", r).
				Str("file", common.GetFileName(path)).
				Msg("Panic occurred during extraction")

			notifications.NotifyError(fmt.Errorf("panic occurred: %v", r))
		}
	}()

	node, ok := NodeMap["path"]
	if !ok {
		notifications.NotifyError(fmt.Errorf("node not found for path: %s", path))
		return
	}

	if node.Data.Source.Type == models.Folder {
		if err := e.processDirectory(path); err != nil {
			notifications.NotifyError(err)
			return
		}

		notifications.NotifySuccess(fmt.Sprintf("Directory %s extracted successfully!", node.Label))
		return
	}

	processor := node.Data.FileProcessor
	if processor != nil {
		if err := processor.Extract(); err != nil {
			notifications.NotifyError(err)
			return
		}
		notifications.NotifySuccess(fmt.Sprintf("File %s extracted successfully!", node.Label))
	}

}

func (e *ExtractService) processDirectory(targetPath string) error {
	filesProcessorList := components.NewEmptyList[interfaces.IFileProcessor]()

	err := filepath.WalkDir(targetPath, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if node, ok := NodeMap[path]; ok {
			if node.Data.Source.IsDir {
				return nil
			}

			processor := node.Data.FileProcessor
			if processor != nil {
				filesProcessorList.Add(processor)
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	context := interactions.NewInteractionService().Ctx

	progress := common.NewProgress(context)
	progress.SetMax(filesProcessorList.GetLength())
	progress.Start()

	filesProcessorList.ParallelForEach(func(_ int, processor interfaces.IFileProcessor) {
		if err := processor.Extract(); err != nil {
			notifications.NotifyError(err)
			return
		}

		progress.StepFile(processor.Source().Get().Name)
	})

	//progress.Stop()

	return nil
}
