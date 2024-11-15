package internal

import "ffxresources/backend/fileFormats/util"

func getLockitFileHandler() (string, error) {
	targetFile, err := util.GetFromResources(util.LOCKIT_HANDLER_APPLICATION, util.DEFAULT_APPLICATION_EXTENSION)
	if err != nil {
		return "", err
	}

	return targetFile, nil
}

func getLockitFileUtf8BomNormalizer() (string, error) {
	targetFile, err := util.GetFromResources(util.UTF8BOM_NORMALIZER_APPLICATION, util.DEFAULT_APPLICATION_EXTENSION)
	if err != nil {
		return "", err
	}

	return targetFile, nil
}
