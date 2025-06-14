package components

import (
	"bytes"
	"encoding/binary"
)

const KeyItemDataObjectLength = 0x14

type (
	keyItemExtraData struct {
		IsAlBhedPrimer uint8
		AlwaysZero     uint8
		UnknownByte12  uint8
		Ordering       uint8
	}

	KeyItemDataObject struct {
		*DataObjectBase[*KeyItemDataObject]
		bytes          []byte
		IsAlBhedPrimer byte
		AlwaysZero     byte
		UnknownByte12  byte
		Ordering       byte
	}
)

func NewKeyItemDataObject(bytes, stringBytes []byte, localization string) *KeyItemDataObject {
	obj := &KeyItemDataObject{
		DataObjectBase: NewDataObjectBase[*KeyItemDataObject](bytes, stringBytes, localization),
		bytes:          bytes,
	}
	obj.mapBytes()
	return obj
}

func (k *KeyItemDataObject) mapBytes() {
	if len(k.bytes) >= KeyItemDataObjectLength {
		k.IsAlBhedPrimer = k.bytes[0x10]
		k.AlwaysZero = k.bytes[0x11]
		k.UnknownByte12 = k.bytes[0x12]
		k.Ordering = k.bytes[0x13]
	}
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
