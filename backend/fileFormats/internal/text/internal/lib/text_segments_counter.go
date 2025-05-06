package lib

import (
	"ffxresources/backend/core/components"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/models"
	"fmt"
)

func TextSegmentsCounter(file string, fileType models.NodeType, gameVersion models.GameVersion) (int, error) {
	switch fileType {
	case models.Dialogs, models.DialogsSpecial, models.Tutorial:
		return dialogsSegmentsCounter(file, fileType)
	case models.Kernel:
		return kernelSegmentsCounter(file, gameVersion)
	default:
		return 0, fmt.Errorf("cannot count segments for file type: %v", fileType)
	}
}

func dialogsSegmentsCounter(dialogFile string, dialogType models.NodeType) (int, error) {
	encoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextDlgEncoding(dialogType)
	defer encoding.Dispose()

	args := createSegmentsCountArgs(dialogFile)

	executable, err := encoding.GetDlgHandler().GetDlgHandlerApp()
	if err != nil {
		return 0, err
	}

	return components.GetDialogSegmentsCount(executable, args)
}

func kernelSegmentsCounter(kernelFile string, gameVersion models.GameVersion) (int, error) {
	encoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextKrnlEncoding()
	defer encoding.Dispose()

	args := createSegmentsCountArgs(kernelFile)

	executable, err := encoding.GetKrnlHandler().GetKernelTextHandler(gameVersion)
	if err != nil {
		return 0, err
	}

	return components.GetKernelSegmentsCount(executable, args)
}

func createSegmentsCountArgs(file string) []string {
	return []string{"-p", file}
}
