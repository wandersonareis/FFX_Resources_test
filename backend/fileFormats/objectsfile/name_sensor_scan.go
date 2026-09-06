package objectsfile

import (
	"bytes"
	"fmt"

	"ffxresources/backend/common"
	"ffxresources/backend/datastore"
	"ffxresources/backend/models"
)

type NameSensorScan struct {
	Bytes                 []byte
	Name                  datastore.IGlobalLocalizedKeyedStringObject
	SensorText            datastore.IGlobalLocalizedKeyedStringObject
	SimplifiedSensorText  datastore.IGlobalLocalizedKeyedStringObject
	ScanText              datastore.IGlobalLocalizedKeyedStringObject
	SimplifiedScanText    datastore.IGlobalLocalizedKeyedStringObject
	HeaderLength          int
}

func NewNameSensorScan(
	bytes []byte,
	stringBytes []byte,
	headerLength int,
	languageCode string,
) *NameSensorScan {
	if len(bytes) < headerLength {
		common.LogVerbose("Insufficient data to create NameSensorScan!")
		return nil
	}

	p := &NameSensorScan{
		Bytes:                bytes,
		Name:                 NewLocalizedKeyedStringObject(),
		SensorText:           NewLocalizedKeyedStringObject(),
		SimplifiedSensorText: NewLocalizedKeyedStringObject(),
		ScanText:             NewLocalizedKeyedStringObject(),
		SimplifiedScanText:   NewLocalizedKeyedStringObject(),
		HeaderLength:         headerLength,
	}

	p.mapBytes(stringBytes, languageCode)

	return p
}

func (p *NameSensorScan) mapBytes(
	stringBytes []byte,
	languageCode string,
) {
	r := bytes.NewReader(p.Bytes)

	nameSeg, err := models.ReadSegment(r)
	if err != nil {
		common.LogVerbose("Error reading NameSensorScan name: %v", err)
		return
	}

	p.Name.ReadAndSetLocalizedContent(
		languageCode,
		stringBytes,
		nameSeg.Offset,
		nameSeg.Key,
	)

	sensorSeg, err := models.ReadSegment(r)
	if err != nil {
		common.LogVerbose("Error reading NameSensorScan sensorText: %v", err)
		return
	}

	p.SensorText.ReadAndSetLocalizedContent(
		languageCode,
		stringBytes,
		sensorSeg.Offset,
		sensorSeg.Key,
	)

	simplifiedSensorSeg, err := models.ReadSegment(r)
	if err != nil {
		common.LogVerbose("Error reading NameSensorScan simplifiedSensorText: %v", err)
		return
	}

	p.SimplifiedSensorText.ReadAndSetLocalizedContent(
		languageCode,
		stringBytes,
		simplifiedSensorSeg.Offset,
		simplifiedSensorSeg.Key,
	)

	scanSeg, err := models.ReadSegment(r)
	if err != nil {
		common.LogVerbose("Error reading NameSensorScan scanText: %v", err)
		return
	}

	p.ScanText.ReadAndSetLocalizedContent(
		languageCode,
		stringBytes,
		scanSeg.Offset,
		scanSeg.Key,
	)

	simplifiedScanSeg, err := models.ReadSegment(r)
	if err != nil {
		common.LogVerbose("Error reading NameSensorScan simplifiedScanText: %v", err)
		return
	}

	p.SimplifiedScanText.ReadAndSetLocalizedContent(
		languageCode,
		stringBytes,
		simplifiedScanSeg.Offset,
		simplifiedScanSeg.Key,
	)
}

func (p *NameSensorScan) ToBytes(languageCode string) []byte {
	data := make([]byte, len(p.Bytes))
	copy(data, p.Bytes)

	if nameContent := p.Name.GetLocalizedContent(languageCode); nameContent != nil {
		models.WriteSegmentAt(data, nameSegmentDefaultPosition, getSegment(nameContent))
	}

	if sensorContent := p.SensorText.GetLocalizedContent(languageCode); sensorContent != nil {
		models.WriteSegmentAt(data, sensorTextSegmentDefaultPosition, getSegment(sensorContent))
	}

	if simplifiedSensorContent := p.SimplifiedSensorText.GetLocalizedContent(languageCode); simplifiedSensorContent != nil {
		models.WriteSegmentAt(data, simplifiedSensorTextSegmentPosition, getSegment(simplifiedSensorContent))
	}

	if scanContent := p.ScanText.GetLocalizedContent(languageCode); scanContent != nil {
		models.WriteSegmentAt(data, scanTextSegmentDefaultPosition, getSegment(scanContent))
	}

	if simplifiedScanContent := p.SimplifiedScanText.GetLocalizedContent(languageCode); simplifiedScanContent != nil {
		models.WriteSegmentAt(data, simplifiedScanTextSegmentPosition, getSegment(simplifiedScanContent))
	}

	return data
}

func (p *NameSensorScan) GetName(languageCode string) string {
	return p.Name.GetLocalizedString(languageCode)
}

func (p *NameSensorScan) GetKeyedString(title string) datastore.IGlobalLocalizedKeyedStringObject {
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

func (p *NameSensorScan) GetHeaderLength() int {
	return p.HeaderLength
}

func (p *NameSensorScan) GetTextObject() datastore.IGlobalLocalizedTextObject {
	return p
}

func (p *NameSensorScan) SetLocalizations(other datastore.IGlobalLocalizationSetter) {
	if o, ok := other.(*NameSensorScan); ok {
		o.Name.CopyInto(p.Name)
		o.SensorText.CopyInto(p.SensorText)
		o.SimplifiedSensorText.CopyInto(p.SimplifiedSensorText)
		o.ScanText.CopyInto(p.ScanText)
		o.SimplifiedScanText.CopyInto(p.SimplifiedScanText)
	}
}

func (p *NameSensorScan) GetLocalizedKeyedStrings(localization string) []datastore.IGlobalKeyedString {
	return []datastore.IGlobalKeyedString{
		p.Name.GetLocalizedContent(localization),
		p.SensorText.GetLocalizedContent(localization),
		p.SimplifiedSensorText.GetLocalizedContent(localization),
		p.ScanText.GetLocalizedContent(localization),
		p.SimplifiedScanText.GetLocalizedContent(localization),
	}
}

func (p *NameSensorScan) ToString(languageCode string) string {
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

func (p *NameSensorScan) String() string {
	return p.ToString(common.DefaultLocalization)
}
