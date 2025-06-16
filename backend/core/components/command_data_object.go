package components

type CommandDataObject struct {
	*DataObjectBase[*CommandDataObject]
}

func NewCommandDataObject(data []byte, stringBytes []byte, languageCode string) *CommandDataObject {
	base := NewDataObjectBase[*CommandDataObject](data, stringBytes, languageCode)

	return &CommandDataObject{
		DataObjectBase: base,
	}
}

func (c *CommandDataObject) ToList(filename string, languageCode string) []*CommandDataObject {
	creator := func(data []byte, stringBytes []byte, loc string) *CommandDataObject {
		return NewCommandDataObject(data, stringBytes, loc)
	}
	return readDataList(filename, languageCode, creator)
}

func (k *CommandDataObject) GetNameDescriptionTextObject() *NameDescriptionTextObject {
	return k.NameDescriptionTextObject
}

func (c *CommandDataObject) SetLocalizations(other LocalizationSetter) {
	if otherCmd, ok := other.(*CommandDataObject); ok {
		c.NameDescriptionTextObject.SetLocalizations(otherCmd.NameDescriptionTextObject)
	}
}
