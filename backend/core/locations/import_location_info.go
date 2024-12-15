package locations

import "ffxresources/backend/bases"

type IImportLocationInfo interface {
	Get() *ImportLocation
	Set(importLocation ImportLocation)
}

type ImportLocationInfo struct {
	ImportLocation ImportLocation `json:"import_location"`
}

func NewImportLocationInfo(opts ...bases.LocationBaseOption) ImportLocationInfo {
	options := bases.ProcessOpts(opts)
	return ImportLocationInfo{
		ImportLocation: *NewImportLocation(options),
	}
}

func (i *ImportLocationInfo) Get() *ImportLocation {
	return &i.ImportLocation
}

func (i *ImportLocationInfo) Set(importLocation ImportLocation) {
	i.ImportLocation = importLocation
}
