package models

import (
	"path/filepath"

	"ffxresources/backend/common"
)

type EventFileInfo struct {
	EventID             string             `json:"event_id"`
	Shortened           string             `json:"shortened"`
	MidPath             string             `json:"mid_path"`
	EventFilePath       string             `json:"event_file_path"`
	LocalizationPattern string             `json:"localization_pattern"`
	Version             common.GameVersion `json:"version"`
	DirPattern          string             `json:"dir_pattern"`
	FileName            string             `json:"file_name"`
	Key                 string             `json:"key"`
}

func NewEventFileInfo(eventID string, version common.GameVersion) *EventFileInfo {
	info := &EventFileInfo{
		EventID:  eventID,
		Version:  version,
	}
	info.populate()
	return info
}

func (info *EventFileInfo) populate() {
	if len(info.EventID) < 2 {
		return
	}

	info.Shortened = info.EventID[:2]
	info.MidPath = filepath.ToSlash(filepath.Join(info.Shortened, info.EventID, info.EventID))
	info.LocalizationPattern = "event/obj_ps3/" + info.MidPath + ".bin"
	info.EventFilePath = filepath.ToSlash(filepath.Join(
		common.GetLocalizationRootForVersion(info.Version, common.DefaultLocalization),
		info.LocalizationPattern,
	))
	info.DirPattern = "event/obj_ps3"
	info.FileName = info.EventID + ".bin"
	info.Key = common.VersionPathName(info.Version) + "/" + info.LocalizationPattern
}

func (info *EventFileInfo) LocalizedFilePath(localization string) string {
	return filepath.ToSlash(filepath.Join(
		common.GetLocalizationRootForVersion(info.Version, localization),
		info.LocalizationPattern,
	))
}

func (info *EventFileInfo) SetEventID(eventID string, version common.GameVersion) {
	info.EventID = eventID
	info.Version = version
	info.populate()
}
