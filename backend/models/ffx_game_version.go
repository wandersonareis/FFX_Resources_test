package models

type (
	IGameVersionProvider interface {
		GetGameVersion() GameVersion
		GetGameVersionNumber() int
		SetGameVersionNumber(int)
	}

	FFXGameVersion struct {
		gameVersion GameVersion
	}
)

func NewFFXGameVersion(version int) *FFXGameVersion {
	gameVersion := &FFXGameVersion{}
	gameVersion.SetGameVersionNumber(version)
	return gameVersion
}

func (f *FFXGameVersion) GetGameVersion() GameVersion {
	return f.gameVersion
}

func (f *FFXGameVersion) GetGameVersionNumber() int {
	return int(f.gameVersion)
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
