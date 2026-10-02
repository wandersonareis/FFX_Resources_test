package ddsphyre

import (
	"bytes"
	"fmt"
	"image"
	"image/png"

	"github.com/woozymasta/bcn"
)

// bcnFormat mapeia o formato Phyre/DDS para o decoder BCn.
func bcnFormat(format string) (bcn.Format, error) {
	switch format {
	case FormatDXT1:
		return bcn.FormatBC1, nil
	case FormatDXT3:
		return bcn.FormatBC2, nil
	case FormatDXT5:
		return bcn.FormatBC3, nil
	case FormatBC5:
		return bcn.FormatBC5, nil
	case FormatARGB8:
		// Máscaras R=0x00FF0000 … A=0xFF000000 ⇒ ordem de bytes BGRA.
		return bcn.FormatBGRA8, nil
	case FormatA8:
		return bcn.FormatA8, nil
	case FormatL8:
		return bcn.FormatR8, nil
	}
	return bcn.FormatUnknown, fmt.Errorf("formato %q sem decoder", format)
}

// DDSToPNG decodifica um DDS completo em PNG (mip 0). Serve tanto para o
// arquivo extraído em disco quanto para o payload extraído do .dds.phyre.
func DDSToPNG(dds []byte) ([]byte, error) {
	hdr, err := parseDDSHeader(dds)
	if err != nil {
		return nil, err
	}
	if hdr.width == 0 || hdr.height == 0 {
		return nil, fmt.Errorf("dds: dimensões inválidas %dx%d", hdr.width, hdr.height)
	}
	format, err := ddsFormat(hdr)
	if err != nil {
		return nil, err
	}
	bf, err := bcnFormat(format)
	if err != nil {
		return nil, err
	}
	if len(dds) < ddsHeaderLen {
		return nil, fmt.Errorf("dds: sem payload")
	}
	payload := dds[ddsHeaderLen:]
	img, err := bcn.DecodeImage(payload, int(hdr.width), int(hdr.height), bf)
	if err != nil {
		return nil, fmt.Errorf("decodificando %s %dx%d: %w", format, hdr.width, hdr.height, err)
	}
	// O payload é gravado de baixo para cima (herança DX/DIB): sem o flip a
	// textura sai de cabeça para baixo. Aqui é o ÚNICO lugar que vira — é a
	// pré-visualização que tem que parecer com o jogo. O DDS exportado
	// mantém os bytes originais, é ele que volta no repack.
	flipVertical(img)
	if format == FormatA8 {
		// A8 é só canal alfa (o decoder deixa RGB=0): como máscara, o mais
		// legível é o próprio alfa em escala de cinza — vira silhueta.
		pix := img.Pix
		for i := 0; i+3 < len(pix); i += 4 {
			v := pix[i+3]
			pix[i], pix[i+1], pix[i+2], pix[i+3] = v, v, v, 255
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("gerando png: %w", err)
	}
	return buf.Bytes(), nil
}

// flipVertical espelha as linhas da imagem in place (a última passa a ser a
// primeira). Linhas são trocadas de ponta a ponta — o conteúdo pixel a pixel
// não muda, só a ordem, então a qualidade é exata (nada é reamostrado).
func flipVertical(img *image.NRGBA) {
	if img == nil || img.Rect.Dx() == 0 || img.Rect.Dy() < 2 {
		return
	}
	w, h := img.Rect.Dx(), img.Rect.Dy()
	row := w * 4
	if row > img.Stride {
		return
	}
	tmp := make([]byte, row)
	for y := 0; y < h/2; y++ {
		top := img.Pix[y*img.Stride : y*img.Stride+row]
		bottom := img.Pix[(h-1-y)*img.Stride : (h-1-y)*img.Stride+row]
		copy(tmp, top)
		copy(top, bottom)
		copy(bottom, tmp)
	}
}

// ToPNG decodifica o payload do próprio .dds.phyre em PNG.
func (t *Texture) ToPNG() ([]byte, error) {
	dds, err := t.ExtractToDDS()
	if err != nil {
		return nil, err
	}
	return DDSToPNG(dds)
}
