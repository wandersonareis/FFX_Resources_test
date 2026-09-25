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

		It("aplica o fix ç→Ç no slot 0xA7 da tabela us", func() {
			Expect(ffxencoding.PrepareCharset(common.GameVersionFFX, "us")).To(Succeed())
			b2c := ffxencoding.GetByteToCharMap(common.GameVersionFFX, "us")
			Expect(b2c[0xA7]).To(Equal('Ç'), "slot 0xA7 corrigido na fonte: bloco de maiúsculas")
			Expect(b2c[0xBE]).To(Equal('ç'), "ç minúsculo legítimo permanece em 0xBE")
			Expect(b2c[0x3C]).To(Equal('”'), "tabela guarda ” no slot do \" ASCII (duplicata genuína)")
			Expect(b2c[0x95]).To(Equal('”'))
			c2b := ffxencoding.GetCharToByteMap(common.GameVersionFFX, "us")
			Expect(c2b['Ç']).To(Equal(uint(0xA7)), "canônico de Ç é o slot corrigido")
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

	Context("when charset is unknown", func() {
		It("should return an error", func() {
			err := ffxencoding.PrepareCharset(common.GameVersionFFX, "nonexistent")
			Expect(err).To(HaveOccurred(), "PrepareCharset should fail for unknown charset")
		})
	})
})
