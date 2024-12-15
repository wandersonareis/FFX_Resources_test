package locations

import "ffxresources/backend/bases"

type (
	ExtractLocationInfo struct {
		ExtractLocation ExtractLocation `json:"extract_location"`
	}
	IExtractLocationInfo interface {
		Get() *ExtractLocation
		Set(extractLocation ExtractLocation)
	}
)

func NewExtractLocationInfo(opts ...bases.LocationBaseOption) ExtractLocationInfo {
	options := bases.ProcessOpts(opts)
	return ExtractLocationInfo{
		ExtractLocation: *NewExtractLocation(options),
	}
}

func (e *ExtractLocationInfo) Get() *ExtractLocation {
	return &e.ExtractLocation
}

func (e *ExtractLocationInfo) Set(extractLocation ExtractLocation) {
	e.ExtractLocation = extractLocation
}
