package components_test

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/converter"
	"ffxresources/backend/core/encoding"
	"ffxresources/backend/core/reader"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestStrings(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "String Conversion Functions Suite")
}

var _ = Describe("String Conversion Functions", func() {
	BeforeEach(func() {
		// Setup basic character maps for testing
		setupBasicCharMaps()
	})

	Describe("StringToBytes", func() {
		Context("when converting simple strings", func() {
			It("should convert basic ASCII characters", func() {
				result := converter.StringToBytes("ABC", "us", common.GameVersionFFX)
				Expect(result).To(Equal([]byte{0x50, 0x51, 0x52}))
			})

			It("should handle newline characters", func() {
				result := converter.StringToBytes("A\nB", "us", common.GameVersionFFX)
				Expect(result).To(ContainElement(byte(0x03))) // newline should become 0x03
			})

			It("should handle empty strings", func() {
				result := converter.StringToBytes("", "us", common.GameVersionFFX)
				Expect(result).To(BeEmpty())
			})
		})

		Context("when converting command strings", func() {
			It("should convert PAUSE command", func() {
				result := converter.StringToBytes("{PAUSE}", "us", common.GameVersionFFX)
				Expect(result).To(Equal([]byte{0x01}))
			})

			It("should convert line break command", func() {
				result := converter.StringToBytes("{\\n}", "us", common.GameVersionFFX)
				Expect(result).To(Equal([]byte{0x03}))
			})

			It("should convert color commands", func() {
				result := converter.StringToBytes("{CLR:WHITE}", "us", common.GameVersionFFX)
				Expect(result).To(Equal([]byte{0x0A, 0x41}))
			})

			It("should convert choice commands", func() {
				result := converter.StringToBytes("{CHOICE:00}", "us", common.GameVersionFFX)
				Expect(result).To(Equal([]byte{0x10, 0x30}))
			})

			It("should convert choice end command", func() {
				result := converter.StringToBytes("{CHOICE-END}", "us", common.GameVersionFFX)
				Expect(result).To(Equal([]byte{0x10, 0xFF}))
			})
		})

		Context("when handling unknown characters", func() {
			It("should handle characters not in charset gracefully", func() {
				// Don't expect it to panic, but result may contain no bytes for unknown chars
				result := converter.StringToBytes("。", "us", common.GameVersionFFX) // Character not in basic map
				Expect(result).To(BeNil())
			})
		})
	})

	Describe("BytesToString", func() {
		Context("when converting basic byte sequences", func() {
			It("should convert basic character bytes", func() {
				// Setup character map
				ffxencoding.SetCharMap(common.GameVersionFFX, "us",
					map[uint]rune{0x50: 'A', 0x51: 'B', 0x52: 'C'},
					map[rune]uint{'A': 0x50, 'B': 0x51, 'C': 0x52})

				b := []uint8{0x50, 0x51, 0x52} // Corresponds to 'A', 'B', 'C'

				result := converter.BytesToString(b, "us", common.GameVersionFFX)
				Expect(result).To(Equal("ABC"))
			})

			It("should handle empty byte arrays", func() {
				result := converter.BytesToString([]byte{}, "us", common.GameVersionFFX)
				Expect(result).To(Equal(""))
			})

			It("should handle null termination", func() {
				result := converter.BytesToString([]byte{0x50, 0x00, 0x42}, "us", common.GameVersionFFX)
				Expect(result).To(Equal("A")) // Should stop at null byte
			})
		})

		Context("when converting command bytes", func() {
			It("should convert pause command bytes", func() {
				result := converter.BytesToString([]byte{0x01}, "us", common.GameVersionFFX)
				Expect(result).To(Equal("{PAUSE}"))
			})

			It("should convert line break command bytes", func() {
				// Assuming WriteLinebreaksAsCommands is true
				converter.WriteLinebreaksAsCommands = true
				result := converter.BytesToString([]byte{0x03}, "us", common.GameVersionFFX)
				Expect(result).To(Equal("{\\n}"))
			})

			It("should convert color command bytes", func() {
				result := converter.BytesToString([]byte{0x0A, 0x41}, "us", common.GameVersionFFX)
				Expect(result).To(Equal("{CLR:WHITE}"))
			})

			It("should convert choice command bytes", func() {
				result := converter.BytesToString([]byte{0x10, 0x30}, "us", common.GameVersionFFX)
				Expect(result).To(Equal("{CHOICE:00}"))
			})

			It("should convert choice end command bytes", func() {
				result := converter.BytesToString([]byte{0x10, 0xFF}, "us", common.GameVersionFFX)
				Expect(result).To(Equal("{CHOICE-END}"))
			})

			It("should convert variable command bytes", func() {
				result := converter.BytesToString([]byte{0x12, 0x30}, "us", common.GameVersionFFX)
				Expect(result).To(Equal("{VAR:00}"))
			})

			It("should convert player character command bytes", func() {
				result := converter.BytesToString([]byte{0x13, 0x30}, "us", common.GameVersionFFX)
				Expect(result).To(Equal("{PC:00:TIDUS}"))
			})
		})
	})

	Describe("Bidirectional conversion", func() {
		Context("when converting strings to bytes and back", func() {
			It("should maintain data integrity for simple strings", func() {
				original := "ABC"

				// Setup character map
				ffxencoding.SetCharMap(common.GameVersionFFX, "us",
					map[uint]rune{0x50: 'A', 0x51: 'B', 0x52: 'C'},
					map[rune]uint{'A': 0x50, 'B': 0x51, 'C': 0x52})

				bytes := converter.StringToBytes(original, "us", common.GameVersionFFX)
				result := converter.BytesToString(bytes, "us", common.GameVersionFFX)

				Expect(result).To(Equal(original))
			})

			It("should maintain data integrity for command strings", func() {
				original := "{PAUSE}"

				bytes := converter.StringToBytes(original, "us", common.GameVersionFFX)
				result := converter.BytesToString(bytes, "us", common.GameVersionFFX)

				Expect(result).To(Equal(original))
			})
		})
	})

	Describe("Different localizations", func() {
		Context("when using different charset localizations", func() {
			It("should handle US localization", func() {
				result := converter.StringToBytes("test", "us", common.GameVersionFFX)
				Expect(result).NotTo(BeNil())
			})

			It("should not handle Japanese localization", func() {
				result := converter.StringToBytes("test", "jp", common.GameVersionFFX)
				Expect(result).To(BeNil())
			})

			It("should not handle Korean localization", func() {
				result := converter.StringToBytes("test", "kr", common.GameVersionFFX)
				Expect(result).To(BeNil())
			})

			It("should not handle Chinese localization", func() {
				result := converter.StringToBytes("test", "ch", common.GameVersionFFX)
				Expect(result).To(BeNil())
			})

			It("should not handle US localiztion", func() {
				result := converter.StringToBytes("你好吗", "us", common.GameVersionFFX)
				Expect(result).To(BeNil())
			})

			It("should handle Japanese localization", func() {
				result := converter.StringToBytes("こんにちは", "jp", common.GameVersionFFX)
				Expect(result).NotTo(BeNil())
			})

			It("should handle Korean localization", func() {
				result := converter.StringToBytes("안녕하세요", "kr", common.GameVersionFFX)
				Expect(result).NotTo(BeNil())
			})

			It("should handle Chinese localization", func() {
				result := converter.StringToBytes("你好", "ch", common.GameVersionFFX)
				Expect(result).NotTo(BeNil())
			})
		})
	})
})

// Helper function to setup basic character maps for testing
func setupBasicCharMaps() {
	// Basic ASCII character mapping for testing
	/* usMap := make(map[int]rune)
	usReverseMap := make(map[rune]int) */

	// Add some basic characters
	/* chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789 "
	for i, char := range chars {
		byteVal := 0x30 + i // Start from 0x30 to avoid command bytes
		usMap[byteVal] = char
		usReverseMap[char] = byteVal
	} */

	/* ffxencoding.SetCharMap("us", usMap, usReverseMap)
	ffxencoding.SetCharMap("jp", usMap, usReverseMap)
	ffxencoding.SetCharMap("kr", usMap, usReverseMap)
	ffxencoding.SetCharMap("ch", usMap, usReverseMap) */

	reader.InitializeInternals() // Initialize character maps and macros
}
