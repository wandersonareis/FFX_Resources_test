package models

type (
	IFfxGameVersion interface {
		GetGameVersion() GameVersion
		GetGameVersionNumber() int
		SetGameVersion(GameVersion)
		SetGameVersionNumber(int)
	}

	FFXGameVersion struct {
		gameVersion GameVersion
	}
)

// TODO: Implement the methods of the IFfxGameVersion interface
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
