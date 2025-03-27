package core

import "ffxresources/backend/models"

type IFfxGameVersion interface {
	GetGameVersion() models.GameVersion
	GetGameVersionNumber() int
	SetGameVersion(models.GameVersion)
	SetGameVersionNumber(int)
}

type FFXGameVersion struct {
	gameVersion models.GameVersion
}

// TODO: Implement the methods of the IFfxGameVersion interface
func NewFFXGameVersion() *FFXGameVersion {
	return &FFXGameVersion{
		gameVersion: models.FFX,
	}
}

func (f *FFXGameVersion) GetGameVersion() models.GameVersion {
	return f.gameVersion
}

func (f *FFXGameVersion) GetGameVersionNumber() int {
	return int(f.gameVersion)
}

func (f *FFXGameVersion) SetGameVersion(partName models.GameVersion) {
	f.gameVersion = partName
}

func (f *FFXGameVersion) SetGameVersionNumber(partNumber int) {
	if partNumber < 1 {
		partNumber = 1
	}

	if partNumber > 2 {
		partNumber = 2
	}

	f.gameVersion = models.GameVersion(partNumber)
}
