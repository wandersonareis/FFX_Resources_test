package interactions

type FfxGamePartNumber int

const (
	Ffx FfxGamePartNumber = iota + 1
	Ffx2
)

type FfxGamePart struct {
	partNumber FfxGamePartNumber
}

func NewFfxGamePart() *FfxGamePart {

	return &FfxGamePart{
		partNumber: Ffx,
	}
}

func (f *FfxGamePart) GetGamePart() FfxGamePartNumber {
	return f.partNumber
}

func (f *FfxGamePart) GetGamePartNumber() int {
	return int(f.partNumber)
}

func (f *FfxGamePart) SetGamePart(partName FfxGamePartNumber) {
	f.partNumber = partName
}

func (f *FfxGamePart) SetGamePartNumber(partNumber int) {
	if partNumber < 1 {
		partNumber = 1
	}

	if partNumber > 2 {
		partNumber = 2
	}
	
	f.partNumber = FfxGamePartNumber(partNumber)
}