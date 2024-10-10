package spira

import (
	"context"
	"ffxresources/backend/fileFormat"
	"ffxresources/backend/lib"
)

func NewFileProcessor(ctx context.Context, fileInfo lib.FileInfo) lib.IFileProcessor {
	fileType := fileInfo.Type

	switch fileType {
	case lib.Dialogs, lib.Tutorial:
		return fileFormat.NewDialogs(ctx, fileInfo)
	case lib.Kernel:
		return fileFormat.NewKernel(ctx, fileInfo)
	case lib.Dcp:
		return fileFormat.NewDcpFile(ctx, fileInfo)
	case lib.Folder:
		return NewSpiraFolder(ctx, fileInfo)
	default:
		return nil
	}
}
