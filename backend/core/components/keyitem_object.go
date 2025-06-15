package components

import (
	"bytes"
	"encoding/binary"
)

const KeyItemDataObjectLength int = 0x14

type (
	keyItemExtraData struct {
		IsAlBhedPrimer uint8
		AlwaysZero     uint8
		UnknownByte12  uint8
		Ordering       uint8
	}

	KeyItemDataObject struct {
		*DataObjectBase[*KeyItemDataObject]
		keyItemExtraData
		bytes []byte
	}
)

func NewKeyItemDataObject(bytes, stringBytes []byte, localization string) *KeyItemDataObject {
	obj := &KeyItemDataObject{
		DataObjectBase: NewDataObjectBase[*KeyItemDataObject](bytes, stringBytes, localization),
		keyItemExtraData: keyItemExtraData{},
		bytes:          bytes,
	}
	obj.mapBytes()
	return obj
}

func (k *KeyItemDataObject) mapBytes() {
	if len(k.bytes) < KeyItemDataObjectLength {
		panic("KeyItemDataObject bytes length is less than expected")
	}
	extra := keyItemExtraData{}
	if err := binary.Read(bytes.NewReader(k.bytes[0x10:]), binary.LittleEndian, &extra); err != nil {
		panic("Failed to read KeyItemDataObject extra data: " + err.Error())
	}
	k.IsAlBhedPrimer = extra.IsAlBhedPrimer
	k.AlwaysZero = extra.AlwaysZero
	k.UnknownByte12 = extra.UnknownByte12
	k.Ordering = extra.Ordering
}

func (k *KeyItemDataObject) ToBytes(localization string) []byte {
	buf := new(bytes.Buffer)
	header := k.NameDescriptionTextObject.ToBytes(localization)
	buf.Write(header)

	extra := keyItemExtraData{
		IsAlBhedPrimer: k.IsAlBhedPrimer,
		AlwaysZero:     k.AlwaysZero,
		UnknownByte12:  k.UnknownByte12,
		Ordering:       k.Ordering,
	}

	binary.Write(buf, binary.LittleEndian, extra)

	return buf.Bytes()
}

func (k *KeyItemDataObject) SetLocalizations(other LocalizationSetter) {
	if otherKeyItem, ok := other.(*KeyItemDataObject); ok {
		k.NameDescriptionTextObject.SetLocalizations(otherKeyItem.NameDescriptionTextObject)
	}
}
