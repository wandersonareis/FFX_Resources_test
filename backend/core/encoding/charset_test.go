package ffxencoding_test

import (
	"testing"

	"ffxresources/backend/common"
	"ffxresources/backend/core/encoding"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCharset(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Charset Suite")
}

var _ = Describe("PrepareCharset", func() {
	AfterEach(func() {
		// Limpa os mapas para não vazar estado entre specs.
		ffxencoding.ClearAllCharMaps()
	})

	// assertVersionedCharset carrega todos os charsets da versão e verifica
	// os mapas + consistência do canônico (encode → decode é identidade).
	assertVersionedCharset := func(version common.GameVersion) {
		for _, charset := range common.Charsets {
			err := ffxencoding.PrepareCharset(version, charset)
			Expect(err).ToNot(HaveOccurred(), "PrepareCharset should succeed for charset: %s", charset)

			byteToChar := ffxencoding.GetByteToCharMap(version, charset)
			Expect(byteToChar).ToNot(BeNil(), "ByteToCharMaps should contain charset: %s", charset)
			Expect(len(byteToChar)).To(BeNumerically(">", 0), "ByteToCharMaps[%s] should not be empty", charset)

			charToByte := ffxencoding.GetCharToByteMap(version, charset)
			Expect(charToByte).ToNot(BeNil(), "CharToByteMaps should contain charset: %s", charset)
			Expect(len(charToByte)).To(BeNumerically(">", 0), "CharToByteMaps[%s] should not be empty", charset)

			// Encode do canônico → decode devolve o mesmo rune.
			for char, byteVal := range charToByte {
				mappedChar, err := ffxencoding.ByteToChar(byteVal, charset, version)
				Expect(err).ToNot(HaveOccurred(), "ByteToChar should contain byte %d from charset %s", byteVal, charset)
				Expect(mappedChar).To(Equal(char), "Forward mapping should be consistent for byte 0x%02X in charset %s", byteVal, charset)
			}
		}
	}

	Context("when testing FFX (ffx)", func() {
		It("should prepare charset maps for all expected charsets", func() {
			assertVersionedCharset(common.GameVersionFFX)
		})

		It("aplica as correções de font nos slots da tabela us", func() {
			Expect(ffxencoding.PrepareCharset(common.GameVersionFFX, "us")).To(Succeed())
			b2c := ffxencoding.GetByteToCharMap(common.GameVersionFFX, "us")
			// Glifos comprovados pela atlas (font_0_0/font_0_1):
			Expect(b2c[0xA7]).To(Equal('Ç'), "font desenha Ç maiúsculo (tabela tinha ç)")
			Expect(b2c[0xBE]).To(Equal('ç'), "ç minúsculo legítimo permanece em 0xBE")
			Expect(b2c[0x93]).To(Equal('Œ'), "font desenha Œ (tabela diz espaço)")
			Expect(b2c[0x97]).To(Equal('œ'), "font desenha œ")
			Expect(b2c[0xA6]).To(Equal('Ã'), "font trocou a-trema por a-til (idioma us)")
			Expect(b2c[0xBD]).To(Equal('ã'), "idem, minúsculo")
			Expect(b2c[0x3C]).To(Equal('"'), "aspas retas (glifo do slot)")
			Expect(b2c[0x95]).To(Equal('”'), "aspas curvas/inclinadas — rune único, sem duplicata")
			Expect(b2c[0x41]).To(Equal('\''), "apóstrofo reto")
			Expect(b2c[0xD5]).To(Equal('’'), "apóstrofo curvo/inclinado — rune único, sem duplicata")
			c2b := ffxencoding.GetCharToByteMap(common.GameVersionFFX, "us")
			Expect(c2b['"']).To(Equal(uint(0x3C)), "canônico de \" é o slot reto")
			Expect(c2b['”']).To(Equal(uint(0x95)), "canônico de ” é o slot do glifo inclinado")
			Expect(c2b['\'']).To(Equal(uint(0x41)), "canônico de ' é o slot reto")
			Expect(c2b['Œ']).To(Equal(uint(0x93)))
			Expect(c2b['Ã']).To(Equal(uint(0xA6)))
		})
	})

	Context("when testing FFX-2 (ffx2)", func() {
		It("should prepare charset maps for all expected charsets", func() {
			assertVersionedCharset(common.GameVersionFFX2)
		})
	})

	Context("lastmiss", func() {
		It("normaliza lastmiss para os mapas de ffx2", func() {
			Expect(ffxencoding.PrepareCharset(common.GameVersionLastMiss, "us")).To(Succeed())
			Expect(ffxencoding.GetByteToCharMap(common.GameVersionFFX2, "us")).ToNot(BeNil())
			Expect(ffxencoding.GetByteToCharMap(common.GameVersionLastMiss, "us")).To(BeNil(),
				"lastmiss não tem bucket próprio: reaproveita ffx2")
		})
	})

	Context("GetCharsetForLanguage", func() {
		It("usa tabela dedicada para ch/jp/kr", func() {
			Expect(ffxencoding.GetCharsetForLanguage("ch")).To(Equal("ch"))
			Expect(ffxencoding.GetCharsetForLanguage("jp")).To(Equal("jp"))
			Expect(ffxencoding.GetCharsetForLanguage("kr")).To(Equal("kr"))
		})

		It("retorna us para idiomas desconhecidos", func() {
			for _, lang := range []string{"us", "en", "zz"} {
				Expect(ffxencoding.GetCharsetForLanguage(lang)).To(Equal("us"),
					"idioma %q deve resolver para a tabela us", lang)
			}
		})

		It("retorna default (a-trema) para sp/fr/de/it", func() {
			// Fonts desses idiomas preservam o glifo do a-trema; a troca
			// trema→til vale só para o font do us.
			for _, lang := range []string{"sp", "fr", "de", "it"} {
				Expect(ffxencoding.GetCharsetForLanguage(lang)).To(Equal("default"),
					"idioma %q deve resolver para a tabela default", lang)
			}
		})
	})

	Context("tabela default (a-trema)", func() {
		It("mantém Ä/ä e as demais correções de font", func() {
			Expect(ffxencoding.PrepareCharset(common.GameVersionFFX, "default")).To(Succeed())
			b2c := ffxencoding.GetByteToCharMap(common.GameVersionFFX, "default")
			Expect(b2c[0xA6]).To(Equal('Ä'), "default preserva o a-trema")
			Expect(b2c[0xBD]).To(Equal('ä'), "default preserva o a-trema minúsculo")
			// Demais correções de font permanecem:
			Expect(b2c[0xA7]).To(Equal('Ç'))
			Expect(b2c[0x93]).To(Equal('Œ'))
			Expect(b2c[0x97]).To(Equal('œ'))
			Expect(b2c[0x3C]).To(Equal('"'))
			Expect(b2c[0x95]).To(Equal('”'))
			Expect(b2c[0x41]).To(Equal('\''))
			Expect(b2c[0xD5]).To(Equal('’'))
		})
	})

	Context("when charset is unknown", func() {
		It("should return an error", func() {
			err := ffxencoding.PrepareCharset(common.GameVersionFFX, "nonexistent")
			Expect(err).To(HaveOccurred(), "PrepareCharset should fail for unknown charset")
		})
	})
})
