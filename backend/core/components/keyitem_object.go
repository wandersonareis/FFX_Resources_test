package components

import (
	"fmt"
)

const KeyItemDataObjectLength = 0x14

type KeyItemDataObject struct {
	*NameDescriptionTextObject
	bytes          []byte
	IsAlBhedPrimer byte
	AlwaysZero     byte
	UnknownByte12  byte
	Ordering       byte
}

// NewKeyItemDataObject constructs from raw bytes, stringBytes, and localization
func NewKeyItemDataObject(bytes, stringBytes []byte, localization string) *KeyItemDataObject {
	obj := &KeyItemDataObject{
		NameDescriptionTextObject: NewNameDescriptionTextObject(bytes, stringBytes, localization),
		bytes:                     bytes,
	}
	obj.mapBytes()
	obj.mapFlags()
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

func (k *KeyItemDataObject) mapFlags() {
	// No flags to map yet
}

// GetName returns the localized name
func (k *KeyItemDataObject) GetName(localization string) string {
	return k.Name.GetLocalizedString(localization)
}

// ToBytes serializes the object back to a byte slice
func (k *KeyItemDataObject) ToBytes(localization string) []byte {
	out := make([]byte, KeyItemDataObjectLength)
	copy(out[0:0x10], k.NameDescriptionTextObject.ToBytes(localization))
	out[0x10] = k.IsAlBhedPrimer
	out[0x11] = k.AlwaysZero
	out[0x12] = k.UnknownByte12
	out[0x13] = k.Ordering
	return out
}

func ifG0(value byte, prefix, postfix string) string {
	if value > 0 {
		return fmt.Sprintf("%s%d%s", prefix, value, postfix)
	}
	return ""
}

func formatUnknownByte(bt byte) string {
	bin := fmt.Sprintf("%08b", bt)
	return fmt.Sprintf("%02X=%03d(%s)", bt, bt, bin)
}
