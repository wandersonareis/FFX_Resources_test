package components_test

import (
	"ffxresources/backend/core/components"
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
				result := components.StringToBytes("ABC", "us")
				Expect(result).To(Equal([]byte{0x50, 0x51, 0x52}))
			})

			It("should handle newline characters", func() {
				result := components.StringToBytes("A\nB", "us")
				Expect(result).To(ContainElement(byte(0x03))) // newline should become 0x03
			})

			It("should handle empty strings", func() {
				result := components.StringToBytes("", "us")
				Expect(result).To(BeEmpty())
			})
		})

		Context("when converting command strings", func() {
			It("should convert PAUSE command", func() {
				result := components.StringToBytes("{PAUSE}", "us")
				Expect(result).To(Equal([]byte{0x01}))
			})

			It("should convert line break command", func() {
				result := components.StringToBytes("{\\n}", "us")
				Expect(result).To(Equal([]byte{0x03}))
			})

			It("should convert color commands", func() {
				result := components.StringToBytes("{CLR:WHITE}", "us")
				Expect(result).To(Equal([]byte{0x0A, 0x41}))
			})

			It("should convert choice commands", func() {
				result := components.StringToBytes("{CHOICE:00}", "us")
				Expect(result).To(Equal([]byte{0x10, 0x30}))
			})

			It("should convert choice end command", func() {
				result := components.StringToBytes("{CHOICE-END}", "us")
				Expect(result).To(Equal([]byte{0x10, 0xFF}))
			})
		})

		Context("when handling unknown characters", func() {
			It("should handle characters not in charset gracefully", func() {
				// Don't expect it to panic, but result may contain no bytes for unknown chars
				result := components.StringToBytes("。", "us") // Character not in basic map
				Expect(result).To(BeNil())
			})
		})
	})

	Describe("BytesToString", func() {
		Context("when converting basic byte sequences", func() {
			It("should convert basic character bytes", func() {
				// Setup character map
				/* components.SetCharMap("us",
					map[int]rune{0x41: 'A', 0x42: 'B', 0x43: 'C'},
					map[rune]int{'A': 0x41, 'B': 0x42, 'C': 0x43}) */

				b := []uint8{19, 52, 3, 207, 209, 59, 58, 47, 203, 107, 46, 178, 47, 154, 137, 96, 142, 170, 162, 106, 129, 132, 59, 3, 25, 127, 141, 58, 136, 137, 106, 117, 132, 102, 174, 126, 59, 79} // Corresponds to 'A', 'B', 'C'

				result := components.BytesToString(b, "us")
				Expect(result).To(Equal("ABC"))
			})

			It("should handle empty byte arrays", func() {
				result := components.BytesToString([]byte{}, "us")
				Expect(result).To(Equal(""))
			})

			It("should handle null termination", func() {
				result := components.BytesToString([]byte{0x50, 0x00, 0x42}, "us")
				Expect(result).To(Equal("A")) // Should stop at null byte
			})
		})

		Context("when converting command bytes", func() {
			It("should convert pause command bytes", func() {
				result := components.BytesToString([]byte{0x01}, "us")
				Expect(result).To(Equal("{PAUSE}"))
			})

			It("should convert line break command bytes", func() {
				// Assuming WriteLinebreaksAsCommands is true
				components.WriteLinebreaksAsCommands = true
				result := components.BytesToString([]byte{0x03}, "us")
				Expect(result).To(Equal("{\\n}"))
			})

			It("should convert color command bytes", func() {
				result := components.BytesToString([]byte{0x0A, 0x41}, "us")
				Expect(result).To(Equal("{CLR:WHITE}"))
			})

			It("should convert choice command bytes", func() {
				result := components.BytesToString([]byte{0x10, 0x30}, "us")
				Expect(result).To(Equal("{CHOICE:00}"))
			})

			It("should convert choice end command bytes", func() {
				result := components.BytesToString([]byte{0x10, 0xFF}, "us")
				Expect(result).To(Equal("{CHOICE-END}"))
			})

			It("should convert variable command bytes", func() {
				result := components.BytesToString([]byte{0x12, 0x30}, "us")
				Expect(result).To(Equal("{VAR:00}"))
			})

			It("should convert player character command bytes", func() {
				result := components.BytesToString([]byte{0x13, 0x30}, "us")
				Expect(result).To(Equal("{PC:00:TIDUS}"))
			})
		})

		Context("when handling unknown bytes", func() {
			It("should handle unknown character bytes", func() {
				result := components.BytesToString([]byte{0xFF}, "us")
				Expect(result).To(ContainSubstring("{UNKCHR:FF}"))
			})
		})
	})

	Describe("Bidirectional conversion", func() {
		Context("when converting strings to bytes and back", func() {
			It("should maintain data integrity for simple strings", func() {
				original := "ABC"

				// Setup character map
				/* components.SetCharMap("us",
					map[int]rune{0x41: 'A', 0x42: 'B', 0x43: 'C'},
					map[rune]int{'A': 0x41, 'B': 0x42, 'C': 0x43}) */

				bytes := components.StringToBytes(original, "us")
				result := components.BytesToString(bytes, "us")

				Expect(result).To(Equal(original))
			})

			It("should maintain data integrity for command strings", func() {
				original := "{PAUSE}"

				bytes := components.StringToBytes(original, "us")
				result := components.BytesToString(bytes, "us")

				Expect(result).To(Equal(original))
			})
		})
	})

	Describe("Different localizations", func() {
		Context("when using different charset localizations", func() {
			It("should handle US localization", func() {
				result := components.StringToBytes("test", "us")
				Expect(result).NotTo(BeNil())
			})

			It("should not handle Japanese localization", func() {
				result := components.StringToBytes("test", "jp")
				Expect(result).To(BeNil())
			})

			It("should not handle Korean localization", func() {
				result := components.StringToBytes("test", "kr")
				Expect(result).To(BeNil())
			})

			It("should not handle Chinese localization", func() {
				result := components.StringToBytes("test", "ch")
				Expect(result).To(BeNil())
			})

			It("should not handle US localiztion", func() {
				result := components.StringToBytes("你好吗", "us")
				Expect(result).To(BeNil())
			})

			It("should handle Japanese localization", func() {
				result := components.StringToBytes("こんにちは", "jp")
				Expect(result).NotTo(BeNil())
			})

			It("should handle Korean localization", func() {
				result := components.StringToBytes("안녕하세요", "kr")
				Expect(result).NotTo(BeNil())
			})

			It("should handle Chinese localization", func() {
				result := components.StringToBytes("你好", "ch")
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

	/* components.SetCharMap("us", usMap, usReverseMap)
	components.SetCharMap("jp", usMap, usReverseMap)
	components.SetCharMap("kr", usMap, usReverseMap)
	components.SetCharMap("ch", usMap, usReverseMap) */

	reader.InitializeInternals() // Initialize character maps and macros
}
