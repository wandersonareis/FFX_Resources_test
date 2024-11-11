package internal

import "ffxresources/backend/fileFormats/lib"

func getLockitFileHandler() (string, error) {
	targetFile, err := lib.GetFromResources(lib.LOCKIT_HANDLER_APPLICATION, lib.DEFAULT_APPLICATION_EXTENSION)
	if err != nil {
		return "", err
	}
	
	return targetFile, nil
}

func getLockitFileUtf8BomNormalizer() (string, error) {
	targetFile, err := lib.GetFromResources(lib.UTF8BOM_NORMALIZER_APPLICATION, lib.DEFAULT_APPLICATION_EXTENSION)
	if err != nil {
		return "", err
	}

	return targetFile, nil
}
