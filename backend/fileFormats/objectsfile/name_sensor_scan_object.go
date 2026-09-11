package objectsfile

import (
	"bytes"
	"fmt"

	"ffxresources/backend/common"
	"ffxresources/backend/datastore"
	"ffxresources/backend/models"
)

type NameSensorScanTextObject struct {
	Bytes                 []byte
	Name                  datastore.IGlobalLocalizedKeyedStringObject
	SensorText            datastore.IGlobalLocalizedKeyedStringObject
	SimplifiedSensorText  datastore.IGlobalLocalizedKeyedStringObject
	ScanText              datastore.IGlobalLocalizedKeyedStringObject
	SimplifiedScanText    datastore.IGlobalLocalizedKeyedStringObject
	HeaderLength          int
	Version               int
}

func NewNameSensorScanTextObject(
	bytes []byte,
	stringBytes []byte,
	headerLength int,
	languageCode string,
	version int,
) (*NameSensorScanTextObject, error) {
	if len(bytes) < headerLength {
		return nil, fmt.Errorf("insufficient data to create NameSensorScanTextObject: have %d bytes, need at least %d", len(bytes), headerLength)
	}

	p := &NameSensorScanTextObject{
		Bytes:                bytes,
		Name:                 NewLocalizedKeyedStringObject(),
		SensorText:           NewLocalizedKeyedStringObject(),
		SimplifiedSensorText: NewLocalizedKeyedStringObject(),
		ScanText:             NewLocalizedKeyedStringObject(),
		SimplifiedScanText:   NewLocalizedKeyedStringObject(),
		HeaderLength:         headerLength,
		Version:              version,
	}

	if err := p.mapBytes(stringBytes, languageCode, version); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *NameSensorScanTextObject) mapBytes(
	stringBytes []byte,
	languageCode string,
	version int,
) error {
	r := bytes.NewReader(p.Bytes)
	return readStringSegments(r, stringBytes, languageCode, version,
		p.Name,
		p.SensorText,
		p.SimplifiedSensorText,
		p.ScanText,
		p.SimplifiedScanText,
	)
}

func (p *NameSensorScanTextObject) ToBytes(languageCode string) ([]byte, error) {
	data := make([]byte, len(p.Bytes))
	copy(data, p.Bytes)

	if nameContent := p.Name.GetLocalizedContent(languageCode); nameContent != nil {
		if err := models.WriteSegmentAt(data, nameSegmentDefaultPosition, getSegment(nameContent)); err != nil {
			return nil, err
		}
	}

	if sensorContent := p.SensorText.GetLocalizedContent(languageCode); sensorContent != nil {
		if err := models.WriteSegmentAt(data, sensorTextSegmentDefaultPosition, getSegment(sensorContent)); err != nil {
			return nil, err
		}
	}

	if simplifiedSensorContent := p.SimplifiedSensorText.GetLocalizedContent(languageCode); simplifiedSensorContent != nil {
		if err := models.WriteSegmentAt(data, simplifiedSensorTextSegmentPosition, getSegment(simplifiedSensorContent)); err != nil {
			return nil, err
		}
	}

	if scanContent := p.ScanText.GetLocalizedContent(languageCode); scanContent != nil {
		if err := models.WriteSegmentAt(data, scanTextSegmentDefaultPosition, getSegment(scanContent)); err != nil {
			return nil, err
		}
	}

	if simplifiedScanContent := p.SimplifiedScanText.GetLocalizedContent(languageCode); simplifiedScanContent != nil {
		if err := models.WriteSegmentAt(data, simplifiedScanTextSegmentPosition, getSegment(simplifiedScanContent)); err != nil {
			return nil, err
		}
	}

	return data, nil
}

func (p *NameSensorScanTextObject) GetName(languageCode string) string {
	return p.Name.GetLocalizedString(languageCode)
}

func (p *NameSensorScanTextObject) GetKeyedString(title string) datastore.IGlobalLocalizedKeyedStringObject {
	switch title {
	case "name":
		return p.Name
	case "sensorText":
		return p.SensorText
	case "simplifiedSensorText":
		return p.SimplifiedSensorText
	case "scanText":
		return p.ScanText
	case "simplifiedScanText":
		return p.SimplifiedScanText
	default:
		return nil
	}
}

func (p *NameSensorScanTextObject) GetHeaderLength() int {
	return p.HeaderLength
}

func (p *NameSensorScanTextObject) GetTextObject() datastore.IGlobalLocalizedTextObject {
	return p
}

func (p *NameSensorScanTextObject) SetLocalizations(other datastore.IGlobalLocalizationSetter) {
	if o, ok := other.(*NameSensorScanTextObject); ok {
		o.Name.CopyInto(p.Name)
		o.SensorText.CopyInto(p.SensorText)
		o.SimplifiedSensorText.CopyInto(p.SimplifiedSensorText)
		o.ScanText.CopyInto(p.ScanText)
		o.SimplifiedScanText.CopyInto(p.SimplifiedScanText)
	}
}

func (p *NameSensorScanTextObject) GetLocalizedKeyedStrings(localization string) []datastore.IGlobalKeyedString {
	return []datastore.IGlobalKeyedString{
		p.Name.GetLocalizedContent(localization),
		p.SensorText.GetLocalizedContent(localization),
		p.SimplifiedSensorText.GetLocalizedContent(localization),
		p.ScanText.GetLocalizedContent(localization),
		p.SimplifiedScanText.GetLocalizedContent(localization),
	}
}

func (p *NameSensorScanTextObject) ToString(languageCode string) string {
	nameStr := p.GetName(languageCode)
	sensorStr := ""
	if sensorContent := p.SensorText.GetLocalizedContent(languageCode); sensorContent != nil {
		sensorStr = sensorContent.GetString()
	}
	simplifiedSensorStr := ""
	if simplifiedSensorContent := p.SimplifiedSensorText.GetLocalizedContent(languageCode); simplifiedSensorContent != nil {
		simplifiedSensorStr = simplifiedSensorContent.GetString()
	}
	scanStr := ""
	if scanContent := p.ScanText.GetLocalizedContent(languageCode); scanContent != nil {
		scanStr = scanContent.GetString()
	}
	simplifiedScanStr := ""
	if simplifiedScanContent := p.SimplifiedScanText.GetLocalizedContent(languageCode); simplifiedScanContent != nil {
		simplifiedScanStr = simplifiedScanContent.GetString()
	}
	return fmt.Sprintf("%s - %s - %s - %s - %s", nameStr, sensorStr, simplifiedSensorStr, scanStr, simplifiedScanStr)
}

func (p *NameSensorScanTextObject) String() string {
	return p.ToString(common.DefaultLocalization)
}
