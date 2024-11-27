package tags

import "fmt"

type FFXTextTagUnknown struct{}

func NewTextTagUnknown() *FFXTextTagUnknown {
	return &FFXTextTagUnknown{}
}

func (u *FFXTextTagUnknown) FFXUnknownBytesCodePage() []string {
	unknownUBytesList := u.getUnknownUBytesList()
	unknownXBytesList := u.getnknownXBytesList()

	unknownList := make([]string, 0, len(*unknownUBytesList)+len(*unknownXBytesList))

	u.generateUnknownUCodePage(&unknownList)
	u.generateUnknownXCodePage(&unknownList)

	return unknownList
}

func (u *FFXTextTagUnknown) generateUnknownUCodePage(codePage *[]string) {
	unknownBytesList := u.getUnknownUBytesList()

	for _, unknownByte := range *unknownBytesList {
		*codePage = append(*codePage, u.generateUnknownUCode(unknownByte))
	}
}

func (u *FFXTextTagUnknown) generateUnknownUCode(unknownByte byte) string {
	return fmt.Sprintf("\\x%02X={u%02X}", unknownByte, unknownByte)
}

func (u *FFXTextTagUnknown) generateUnknownXCodePage(codePage *[]string) {
	unknownBytesList := u.getnknownXBytesList()

	for _, unknownByte := range *unknownBytesList {
		*codePage = append(*codePage, u.generateUnknownXCode(unknownByte))
	}
}

func (u *FFXTextTagUnknown) generateUnknownXCode(unknownByte byte) string {
	return fmt.Sprintf("\\x%02X={x%02X}", unknownByte, unknownByte)
}

func (u *FFXTextTagUnknown) getUnknownUBytesList() *[]byte {
	unknowBytes := []byte{
		0x02, 0x04, 0x05, 0x06, 0x08, 0x11, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F, 0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2A, 0x2B, 0x2C, 0x2D, 0x2E, 0x2F,
	}

	return &unknowBytes
}

func (u *FFXTextTagUnknown) getnknownXBytesList() *[]byte {
	unknownXBytes := []byte{
		0x8A, 0x8C, 0x93, 0x97, 0xA0, 0xA2, 0xB5, 0xD7, 0xDA, 0xDE, 0xEA, 0xEB, 0xEC, 0xED, 0xEE, 0xEF, 0xF0, 0xF1, 0xF2, 0xF3, 0xF4, 0xF5, 0xF6, 0xF7, 0xF8, 0xF9, 0xFA, 0xFB, 0xFC, 0xFD, 0xFE, 0xFF,
	}

	return &unknownXBytes
}
