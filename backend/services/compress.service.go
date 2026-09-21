package services

import (
	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats"
	"ffxresources/backend/interactions"
	"ffxresources/backend/models"
	"ffxresources/backend/spira"
	"fmt"
	"sync"
)

type CompressService struct {
	dirCompressServideOnce sync.Once
	dirCompressService     IDirectoryService
	notifierService        INotificationService
	progressService        IProgressService
}

func NewCompressService(notificationService INotificationService, progressService IProgressService) *CompressService {
	return &CompressService{
		notifierService: notificationService,
		progressService: progressService,
	}
}

// Compress monta o nó do caminho sob demanda (sem store global da árvore) e
// delega para arquivo ou diretório.
func (c *CompressService) Compress(path string) error {
	if err := common.CheckArgumentNil(path, "path"); err != nil {
		return err
	}

	node, err := c.newNode(path)
	if err != nil {
		return err
	}

	var cb func(node *fileFormats.MapNode) error

	switch node.Data.Source.Type {
	case models.Folder:
		cb = func(n *fileFormats.MapNode) error {
			return c.compressDirectory(n, path)
		}
	default:
		cb = c.compressFile
	}

	if err := cb(node); err != nil {
		c.notifierService.NotifyError(err)
		return err
	}

	return nil
}

func (c *CompressService) newNode(path string) (*fileFormats.MapNode, error) {
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

func (c *CompressService) compressFile(node *fileFormats.MapNode) error {
	if node == nil || node.Data == nil {
		return fmt.Errorf("node is invalid")
	}

	if node.Data.FileProcessor == nil {
		return fmt.Errorf("file processor is nil")
	}

	if err := node.Data.FileProcessor.Compress(); err != nil {
		return err
	}

	c.notifierService.NotifySuccess(fmt.Sprintf("File %s compressed successfully!", node.Data.Source.Name))
	return nil
}

func (c *CompressService) compressDirectory(node *fileFormats.MapNode, path string) error {
	if node == nil || node.Data == nil {
		return fmt.Errorf("node is invalid")
	}

	if node.Data.Source.Type != models.Folder {
		return fmt.Errorf("node is not a directory")
	}

	if c.dirCompressService == nil {
		c.dirCompressServideOnce.Do(func() {
			c.dirCompressService = NewDirectoryCompressService(c.notifierService, c.progressService)
		})
	}

	formatter := interactions.NewInteractionService().TextFormatter()
	store := NewNodeStore(spira.CreateNodeMap(path, formatter))

	if err := c.dirCompressService.ProcessDirectory(node.Data.Source.Path, store); err != nil {
		return err
	}

	c.notifierService.NotifySuccess(fmt.Sprintf("Directory %s compressed successfully!", node.Data.Source.Name))
	return nil
}
