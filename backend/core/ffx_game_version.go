package core

type IFfxGameVersion interface {
	GetGameVersion() GameVersion
	GetGameVersionNumber() int
	SetGameVersion(GameVersion)
	SetGameVersionNumber(int)
}

type GameVersion int

const (
	FFX GameVersion = iota + 1
	FFX2
)

type FFXGameVersion struct {
	gameVersion GameVersion
}

func NewFFXGameVersion() *FFXGameVersion {
	return &FFXGameVersion{
		gameVersion: FFX,
	}
}

func (f *FFXGameVersion) GetGameVersion() GameVersion {
	return f.gameVersion
}

func (f *FFXGameVersion) GetGameVersionNumber() int {
	return int(f.gameVersion)
}

func (f *FFXGameVersion) SetGameVersion(partName GameVersion) {
	f.gameVersion = partName
}

func (f *FFXGameVersion) SetGameVersionNumber(partNumber int) {
	if partNumber < 1 {
		partNumber = 1
	}

	if partNumber > 2 {
		partNumber = 2
	}

	f.gameVersion = GameVersion(partNumber)
}
