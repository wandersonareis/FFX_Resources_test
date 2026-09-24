package models

import (
	"ffxresources/backend/common"
)

// EventFileInfo descreve onde o binário de um evento vive para uma versão.
// O fluxo de export/import consome apenas EventID, Version e LocalizationPattern.
type EventFileInfo struct {
	EventID             string             `json:"event_id"`
	Version             common.GameVersion `json:"version"`
	LocalizationPattern string             `json:"localization_pattern"`
}

func NewEventFileInfo(eventID string, version common.GameVersion) *EventFileInfo {
	info := &EventFileInfo{
		EventID: eventID,
		Version: version,
	}
	info.populate()
	return info
}

func (info *EventFileInfo) populate() {
	if len(info.EventID) < 2 {
		return
	}

	info.LocalizationPattern = "event/obj_ps3/" + info.EventID[:2] + "/" + info.EventID + "/" + info.EventID + ".bin"
}
