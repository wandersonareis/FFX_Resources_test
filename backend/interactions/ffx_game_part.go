package interactions

type GamePart int

const (
	Ffx GamePart = iota + 1
	Ffx2
)

type FfxGamePart struct {
	gamePart GamePart
}

func NewFfxGamePart() *FfxGamePart {

	return &FfxGamePart{
		gamePart: Ffx,
	}
}

func (f *FfxGamePart) GetGamePart() GamePart {
	return f.gamePart
}

func (f *FfxGamePart) GetGamePartNumber() int {
	return int(f.gamePart)
}

func (f *FfxGamePart) SetGamePart(partName GamePart) {
	f.gamePart = partName
}

func (f *FfxGamePart) SetGamePartNumber(partNumber int) {
	if partNumber < 1 {
		partNumber = 1
	}

	if partNumber > 2 {
		partNumber = 2
	}

	f.gamePart = GamePart(partNumber)
}
