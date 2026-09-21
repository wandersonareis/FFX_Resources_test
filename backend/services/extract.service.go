package services

import (
	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats"
	"ffxresources/backend/interactions"
	"ffxresources/backend/models"
	"ffxresources/backend/spira"
	"fmt"
)

type ExtractService struct {
	dirExtractService   IDirectoryService
	NotificationService INotificationService
	ProgressService     IProgressService
}

func NewExtractService(notificationService INotificationService, progressService IProgressService) *ExtractService {
	return &ExtractService{
		NotificationService: notificationService,
		ProgressService:     progressService,
	}
}

// Extract monta o nó do caminho sob demanda (sem store global da árvore) e
// delega para arquivo ou diretório.
func (e *ExtractService) Extract(path string) error {
	if err := common.CheckArgumentNil(path, "path"); err != nil {
		return err
	}

	node, err := e.newNode(path)
	if err != nil {
		return err
	}

	var cb func(node *fileFormats.MapNode) error

	switch node.Data.Source.Type {
	case models.Folder:
		cb = func(n *fileFormats.MapNode) error {
			return e.extractDirectory(n, path)
		}
	default:
		cb = e.extractFile
	}

	if err := cb(node); err != nil {
		e.NotificationService.NotifyError(err)
		return err
	}

	return nil
}

func (e *ExtractService) newNode(path string) (*fileFormats.MapNode, error) {
	formatter := interactions.NewInteractionService().TextFormatter()
	node, err := spira.BuildNode(path, formatter)
	if err != nil {
		return nil, fmt.Errorf("node not found for path %s: %w", path, err)
	}
	if node == nil || node.Data == nil {
		return nil, fmt.Errorf("node is invalid for path: %s", path)
	}
	return node, nil
}

func (e *ExtractService) extractFile(node *fileFormats.MapNode) error {
	if node == nil || node.Data == nil {
		return fmt.Errorf("node is invalid")
	}

	if node.Data.Source.Type == models.Folder {
		return fmt.Errorf("node is not a file")
	}

	if node.Data.FileProcessor == nil {
		return fmt.Errorf("file processor is nil")
	}

	if err := node.Data.FileProcessor.Extract(); err != nil {
		return err
	}

	e.NotificationService.NotifySuccess(fmt.Sprintf("File %s extracted successfully!", node.Data.Source.Name))
	return nil
}

func (e *ExtractService) extractDirectory(node *fileFormats.MapNode, path string) error {
	if node == nil || node.Data == nil {
		return fmt.Errorf("node is invalid")
	}

	if node.Data.Source.Type != models.Folder {
		return fmt.Errorf("node is not a directory")
	}

	if e.dirExtractService == nil {
		e.dirExtractService = NewDirectoryExtractService(e.NotificationService, e.ProgressService)
	}

	formatter := interactions.NewInteractionService().TextFormatter()
	store := NewNodeStore(spira.CreateNodeMap(path, formatter))

	if err := e.dirExtractService.ProcessDirectory(node.Data.Source.Path, store); err != nil {
		return err
	}

	e.NotificationService.NotifySuccess(fmt.Sprintf("Directory %s extracted successfully!", node.Data.Source.Name))
	return nil
}
