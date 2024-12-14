package locations

type (
	ExtractLocationInfo struct {
		ExtractLocation ExtractLocation `json:"extract_location"`
	}
	IExtractLocationInfo interface {
		Get() *ExtractLocation
		Set(extractLocation ExtractLocation)
	}
)

func NewExtractLocationInfo() ExtractLocationInfo {
	return ExtractLocationInfo{
		ExtractLocation: *NewExtractLocation(),
	}
}

func (e *ExtractLocationInfo) Get() *ExtractLocation {
	return &e.ExtractLocation
}

func (e *ExtractLocationInfo) Set(extractLocation ExtractLocation) {
	e.ExtractLocation = extractLocation
}
