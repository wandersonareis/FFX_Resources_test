package formats

import (
	"ffxresources/backend/common"
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
	"sync"
)

type LockitFile struct {
	dataInfo *interactions.GameDataInfo
	Parts    *[]LockitFileParts
}

func NewLockitFile(dataInfo *interactions.GameDataInfo) *LockitFile {
	parts := &[]LockitFileParts{}

	err := findParts(dataInfo, parts)
	if err != nil {
		lib.NotifyError(err)
		return nil
	}

	dataInfo.ExtractLocation.GenerateTargetOutput(NewTxtFormatter(), dataInfo)
	dataInfo.TranslateLocation.GenerateTargetOutput(NewTxtFormatter(), dataInfo)
	dataInfo.ImportLocation.GenerateTargetOutput(NewTxtFormatter(), dataInfo)

	return &LockitFile{
		dataInfo: dataInfo,
		Parts:    parts,
	}
}

func (l *LockitFile) GetFileInfo() *interactions.GameDataInfo {
	return l.dataInfo
}

func (l *LockitFile) Extract() {
	if len(*l.Parts) == 0 {
		err := ffx2Splitter(l.dataInfo)
		if err != nil {
			lib.NotifyError(err)
			return
		}

		newLockitFile := NewLockitFile(l.dataInfo)
		l.dataInfo = newLockitFile.GetFileInfo()
		l.Parts = newLockitFile.Parts
	}

	var wg sync.WaitGroup

	for i, part := range *l.Parts {
		wg.Add(1)

		go func(index int, extractor *LockitFileParts) {
			defer wg.Done()

			if index == 0 || index%2 != 0 {
				extractor.Extract(FfxEnc)
			}

			if index%2 == 0 {
				extractor.Extract(LocEnc)
			}
		}(i, &part)
	}

	wg.Wait()
}

func (l *LockitFile) Compress() {}

func findParts(dataInfo *interactions.GameDataInfo, parts *[]LockitFileParts) error {
	fileParts := make([]string, 0, 16)

	err := common.EnumerateFilesByPattern(&fileParts, dataInfo.ExtractLocation.TargetPath, common.LOCKIT_FILE_PARTS_PATTERN)
	if err != nil {
		return err
	}

	for _, file := range fileParts {
		info := interactions.NewGameDataInfo(file)
		if info.GameData.Size == 0 {
			continue
		}

		lockitPart := NewLockitFileParts(info)
		if lockitPart == nil {
			continue
		}

		*parts = append(*parts, *lockitPart)
	}

	return nil
}

func ffx2Splitter(dataInfo *interactions.GameDataInfo) error {
	sizes := []int{80, 88, 90, 93, 94, 95, 102, 1223, 1224, 1230, 1232, 1233, 1240, 1241, 1502, 1534} // Exemplo de contagens de segmentos

	handler := newLockitFileHandler(dataInfo)
	common.EnsurePathExists(dataInfo.ExtractLocation.TargetPath)

	err := handler.SplitFile(sizes, common.LOCKIT_NAME_BASE, dataInfo.ExtractLocation.TargetPath)
	if err != nil {
		return err
	}

	return nil
}

/* func processLockitPartsParallel(fileParts *[]LockitFileParts, callback func(index int, handler *LockitFileParts)) error {
	var wg sync.WaitGroup

	errChan := make(chan error, len(*fileParts))

	for index, fileParts := range *fileParts {
		wg.Add(1)

		go func(file LockitFileParts) {
			defer wg.Done()

			callback(index, &file)

		}(fileParts)
	}

	wg.Wait()

	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
} */
