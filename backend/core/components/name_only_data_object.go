package components

type NameOnlyDataObject struct {
	*NameOnlyTextObject
	headerParameters []byte
	HeaderLength     int
}

const NameOnlyDataObjectLength int = 0x10

func NewNameOnlyDataObject(data []byte, stringBytes []byte, headerLength int, languageCode string) *NameOnlyDataObject {
	headerStringData := data[:NameOnlyTextObjectLength]
	n := &NameOnlyDataObject{
		NameOnlyTextObject: NewNameOnlyTextObject(headerStringData, stringBytes, headerLength, languageCode),
		HeaderLength:       headerLength,
	}
	if headerLength > NameOnlyDataObjectLength {
		n.headerParameters = data[NameOnlyTextObjectLength:headerLength]
	}
	return n
}

func (n *NameOnlyDataObject) ToBytes(localization string) []byte {
	result := make([]byte, n.HeaderLength)

	header := n.NameOnlyTextObject.ToBytes(localization)
	copy(result, header)

	if len(n.headerParameters) > 0 && NameOnlyDataObjectLength + len(n.headerParameters) <= n.HeaderLength {
        copy(result[NameOnlyDataObjectLength:], n.headerParameters)
    }

	return result
}

func (n *NameOnlyDataObject) ToList(filename string, languageCode string) IList[*NameOnlyDataObject] {
	creator := func(data []byte, stringBytes []byte, headerLength int, loc string) *NameOnlyDataObject {
		return NewNameOnlyDataObject(data, stringBytes, headerLength, loc)
	}
	return ReadDataList(filename, languageCode, creator)
}

func (n *NameOnlyDataObject) GetNameOnlyTextObject() *NameOnlyTextObject {
	return n.NameOnlyTextObject
}

func (n *NameOnlyDataObject) SetLocalizations(other LocalizationSetter) {
	if otherName, ok := other.(*NameOnlyDataObject); ok {
		n.GetNameOnlyTextObject().SetLocalizations(otherName.GetNameOnlyTextObject())
	}
}

func (n *NameOnlyDataObject) GetLocalizedKeyedStrings(localization string) []*KeyedString {
	return n.NameOnlyTextObject.GetLocalizedKeyedStrings(localization)
}
