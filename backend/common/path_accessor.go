package common

import "path/filepath"

func GetEncodingPath(charset string) string {
	gameVersion := GetGameVersionString()
	switch gameVersion {
	case "ffx":
		return filepath.Join("ffx_encoding", "ffxsjistbl_"+charset+".bin")
	case "ffx2":
		return filepath.Join("ffx2_encoding", "ffx2sjistbl_"+charset+".bin")
	default:
		return filepath.Join("ffx_encoding", "ffxsjistbl_"+charset+".bin")
	}
}

func GetPathRoot() string {
	gameVersion := GetGameVersionString()
	return  filepath.Join("ffx_ps2", gameVersion, "master")
}

func GetPathOriginalsRoot() string {
	return filepath.Join(GetPathRoot(), OriginalsFolder)
}

func GetPathOriginalsKernel() string {
	return filepath.Join(GetPathOriginalsRoot(), "battle", "kernel")
}

func GetPathMonsterFolder() string {
	return filepath.Join(GetPathOriginalsRoot(), "battle", "mon")
}

func GetPathOriginalsEncounter() string {
	return filepath.Join(GetPathOriginalsRoot(), "battle", "btl")
}

func GetInternationalEncounterPath() string {
	return filepath.Join(GetPathRoot(), "inpc", "battle", "btl")
}

func GetPathOriginalsEvent() string {
	return filepath.Join(GetPathOriginalsRoot(), "event", "obj")
}

func GetAbmapPath() string {
	return filepath.Join(GetPathOriginalsRoot(), "menu", "abmap")
}

func GetLocalizationRoot(localization string) string {
	return filepath.Join(GetPathRoot(), "new_"+localization+"pc")
}