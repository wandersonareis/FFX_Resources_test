package formats

import (
	"ffxresources/backend/common"
	"ffxresources/backend/interactions"
	"ffxresources/backend/lib"
	"fmt"
	"sync"
)

type LockitFile struct {
	dataInfo *interactions.GameDataInfo
	Parts    *[]LockitFileParts
}

var ffxLockitSizes = []int{}
var ffx2LockitSizes = []int{80, 88, 90, 93, 94, 95, 102, 1223, 1224, 1230, 1232, 1233, 1240, 1241, 1502, 1534}

func NewLockitFile(dataInfo *interactions.GameDataInfo) *LockitFile {
	parts := make([]LockitFileParts, 0, 17)

	gameFilesPath := interactions.NewInteraction().GameLocation.TargetDirectory

	relative := common.GetDifferencePath(dataInfo.GameData.AbsolutePath, gameFilesPath)
	dataInfo.GameData.RelativePath = relative

	dataInfo.ExtractLocation.GenerateTargetOutput(NewTxtFormatter(), dataInfo)
	dataInfo.TranslateLocation.GenerateTargetOutput(NewTxtFormatter(), dataInfo)
	dataInfo.ImportLocation.GenerateTargetOutput(NewTxtFormatter(), dataInfo)

	err := findLockitParts(&parts, dataInfo.ExtractLocation.TargetPath, common.LOCKIT_FILE_PARTS_PATTERN)
	if err != nil {
		lib.NotifyError(err)
		return nil
	}

	return &LockitFile{
		dataInfo: dataInfo,
		Parts:    &parts,
	}
}

func (l *LockitFile) GetFileInfo() *interactions.GameDataInfo {
	return l.dataInfo
}

func (l *LockitFile) Extract() {
	if len(*l.Parts) != len(getLockitFileSizes()) {
		err := ffx2Xplitter(l.dataInfo)
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

func (l *LockitFile) Compress() {
	sizes := getLockitFileSizes()
	translatedParts := make([]LockitFileParts, 0, len(sizes))

	err := findLockitParts(&translatedParts, l.dataInfo.TranslateLocation.TargetPath, common.LOCKIT_TXT_PARTS_PATTERN)
	if err != nil {
		lib.NotifyError(err)
		return
	}

	if len(translatedParts) != len(sizes)+1 {
		lib.NotifyError(fmt.Errorf("invalid number of translated parts: %d expected: %d", len(translatedParts), len(sizes)))
		return
	}
	/* var wg sync.WaitGroup

	for index, part := range *l.Parts {
		wg.Add(1)

		go func(index int, compressor *LockitFileParts) {
			defer wg.Done()

			if index == 0 || index%2 != 0 {
				compressor.Compress(FfxEnc)
			}

			if index%2 == 0 {
				compressor.Compress(LocEnc)
			}
		}(index, &part)
	}

	wg.Wait() */
	for index, part := range *l.Parts {
		if !part.dataInfo.TranslateLocation.TargetFileExists() {
			continue
		}

		compressor := &part

		if index == 0 || index % 2 != 0 {
			compressor.Compress(FfxEnc)
		}

		if index != 0 && index % 2 == 0 {
			compressor.Compress(LocEnc)
		}
	}

	err = ffx2LockitJoiner(l.dataInfo)
	if err != nil {
		lib.NotifyError(err)
		return
	}
}

func findLockitParts(parts *[]LockitFileParts, targetPath, pattern string) error {
	fileParts := make([]string, 0, 16)

	common.EnsurePathExists(targetPath)

	err := common.EnumerateFilesByPattern(&fileParts, targetPath, pattern)
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

func ffx2Xplitter(dataInfo *interactions.GameDataInfo) error {
	handler := newLockitFileXplit(dataInfo)
	common.EnsurePathExists(dataInfo.ExtractLocation.TargetPath)

	err := handler.XplitFile(ffx2LockitSizes, common.LOCKIT_NAME_BASE, dataInfo.ExtractLocation.TargetPath)
	if err != nil {
		return err
	}

	return nil
}

func ffx2LockitJoiner(dataInfo *interactions.GameDataInfo) error {
	joiner := newLockitFileJoin(dataInfo)

	common.EnsurePathExists(dataInfo.TranslateLocation.TargetPath)

	sizes := getLockitFileSizes()
	err := joiner.JoinFile(&sizes)
	if err != nil {
		return err
	}

	return nil
}

func getLockitFileSizes() []int {
	gamePart := interactions.NewInteraction().GamePart.GetGamePart()

	if gamePart == interactions.Ffx {
		return ffxLockitSizes
	}

	if gamePart == interactions.Ffx2 {
		return ffx2LockitSizes
	}

	return nil
}
