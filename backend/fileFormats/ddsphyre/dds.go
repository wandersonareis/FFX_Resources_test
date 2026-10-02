package ddsphyre

import "fmt"

// Flags DDS (idênticas às enums do tool C++).
const (
	ddsFlagCaps        = 0x1
	ddsFlagHeight      = 0x2
	ddsFlagWidth       = 0x4
	ddsFlagPitch       = 0x8
	ddsFlagPixelFormat = 0x1000
	ddsFlagMipmapCount = 0x20000
	ddsFlagLinearSize  = 0x80000

	ddpfAlphaPixels = 0x1
	ddpfAlpha       = 0x2
	ddpfFourCC      = 0x4
	ddpfRGB         = 0x40
	ddpfLuminance   = 0x20000

	ddsCapsComplex = 0x8
	ddsCapsTexture = 0x1000
	ddsCapsMipmap  = 0x400000
)

// FourCC (little-endian como gravados no arquivo).
const (
	fccDXT5 = 0x35545844 // "DXT5"
	fccDXT3 = 0x33545844 // "DXT3"
	fccDXT1 = 0x31545844 // "DXT1"
	fccBC5U = 0x55354342 // "BC5U" (equivale a ATI2)
	fccATI2 = 0x32495441 // "ATI2" (equivale a BC5U)
)

// Formatos suportados pelo prepareDDSHeader.
const (
	FormatDXT1  = "DXT1"
	FormatDXT3  = "DXT3"
	FormatDXT5  = "DXT5"
	FormatBC5   = "BC5"
	FormatARGB8 = "ARGB8"
	FormatA8    = "A8"
	FormatL8    = "L8"
)

// ddsHeader é o _tDDS_HEADER (magia + 124 bytes = 128).
type ddsHeader struct {
	magic    [4]byte
	size     uint32
	flags    uint32
	height   uint32
	width    uint32
	pitch    uint32
	depth    uint32
	mips     uint32
	reserved [11]uint32
	pfSize   uint32
	pfFlags  uint32
	pfFourCC uint32
	pfBits   uint32
	pfR      uint32
	pfG      uint32
	pfB      uint32
	pfA      uint32
	caps     uint32
	caps2    uint32
	caps3    uint32
	caps4    uint32
	res2     uint32
}

// newDDSHeader devolve o header com os valores por omissão da struct C++.
func newDDSHeader() ddsHeader {
	h := ddsHeader{magic: [4]byte{'D', 'D', 'S', ' '}}
	h.size = ddsHeaderBodyLen
	h.flags = ddsFlagCaps | ddsFlagHeight | ddsFlagWidth | ddsFlagPixelFormat
	h.depth = 1
	h.mips = 1
	h.pfSize = 32
	h.pfFlags = ddpfFourCC
	h.pfBits = 32
	h.pfR = 0x00FF0000
	h.pfG = 0x0000FF00
	h.pfB = 0x000000FF
	h.pfA = 0xFF000000
	h.caps = ddsCapsTexture
	return h
}

// bytes serializa o header (layout idêntico ao struct C++, sem padding).
func (h ddsHeader) bytes() []byte {
	b := make([]byte, ddsHeaderLen)
	copy(b[0:4], h.magic[:])
	put32(b, 4, h.size)
	put32(b, 8, h.flags)
	put32(b, 12, h.height)
	put32(b, 16, h.width)
	put32(b, 20, h.pitch)
	put32(b, 24, h.depth)
	put32(b, 28, h.mips)
	for i, v := range h.reserved {
		put32(b, 32+uint64(i)*4, v)
	}
	put32(b, 76, h.pfSize)
	put32(b, 80, h.pfFlags)
	put32(b, 84, h.pfFourCC)
	put32(b, 88, h.pfBits)
	put32(b, 92, h.pfR)
	put32(b, 96, h.pfG)
	put32(b, 100, h.pfB)
	put32(b, 104, h.pfA)
	put32(b, 108, h.caps)
	put32(b, 112, h.caps2)
	put32(b, 116, h.caps3)
	put32(b, 120, h.caps4)
	put32(b, 124, h.res2)
	return b
}

// parseDDSHeader lê um header DDS de um arquivo completo.
func parseDDSHeader(b []byte) (ddsHeader, error) {
	var h ddsHeader
	if len(b) < ddsHeaderLen {
		return h, fmt.Errorf("dds: %d bytes, esperado ao menos %d", len(b), ddsHeaderLen)
	}
	if string(b[0:4]) != "DDS " {
		return h, fmt.Errorf("dds: magia inválida %q (esperado \"DDS \")", string(b[0:4]))
	}
	if size := mustLE32(b, 4); size != ddsHeaderBodyLen {
		return h, fmt.Errorf("dds: dwSize=%d (esperado %d)", size, ddsHeaderBodyLen)
	}
	h.magic = [4]byte{b[0], b[1], b[2], b[3]}
	h.size = mustLE32(b, 4)
	h.flags = mustLE32(b, 8)
	h.height = mustLE32(b, 12)
	h.width = mustLE32(b, 16)
	h.pitch = mustLE32(b, 20)
	h.depth = mustLE32(b, 24)
	h.mips = mustLE32(b, 28)
	for i := range h.reserved {
		h.reserved[i] = mustLE32(b, 32+uint64(i)*4)
	}
	h.pfSize = mustLE32(b, 76)
	h.pfFlags = mustLE32(b, 80)
	h.pfFourCC = mustLE32(b, 84)
	h.pfBits = mustLE32(b, 88)
	h.pfR = mustLE32(b, 92)
	h.pfG = mustLE32(b, 96)
	h.pfB = mustLE32(b, 100)
	h.pfA = mustLE32(b, 104)
	h.caps = mustLE32(b, 108)
	h.caps2 = mustLE32(b, 112)
	h.caps3 = mustLE32(b, 116)
	h.caps4 = mustLE32(b, 120)
	h.res2 = mustLE32(b, 124)
	return h, nil
}

// prepareDDSHeader é o prepareDDSHeader do tool C++: monta o header DDS a
// partir do formato Phyre, dimensões e contagem de mips PHYRE (a DDS conta
// a textura base como mipmap → dwMipMapCount = mipmaps + 1).
func prepareDDSHeader(format string, width, height, mipmaps uint32) ([]byte, error) {
	h := newDDSHeader()
	h.width = width
	h.height = height
	h.mips = mipmaps + 1
	if mipmaps > 0 {
		h.flags |= ddsFlagMipmapCount
		h.caps |= ddsCapsMipmap | ddsCapsComplex
	}

	switch format {
	case FormatDXT5, FormatDXT3, FormatDXT1, FormatBC5:
		h.flags |= ddsFlagLinearSize
		h.pfR, h.pfG, h.pfB, h.pfA = 0, 0, 0, 0
		h.pitch = width * height
		switch format {
		case FormatDXT5:
			h.pfFourCC = fccDXT5
		case FormatDXT3:
			h.pfFourCC = fccDXT3
		case FormatDXT1:
			h.pfFourCC = fccDXT1
			h.pitch = width * height / 2
		default:
			h.pfFourCC = fccBC5U
		}
	case FormatARGB8:
		h.flags |= ddsFlagPitch
		h.pitch = 4 * width
		h.pfFlags = ddpfAlphaPixels | ddpfRGB
	case FormatA8:
		h.flags |= ddsFlagPitch
		h.pitch = width
		h.pfFlags = ddpfAlpha
		h.pfBits = 8
		h.pfR, h.pfG, h.pfB = 0, 0, 0
		h.pfA = 0xFF
	case FormatL8:
		h.flags |= ddsFlagPitch
		h.pitch = width
		h.pfFlags = ddpfLuminance
		h.pfBits = 8
		h.pfR = 0xFF
		h.pfG, h.pfB, h.pfA = 0, 0, 0
	default:
		return nil, fmt.Errorf("formato não suportado: %q", format)
	}
	return h.bytes(), nil
}

// ddsFormat devolve o formato Phyre de um header DDS (é o getDDSFormat do
// tool C++); erro quando não é reconhecido.
func ddsFormat(h ddsHeader) (string, error) {
	if h.pfFlags&ddpfFourCC != 0 {
		switch h.pfFourCC {
		case fccDXT1:
			return FormatDXT1, nil
		case fccDXT3:
			return FormatDXT3, nil
		case fccDXT5:
			return FormatDXT5, nil
		case fccBC5U, fccATI2:
			return FormatBC5, nil
		}
		return "", fmt.Errorf("dds: FourCC 0x%08X não reconhecido", h.pfFourCC)
	}
	switch {
	case h.pfFlags&ddpfAlphaPixels != 0 && h.pfFlags&ddpfRGB != 0 && h.flags&ddsFlagPitch != 0:
		return FormatARGB8, nil
	case (h.pfFlags&ddpfAlphaPixels != 0 || h.pfFlags&ddpfAlpha != 0) &&
		h.pfFlags&ddpfRGB == 0 && h.flags&ddsFlagPitch != 0:
		return FormatA8, nil
	case h.pfFlags&ddpfLuminance != 0 && h.pfFlags&ddpfRGB == 0 && h.flags&ddsFlagPitch != 0:
		return FormatL8, nil
	}
	return "", fmt.Errorf("dds: formato não reconhecido (pfFlags=0x%08X dwFlags=0x%08X)", h.pfFlags, h.flags)
}

// bufferSizeByFormat é o getBufferSizeByFormat do tool C++ (valor gravado
// em maxTextureBufferSize no header DX11).
func bufferSizeByFormat(format string, width, height uint32) uint32 {
	switch format {
	case FormatDXT5, FormatDXT3, FormatBC5, FormatA8, FormatL8:
		return width * height
	case FormatDXT1:
		return width * height / 2
	case FormatARGB8:
		return 4 * width * height
	}
	return 0
}
