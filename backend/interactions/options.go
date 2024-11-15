package interactions

import "ffxresources/backend/core"

type IGamePartOptions interface {
	GetGamePartOptions() *GamePartOptions
}

type GamePartOptions struct {
	*core.FfxGamePart

	DcpPartsLength    int
	LockitPartsLength int
	LockitPartsSizes  []int
}

var gamePartOptionsInstance *GamePartOptions

func NewGamePartOptions(gamePart *core.FfxGamePart) *GamePartOptions {
	if gamePartOptionsInstance == nil {
		gamePartOptionsInstance = &GamePartOptions{
			FfxGamePart: gamePart,
		}
	}

	return gamePartOptionsInstance
}

func (g *GamePartOptions) GetGamePartOptions() *GamePartOptions {
	switch g.FfxGamePart.GetGamePart() {
	case core.FFX:
		return ffxOptions()
	case core.FFX2:
		return ffx2Options()
	}

	return nil
}

func ffxOptions() *GamePartOptions {
	return &GamePartOptions{
		DcpPartsLength: 5,
	}
}

func ffx2Options() *GamePartOptions {
	lockitPartsSizes := []int{80, 88, 90, 93, 94, 95, 102, 1223, 1224, 1230, 1232, 1233, 1240, 1241, 1502, 1534}

	return &GamePartOptions{
		DcpPartsLength:    7,
		LockitPartsLength: len(lockitPartsSizes),
		LockitPartsSizes:  lockitPartsSizes,
	}
}
