package components

import (
	"encoding/binary"
	"fmt"
)

// NameDescriptionTextObject holds keyed strings for name and description
// based on lookup offsets in a binary table.
type NameDescriptionTextObject struct {
	Bytes              []byte
	Name               *LocalizedKeyedStringObject
	Unused0405         *LocalizedKeyedStringObject
	Description        *LocalizedKeyedStringObject
	Unused0C0D         *LocalizedKeyedStringObject

	nameOffset         uint16
	nameKey            uint16
	unused0405Offset   uint16
	unused0405Key      uint16
	descriptionOffset  uint16
	descriptionKey     uint16
	unused0C0DOffset   uint16
	unused0C0DKey      uint16
}

const NameDescriptionTextObjectLength = 0x10

// NewNameDescriptionTextObject initializes and maps bytes and keyed strings
func NewNameDescriptionTextObject(bytes []byte, stringBytes []byte, localization string) *NameDescriptionTextObject {
	n := &NameDescriptionTextObject{
		Bytes:       bytes,
		Name:        NewLocalizedKeyedStringObject(),
		Unused0405:  NewLocalizedKeyedStringObject(),
		Description: NewLocalizedKeyedStringObject(),
		Unused0C0D:  NewLocalizedKeyedStringObject(),
	}
	n.mapBytes()
	n.mapStrings(stringBytes, localization)
	return n
}

func (n *NameDescriptionTextObject) mapBytes() {
	// read two-byte little endian values
	n.nameOffset = binary.LittleEndian.Uint16(n.Bytes[0x00:0x02])
	n.nameKey = binary.LittleEndian.Uint16(n.Bytes[0x02:0x04])
	n.unused0405Offset = binary.LittleEndian.Uint16(n.Bytes[0x04:0x06])
	n.unused0405Key = binary.LittleEndian.Uint16(n.Bytes[0x06:0x08])
	n.descriptionOffset = binary.LittleEndian.Uint16(n.Bytes[0x08:0x0A])
	n.descriptionKey = binary.LittleEndian.Uint16(n.Bytes[0x0A:0x0C])
	n.unused0C0DOffset = binary.LittleEndian.Uint16(n.Bytes[0x0C:0x0E])
	n.unused0C0DKey = binary.LittleEndian.Uint16(n.Bytes[0x0E:0x10])
}

func (n *NameDescriptionTextObject) mapStrings(table []byte, localization string) {
	n.Name.ReadAndSetLocalizedContent(localization, table, int(n.nameOffset), int(n.nameKey))
	n.Unused0405.ReadAndSetLocalizedContent(localization, table, int(n.unused0405Offset), int(n.unused0405Key))
	n.Description.ReadAndSetLocalizedContent(localization, table, int(n.descriptionOffset), int(n.descriptionKey))
	n.Unused0C0D.ReadAndSetLocalizedContent(localization, table, int(n.unused0C0DOffset), int(n.unused0C0DKey))
}

// ToBytes serializes the header offsets back to a byte slice
func (n *NameDescriptionTextObject) ToBytes(localization string) []byte {
	out := make([]byte, NameDescriptionTextObjectLength)
	binary.LittleEndian.PutUint32(out[0x00:0x04], n.Name.GetLocalizedContent(localization).ToHeaderBytes())
	binary.LittleEndian.PutUint32(out[0x04:0x08], n.Unused0405.GetLocalizedContent(localization).ToHeaderBytes())
	binary.LittleEndian.PutUint32(out[0x08:0x0C], n.Description.GetLocalizedContent(localization).ToHeaderBytes())
	binary.LittleEndian.PutUint32(out[0x0C:0x10], n.Unused0C0D.GetLocalizedContent(localization).ToHeaderBytes())
	return out
}

// GetName implements Nameable
func (n *NameDescriptionTextObject) GetName(localization string) string {
	return n.Name.GetLocalizedString(localization)
}

// SetLocalizations copies keyed strings from another object
func (n *NameDescriptionTextObject) SetLocalizations(other *NameDescriptionTextObject) {
	other.Name.CopyInto(n.Name)
	other.Unused0405.CopyInto(n.Unused0405)
	other.Description.CopyInto(n.Description)
	other.Unused0C0D.CopyInto(n.Unused0C0D)
}

// StreamKeyedStrings iterates all keyed strings for a localization
func (n *NameDescriptionTextObject) StreamKeyedStrings(localization string) []*KeyedString {
	return []*KeyedString{
		n.Name.GetLocalizedContent(localization),
		n.Unused0405.GetLocalizedContent(localization),
		n.Description.GetLocalizedContent(localization),
		n.Unused0C0D.GetLocalizedContent(localization),
	}
}

// GetKeyedString retrieves a specific keyed string by title
func (n *NameDescriptionTextObject) GetKeyedString(title string) *LocalizedKeyedStringObject {
	switch title {
	case "name":
		return n.Name
	case "description":
		return n.Description
	default:
		return nil
	}
}

func (n *NameDescriptionTextObject) String() string {
	descStr := ""
	if n.descriptionOffset > 0 {
		descStr = n.Description.GetDefaultContent().String()
	}
	return fmt.Sprintf("%-20s - %s", n.GetName(""), descStr)
}
