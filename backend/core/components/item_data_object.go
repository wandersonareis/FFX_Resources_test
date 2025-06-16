package components

type ItemDataObject struct {
	*DataObjectBase[*ItemDataObject]
}

func NewItemDataObject(data []byte, stringBytes []byte, localization string) *ItemDataObject {
	base := NewDataObjectBase[*ItemDataObject](data, stringBytes, localization)

	return &ItemDataObject{
		DataObjectBase: base,
	}
}

func (i *ItemDataObject) ToList(filename string, languageCode string) []*ItemDataObject {
	creator := func(data []byte, stringBytes []byte, loc string) *ItemDataObject {
		return NewItemDataObject(data, stringBytes, loc)
	}
	return readDataList(filename, languageCode, creator)
}

func LoadItemsFromFile(filename string, languageCode string) []*ItemDataObject {
	creator := func(data []byte, stringBytes []byte, loc string) *ItemDataObject {
		return NewItemDataObject(data, stringBytes, loc)
	}
	return ReadDataArray(filename, languageCode, creator)
}

func (i *ItemDataObject) SetLocalizations(other LocalizationSetter) {
	if otherItem, ok := other.(*ItemDataObject); ok {
		i.NameDescriptionTextObject.SetLocalizations(otherItem.NameDescriptionTextObject)
	}
}
