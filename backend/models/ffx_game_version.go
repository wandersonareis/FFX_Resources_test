package models

type (
	IGameVersionProvider interface {
		GetGameVersion() GameVersion
		GetGameVersionNumber() int
		SetGameVersionNumber(int)
		GetGameVersionModel() GameVersionModel
		GetGameVersionString() string
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

func NewFFXGameVersionFromModel(model GameVersionModel) *FFXGameVersion {
	return NewFFXGameVersion(model.Number())
}

func (f *FFXGameVersion) GetGameVersion() GameVersion {
	return f.gameVersion
}

func (f *FFXGameVersion) GetGameVersionNumber() int {
	return int(f.gameVersion)
}

func (f *FFXGameVersion) GetGameVersionModel() GameVersionModel {
	return f.gameVersion.Model()
}

func (f *FFXGameVersion) GetGameVersionString() string {
	return string(f.gameVersion.Model())
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

func (f *FFXGameVersion) SetGameVersionModel(model GameVersionModel) {
	f.gameVersion = model.ToGameVersion()
}
