package internal

import (
	"ffxresources/backend/common"
	"ffxresources/backend/formats/lib"
	"ffxresources/backend/interactions"
)

func SegmentFile(parts *[]LockitFileParts) {
	worker := common.NewWorker[LockitFileParts]()

	worker.ParallelForEach(*parts,
		func(index int, part LockitFileParts) {
			if index > 0 && index%2 == 0 {
				part.Extract(LocEnc)
			} else {
				part.Extract(FfxEnc)
			}
		})
}

func EnsurePartsExists(info *interactions.GameDataInfo) error {
	partsSizes := interactions.NewInteraction().GamePartOptions.GetGamePartOptions().LockitPartsSizes
	if err := ffx2Xplitter(info, partsSizes); err != nil {
		return err
	}

	return nil
}

func ffx2Xplitter(dataInfo *interactions.GameDataInfo, sizes []int) error {
	handler := NewLockitFileXplit(dataInfo)

	extractLocation := dataInfo.ExtractLocation

	if err := extractLocation.ProvideTargetPath(); err != nil {
		return err
	}

	if err := handler.XplitFile(sizes, lib.LOCKIT_NAME_BASE, extractLocation.TargetPath); err != nil {
		return err
	}

	return nil
}
