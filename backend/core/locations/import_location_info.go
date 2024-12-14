package locations

type IImportLocationInfo interface {
	Get() *ImportLocation
	Set(importLocation ImportLocation)
}

type ImportLocationInfo struct {
	ImportLocation ImportLocation `json:"import_location"`
}

func NewImportLocationInfo() ImportLocationInfo {
	return ImportLocationInfo{
		ImportLocation: *NewImportLocation(),
	}
}

func (i *ImportLocationInfo) Get() *ImportLocation {
	return &i.ImportLocation
}

func (i *ImportLocationInfo) Set(importLocation ImportLocation) {
	i.ImportLocation = importLocation
}
