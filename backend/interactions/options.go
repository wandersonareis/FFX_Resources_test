package interactions

import "ffxresources/backend/core"

type IGamePartOptions interface {
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
type GamePartOptions struct {
	*core.FfxGamePart

	DcpFile    DcpFileOptions
	LockitFile LockitFileOptions
}

var gamePartOptionsInstance *GamePartOptions

func NewGamePartOptions(gamePart *core.FfxGamePart) IGamePartOptions {
	if gamePartOptionsInstance == nil {
		gamePartOptionsInstance = &GamePartOptions{
			FfxGamePart: gamePart,
		}
	}

	return gamePartOptionsInstance
}

func (g *GamePartOptions) getGamePartOptions() GamePartOptions {
	switch g.FfxGamePart.GetGamePart() {
	case core.FFX:
		return ffxOptions()
	case core.FFX2:
		return ffx2Options()
	}

	return GamePartOptions{}
}

func (g *GamePartOptions) GetDcpFileOptions() DcpFileOptions {
	return g.getGamePartOptions().DcpFile
}

func (g *GamePartOptions) GetLockitFileOptions() LockitFileOptions {
	return g.getGamePartOptions().LockitFile
}

func ffxOptions() GamePartOptions {
	return GamePartOptions{
		DcpFile: DcpFileOptions{
			NameBase:    "macrodic",
			PartsLength: 5,
		},
	}
}

func ffx2Options() GamePartOptions {
	lockitPartsSizes := []int{80, 88, 90, 93, 94, 95, 102, 1223, 1224, 1230, 1232, 1233, 1240, 1241, 1502, 1534}

	return GamePartOptions{
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
