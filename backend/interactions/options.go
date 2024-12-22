package interactions

import (
	"ffxresources/backend/core"
	"sync"
)

type IDcpAndLockitOptions interface {
	GetDcpFileOptions() DcpFileOptions
	GetLockitFileOptions() LockitFileOptions
}

type LockitFileOptions struct {
	NameBase        string
	LineBreaksCount int
	PartsLength     int
	PartsSizes      []int
}

type DcpFileOptions struct {
	NameBase    string
	PartsLength int
}
type DcpAndLockitOptions struct {
	*core.FFXGameVersion

	DcpFile    DcpFileOptions
	LockitFile LockitFileOptions
}

var (
	gameOptionsManager  *DcpAndLockitOptions
	initGamePartOptions sync.Once
)

func newDcpAndLockitOptions(gamePart *core.FFXGameVersion) IDcpAndLockitOptions {
	initGamePartOptions.Do(func() {
		gameOptionsManager = &DcpAndLockitOptions{
			FFXGameVersion: gamePart,
		}
	})
	return gameOptionsManager
}

func (g *DcpAndLockitOptions) getDcpOrLockitOptions() DcpAndLockitOptions {
	switch g.FFXGameVersion.GetGameVersion() {
	case core.FFX:
		return ffxOptions()
	case core.FFX2:
		return ffx2Options()
	}

	return DcpAndLockitOptions{}
}

func (g *DcpAndLockitOptions) GetDcpFileOptions() DcpFileOptions {
	return g.getDcpOrLockitOptions().DcpFile
}

func (g *DcpAndLockitOptions) GetLockitFileOptions() LockitFileOptions {
	return g.getDcpOrLockitOptions().LockitFile
}

func ffxOptions() DcpAndLockitOptions {
	return DcpAndLockitOptions{
		DcpFile: DcpFileOptions{
			NameBase:    "macrodic",
			PartsLength: 5,
		},
	}
}

func ffx2Options() DcpAndLockitOptions {
	lockitPartsSizes := []int{80, 88, 90, 93, 94, 95, 102, 1223, 1224, 1230, 1232, 1233, 1240, 1241, 1502, 1534}

	return DcpAndLockitOptions{
		DcpFile: DcpFileOptions{
			NameBase:    "macrodic",
			PartsLength: 7,
		},
		LockitFile: LockitFileOptions{
			NameBase:        "loc_kit_ps3",
			LineBreaksCount: 1696,
			PartsLength:     len(lockitPartsSizes) + 1,
			PartsSizes:      lockitPartsSizes,
		},
	}
}
