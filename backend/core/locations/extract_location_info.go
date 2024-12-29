package locations

import internal "ffxresources/backend/core/locations/base"

type (
	ExtractLocationInfo struct {
		ExtractLocation ExtractLocation `json:"extract_location"`
	}
	IExtractLocationInfo interface {
		Get() *ExtractLocation
		Set(extractLocation ExtractLocation)
	}
)

func NewExtractLocationInfo(opts ...internal.LocationBaseOption) ExtractLocationInfo {
	options := internal.ProcessOpts(opts)
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
