package interactions

type GamePartOptions struct {
	*FfxGamePart

	DcpPartsLength    int
	LockitPartsLength int
	LockitPartsSizes  []int
}

var gamePartOptionsInstance *GamePartOptions

func NewGamePartOptions(gamePart *FfxGamePart) *GamePartOptions {
	if gamePartOptionsInstance == nil {
		gamePartOptionsInstance = &GamePartOptions{
			FfxGamePart: gamePart,
		}
	}

	return gamePartOptionsInstance
}

func (g *GamePartOptions) GetGamePartOptions() *GamePartOptions {
	switch g.FfxGamePart.GetGamePart() {
	case Ffx:
		return ffxOptions()
	case Ffx2:
		return ffx2Options()
	}

	return nil
}

func ffxOptions() *GamePartOptions {
	return &GamePartOptions{
		DcpPartsLength:    5,
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
