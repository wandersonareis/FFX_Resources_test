package objectsfile

import (
    "bytes"
    "fmt"
    "io"

    "ffxresources/backend/common"
    "ffxresources/backend/datastore"
    "ffxresources/backend/models"
)

type PlateTextObject struct {
    Bytes              []byte
    NameSegment        datastore.IGlobalLocalizedKeyedStringObject
    DescriptionSegment datastore.IGlobalLocalizedKeyedStringObject
    AbilitySegments    [4]datastore.IGlobalLocalizedKeyedStringObject
    AbilityData        []byte
    EffectSegment      datastore.IGlobalLocalizedKeyedStringObject
    unknownBytes       []byte
    HeaderLength       int
}

const PlateTextObjectV2Length = 0x08
const AbilityDataLength = 0x30

func NewPlateTextObject(bytes []byte, stringBytes []byte, headerLength int, languageCode string) *PlateTextObject {
    if len(bytes) < PlateTextObjectV2Length {
        common.LogVerbose("Insufficient data to create PlateTextObject!")
        return nil
    }

    p := &PlateTextObject{
        Bytes:              bytes,
        NameSegment:        NewLocalizedKeyedStringObject(),
        DescriptionSegment: NewLocalizedKeyedStringObject(),
        EffectSegment:      NewLocalizedKeyedStringObject(),
        HeaderLength:       headerLength,
    }
    
    for i := range 4 {
        p.AbilitySegments[i] = NewLocalizedKeyedStringObject()
    }
    
    p.mapBytes(stringBytes, languageCode)
    return p
}

func (p *PlateTextObject) mapBytes(stringBytes []byte, languageCode string) {
    r := bytes.NewReader(p.Bytes)

    nameSeg, err := models.ReadSegment(r)
    if err != nil {
        common.LogVerbose("Error reading PlateTextObject name: %v", err)
        return
    }
    p.NameSegment.ReadAndSetLocalizedContent(languageCode, stringBytes, nameSeg.Offset, nameSeg.Key)

    descSeg, err := models.ReadSegment(r)
    if err != nil {
        common.LogVerbose("Error reading PlateTextObject description: %v", err)
        return
    }
    p.DescriptionSegment.ReadAndSetLocalizedContent(languageCode, stringBytes, descSeg.Offset, descSeg.Key)

    for i := 0; i < 4; i++ {
        abSeg, err := models.ReadSegment(r)
        if err != nil {
            common.LogVerbose("Error reading PlateTextObject ability %d: %v", i+1, err)
            continue
        }
        p.AbilitySegments[i].ReadAndSetLocalizedContent(languageCode, stringBytes, abSeg.Offset, abSeg.Key)
    }

    p.AbilityData = make([]byte, AbilityDataLength)
    if _, err := io.ReadFull(r, p.AbilityData); err != nil {
        common.LogVerbose("Error reading PlateTextObject ability data: %v", err)
        p.AbilityData = nil
    }

    effSeg, err := models.ReadSegment(r)
    if err != nil {
        common.LogVerbose("Error reading PlateTextObject effect: %v", err)
        return
    }
    p.EffectSegment.ReadAndSetLocalizedContent(languageCode, stringBytes, effSeg.Offset, effSeg.Key)

    if r.Len() > 0 {
        p.unknownBytes = make([]byte, r.Len())
        if _, err := io.ReadFull(r, p.unknownBytes); err != nil {
            common.LogVerbose("Error reading PlateTextObject unknown bytes: %v", err)
            p.unknownBytes = nil
        }
    }
}

func (p *PlateTextObject) ToBytes(languageCode string) []byte {
    var buf bytes.Buffer

    models.WriteSegment(&buf, getSegment(p.NameSegment.GetLocalizedContent(languageCode)))
    models.WriteSegment(&buf, getSegment(p.DescriptionSegment.GetLocalizedContent(languageCode)))

    for i := 0; i < 4; i++ {
        models.WriteSegment(&buf, getSegment(p.AbilitySegments[i].GetLocalizedContent(languageCode)))
    }

    if len(p.AbilityData) > 0 {
        buf.Write(p.AbilityData)
    }

    models.WriteSegment(&buf, getSegment(p.EffectSegment.GetLocalizedContent(languageCode)))

    if len(p.unknownBytes) > 0 {
        buf.Write(p.unknownBytes)
    }

    return buf.Bytes()
}

func (p *PlateTextObject) GetName(languageCode string) string {
    return p.NameSegment.GetLocalizedString(languageCode)
}

func (p *PlateTextObject) GetKeyedString(title string) datastore.IGlobalLocalizedKeyedStringObject {
    switch title {
    case "name":
        return p.NameSegment
    case "description":
        return p.DescriptionSegment
    case "ability1":
        return p.AbilitySegments[0]
    case "ability2":
        return p.AbilitySegments[1]
    case "ability3":
        return p.AbilitySegments[2]
    case "ability4":
        return p.AbilitySegments[3]
    case "effect":
        return p.EffectSegment
    default:
        return nil
    }
}

func (p *PlateTextObject) GetHeaderLength() int {
    return p.HeaderLength
}

func (p *PlateTextObject) GetTextObject() datastore.IGlobalLocalizedTextObject {
    return p
}

func (p *PlateTextObject) SetLocalizations(other datastore.IGlobalLocalizationSetter) {
    if o, ok := other.(*PlateTextObject); ok {
        o.NameSegment.CopyInto(p.NameSegment)
        o.DescriptionSegment.CopyInto(p.DescriptionSegment)
        
        for i := 0; i < 4; i++ {
            o.AbilitySegments[i].CopyInto(p.AbilitySegments[i])
        }
        
        o.EffectSegment.CopyInto(p.EffectSegment)
        
        if len(o.AbilityData) > 0 {
            p.AbilityData = make([]byte, len(o.AbilityData))
            copy(p.AbilityData, o.AbilityData)
        }
        if len(o.unknownBytes) > 0 {
            p.unknownBytes = make([]byte, len(o.unknownBytes))
            copy(p.unknownBytes, o.unknownBytes)
        }
    }
}

func (p *PlateTextObject) GetLocalizedKeyedStrings(localization string) []datastore.IGlobalKeyedString {
    return []datastore.IGlobalKeyedString{
        p.NameSegment.GetLocalizedContent(localization),
        p.DescriptionSegment.GetLocalizedContent(localization),
        p.AbilitySegments[0].GetLocalizedContent(localization),
        p.AbilitySegments[1].GetLocalizedContent(localization),
        p.AbilitySegments[2].GetLocalizedContent(localization),
        p.AbilitySegments[3].GetLocalizedContent(localization),
        p.EffectSegment.GetLocalizedContent(localization),
    }
}

func (p *PlateTextObject) ToString(languageCode string) string {
    nameStr := p.GetName(languageCode)
    descStr := ""
    if descContent := p.DescriptionSegment.GetLocalizedContent(languageCode); descContent != nil {
        descStr = descContent.GetString()
    }
    return fmt.Sprintf("%s - %s", nameStr, descStr)
}

func (p *PlateTextObject) String() string {
    return p.ToString(common.DefaultLocalization)
}