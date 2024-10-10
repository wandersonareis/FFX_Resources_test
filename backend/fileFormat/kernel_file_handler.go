package fileFormat

import "ffxresources/backend/lib"

func getKernelFileHandler() (string, error) {
	handlerPath, err := lib.GetKernelHandler()
	if err != nil {
		return "", err
	}

	return handlerPath, nil
}
