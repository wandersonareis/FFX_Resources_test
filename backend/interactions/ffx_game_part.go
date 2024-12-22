package interactions

/* type IFfxGamePart interface {
	FFXGameVersion() FFXGameVersion
	GetGameVersionNumber() int
	SetGameVersion(FFXGameVersion)
	SetGameVersionNumber(int)
}

type FFXGameVersion int

const (
	Ffx FFXGameVersion = iota + 1
	Ffx2
)

type FFXGameVersion struct {
	ffxGameVersion FFXGameVersion
}

func NewFfxGamePart() *FFXGameVersion {

	return &FFXGameVersion{
		ffxGameVersion: Ffx,
	}
}

func (f *FFXGameVersion) FFXGameVersion() FFXGameVersion {
	return f.ffxGameVersion
}

func (f *FFXGameVersion) GetGameVersionNumber() int {
	return int(f.ffxGameVersion)
}

func (f *FFXGameVersion) SetGameVersion(partName FFXGameVersion) {
	f.ffxGameVersion = partName
}

func (f *FFXGameVersion) SetGameVersionNumber(partNumber int) {
	if partNumber < 1 {
		partNumber = 1
	}

	if partNumber > 2 {
		partNumber = 2
	}

	f.ffxGameVersion = FFXGameVersion(partNumber)
} */
