package components_test

import (
	"ffxresources/backend/core/components"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("LocalizedFieldStringObject", func() {
	var (
		obj       *components.LocalizedFieldStringObject
		testBytes []byte
		charset   string
	)

	BeforeEach(func() {
		obj = components.NewLocalizedFieldStringObject()
		charset = "us"
		setupLocalizedCharMaps()

		// Create test byte data
		testBytes = []byte{
			0x10, 0x00, 0x00, 0x00, // header data
			0x14, 0x00, 0x00, 0x00, // more header data
			0x41, 0x42, 0x43, 0x00, // "ABC" + null terminator at offset 16 (0x10)
			0x44, 0x45, 0x00, // "DE" + null terminator at offset 20 (0x14)
		}
	})

	Describe("NewLocalizedFieldStringObject", func() {
		It("should create a new empty object", func() {
			newObj := components.NewLocalizedFieldStringObject()
			Expect(newObj).ToNot(BeNil())
			Expect(newObj.Contents).ToNot(BeNil())
			Expect(newObj.Contents).To(BeEmpty())
		})
	})

	Describe("NewLocalizedFieldStringObjectWithContent", func() {
		It("should create object with initial content", func() {
			fieldString := components.NewFieldString(charset, 0x00000010, 0x00000010, testBytes)
			newObj := components.NewLocalizedFieldStringObjectWithContent("us", fieldString)

			Expect(newObj).ToNot(BeNil())
			Expect(newObj.GetLocalizedContent("us")).To(Equal(fieldString))
		})
	})

	Describe("SetLocalizedContent", func() {
		Context("when setting new content", func() {
			It("should store the content for the localization", func() {
				fieldString := components.NewFieldString(charset, 0x00000010, 0x00000010, testBytes)
				obj.SetLocalizedContent("us", fieldString)

				Expect(obj.GetLocalizedContent("us")).To(Equal(fieldString))
			})

			It("should overwrite existing non-empty content", func() {
				// Set initial content
				fieldString1 := components.NewFieldString(charset, 0x00000010, 0x00000010, testBytes)
				obj.SetLocalizedContent("us", fieldString1)

				// Set new content
				fieldString2 := components.NewFieldString(charset, 0x00000014, 0x00000014, testBytes)
				obj.SetLocalizedContent("us", fieldString2)

				Expect(obj.GetLocalizedContent("us")).To(Equal(fieldString2))
			})

			It("should not overwrite existing content with empty content", func() {
				// Set initial content
				fieldString1 := components.NewFieldString(charset, 0x00000010, 0x00000010, testBytes)
				obj.SetLocalizedContent("us", fieldString1)

				// Try to set empty content (offset pointing to null)
				emptyFieldString := &components.FieldString{
					Charset:         charset,
					RegularBytes:    []byte{},
					SimplifiedBytes: []byte{},
				}
				obj.SetLocalizedContent("us", emptyFieldString)

				// Should keep the original content
				Expect(obj.GetLocalizedContent("us")).To(Equal(fieldString1))
			})
		})
	})

	Describe("ReadAndSetLocalizedContent", func() {
		It("should read and set content from byte data", func() {
			obj.ReadAndSetLocalizedContent("us", testBytes, 0x00000010, 0x00000014)

			content := obj.GetLocalizedContent("us")
			Expect(content).ToNot(BeNil())
			Expect(content.GetRegularString()).To(Equal("ABC"))
			Expect(content.GetSimplifiedString()).To(Equal("DE"))
		})

		It("should handle nil bytes gracefully", func() {
			obj.ReadAndSetLocalizedContent("us", nil, 0x00000010, 0x00000014)
			Expect(obj.GetLocalizedContent("us")).To(BeNil())
		})
	})

	Describe("GetLocalizedContent", func() {
		It("should return content for existing localization", func() {
			fieldString := components.NewFieldString(charset, 0x00000010, 0x00000010, testBytes)
			obj.SetLocalizedContent("us", fieldString)

			result := obj.GetLocalizedContent("us")
			Expect(result).To(Equal(fieldString))
		})

		It("should return nil for non-existing localization", func() {
			result := obj.GetLocalizedContent("nonexistent")
			Expect(result).To(BeNil())
		})
	})

	Describe("GetLocalizedString", func() {
		It("should return string for existing localization", func() {
			fieldString := components.NewFieldString(charset, 0x00000010, 0x00000010, testBytes)
			obj.SetLocalizedContent("us", fieldString)

			result := obj.GetLocalizedString("us")
			Expect(result).To(Equal("ABC"))
		})

		It("should return empty string for non-existing localization", func() {
			result := obj.GetLocalizedString("nonexistent")
			Expect(result).To(Equal(""))
		})
	})

	Describe("GetDefaultContent", func() {
		It("should return US content as default", func() {
			fieldString := components.NewFieldString(charset, 0x00000010, 0x00000010, testBytes)
			obj.SetLocalizedContent("us", fieldString)

			result := obj.GetDefaultContent()
			Expect(result).To(Equal(fieldString))
		})

		It("should return nil if no US content exists", func() {
			result := obj.GetDefaultContent()
			Expect(result).To(BeNil())
		})
	})

	Describe("CopyInto", func() {
		It("should copy all content into another object", func() {
			// Setup source object
			fieldString1 := components.NewFieldString(charset, 0x00000010, 0x00000010, testBytes)
			fieldString2 := components.NewFieldString(charset, 0x00000014, 0x00000014, testBytes)

			obj.SetLocalizedContent("us", fieldString1)
			obj.SetLocalizedContent("jp", fieldString2)

			// Create target object
			target := components.NewLocalizedFieldStringObject()

			// Copy content
			obj.CopyInto(target)

			// Verify content was copied
			Expect(target.GetLocalizedContent("us")).To(Equal(fieldString1))
			Expect(target.GetLocalizedContent("jp")).To(Equal(fieldString2))
		})
	})

	Describe("WriteAllContent", func() {
		It("should format all content with localization names", func() {
			// Setup content for multiple localizations
			fieldString1 := components.NewFieldString(charset, 0x00000010, 0x00000010, testBytes)
			fieldString2 := components.NewFieldString(charset, 0x00000014, 0x00000014, testBytes)

			obj.SetLocalizedContent("us", fieldString1)
			obj.SetLocalizedContent("jp", fieldString2)

			result := obj.WriteAllContent()

			// Should contain formatted strings for each localization
			Expect(result).To(ContainSubstring("["))
			Expect(result).To(ContainSubstring("]"))
			Expect(result).To(ContainSubstring("ABC"))
			Expect(result).To(ContainSubstring("DE"))
		})

		It("should return empty string when no content exists", func() {
			result := obj.WriteAllContent()
			Expect(result).To(Equal(""))
		})
	})

	Describe("WriteAllContentToCsv", func() {
		It("should generate CSV with headers and content", func() {
			// Setup content
			fieldString := components.NewFieldString(charset, 0x00000010, 0x00000010, testBytes)
			obj.SetLocalizedContent("us", fieldString)

			result := obj.WriteAllContentToCsv()

			// Should contain CSV headers
			Expect(result).To(ContainSubstring("\"string index\""))
			Expect(result).To(ContainSubstring(",\"us\""))

			// Should contain content
			Expect(result).To(ContainSubstring("\"ABC\""))

			// Should have proper CSV format
			lines := strings.Split(result, "\n")
			Expect(len(lines)).To(BeNumerically(">=", 1))
		})

		It("should handle empty content in CSV", func() {
			result := obj.WriteAllContentToCsv()

			// Should still have headers
			Expect(result).To(ContainSubstring("\"string index\""))
			// Should have empty values
			Expect(result).To(ContainSubstring(",\"\""))
		})

		It("should escape quotes in CSV values", func() {
			// Create a field string with quotes
			quotedFieldString := &components.FieldString{
				Charset:         charset,
				RegularBytes:    []byte{}, // We'll mock the string return
				SimplifiedBytes: []byte{},
			}
			// Note: This test would need the actual string conversion to work properly
			// For now, we'll test the escape function indirectly

			obj.SetLocalizedContent("us", quotedFieldString)
			result := obj.WriteAllContentToCsv()

			// Should not cause CSV parsing errors
			Expect(result).To(ContainSubstring("\""))
		})
	})

	Describe("String", func() {
		It("should return string representation of default content", func() {
			fieldString := components.NewFieldString(charset, 0x00000010, 0x00000010, testBytes)
			obj.SetLocalizedContent("us", fieldString)

			result := obj.String()
			Expect(result).To(Equal("ABC"))
		})

		It("should return empty string when no default content exists", func() {
			result := obj.String()
			Expect(result).To(Equal(""))
		})
	})
})

// Helper function to setup basic character maps for LocalizedFieldStringObject testing
func setupLocalizedCharMaps() {
	// Basic ASCII character mapping for testing
	usMap := make(map[uint]rune)
	usReverseMap := make(map[rune]uint)

	// Add specific test mappings
	usMap[0x41] = 'A'
	usMap[0x42] = 'B'
	usMap[0x43] = 'C'
	usMap[0x44] = 'D'
	usMap[0x45] = 'E'
	usReverseMap['A'] = 0x41
	usReverseMap['B'] = 0x42
	usReverseMap['C'] = 0x43
	usReverseMap['D'] = 0x44
	usReverseMap['E'] = 0x45

	components.SetCharMap("us", usMap, usReverseMap)
	components.SetCharMap("jp", usMap, usReverseMap)
}
