package event_test

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/encoding"
	"ffxresources/backend/fileFormats/event"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLocFieldString(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LocalizedFieldString Suite")
}

var _ = Describe("LocalizedFieldStringObject", func() {
	var (
		obj       *event.LocalizedFieldStringObject
		testBytes []byte
		charset   string
	)

	BeforeEach(func() {
		obj = event.NewLocalizedFieldStringObject()
		charset = "us"
		setupLocalizedCharMaps()

		// Create test byte data
		testBytes = []byte{
			0x08, 0x00, 0x00, 0x00, // header data
			0x0C, 0x00, 0x00, 0x00, // more header data
			0x50, 0x51, 0x52, 0x00, // "ABC" + null terminator at offset 16 (0x10)
			0x53, 0x54, 0x00, // "DE" + null terminator at offset 20 (0x14)
			0x50, 0x51, 0x52, 0x00, // "ABC" + null terminator at offset 16 (0x10)
		}
	})

	Describe("NewLocalizedFieldStringObject", func() {
		It("should create a new empty object", func() {
			newObj := event.NewLocalizedFieldStringObject()
			Expect(newObj).ToNot(BeNil())
			Expect(newObj.Contents).ToNot(BeNil())
			Expect(newObj.Contents).To(BeEmpty())
		})
	})

	Describe("NewLocalizedFieldStringObjectWithContent", func() {
		It("should create object with initial content", func() {
			fieldString := event.NewFieldString(charset, 0x00000010, 0x00000010, testBytes, common.GameVersionFFX)
			newObj := event.NewLocalizedFieldStringObjectWithContent("us", fieldString)

			Expect(newObj).ToNot(BeNil())
			Expect(newObj.GetLocalizedContent("us")).To(Equal(fieldString))
		})
	})

	Describe("SetLocalizedContent", func() {
		Context("when setting new content", func() {
			It("should store the content for the localization", func() {
				fieldString := event.NewFieldString(charset, 0x00000010, 0x00000010, testBytes, common.GameVersionFFX)
				obj.SetLocalizedContent("us", fieldString)

				Expect(obj.GetLocalizedContent("us")).To(Equal(fieldString))
			})

			It("should overwrite existing non-empty content", func() {
				// Set initial content
				fieldString1 := event.NewFieldString(charset, 0x00000010, 0x00000010, testBytes, common.GameVersionFFX)
				obj.SetLocalizedContent("us", fieldString1)

				// Set new content
				fieldString2 := event.NewFieldString(charset, 0x00000011, 0x00000011, testBytes, common.GameVersionFFX)
				obj.SetLocalizedContent("us", fieldString2)

				Expect(obj.GetLocalizedContent("us")).To(Equal(fieldString2))
			})

			It("should not overwrite existing content with empty content", func() {
				// Set initial content
				fieldString1 := event.NewFieldString(charset, 0x00000010, 0x00000010, testBytes, common.GameVersionFFX)
				obj.SetLocalizedContent("us", fieldString1)

				// Try to set empty content (offset pointing to null)
				emptyFieldString := &event.FieldString{
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
			obj.ReadAndSetLocalizedContent("us", testBytes, 0x00000008, 0x0000000C, common.GameVersionFFX)

			content := obj.GetLocalizedContent("us")
			str := content.GetRegularString()
			Expect(content).ToNot(BeNil())
			Expect(str).To(Equal("ABC"))
			str = content.GetSimplifiedString()
			Expect(str).To(Equal("DE"))
		})

		It("should handle nil bytes gracefully", func() {
			obj.ReadAndSetLocalizedContent("us", nil, 0x00000010, 0x00000014, common.GameVersionFFX)
			Expect(obj.GetLocalizedContent("us")).To(BeNil())
		})
	})

	Describe("GetLocalizedContent", func() {
		It("should return content for existing localization", func() {
			fieldString := event.NewFieldString(charset, 0x00000010, 0x00000010, testBytes, common.GameVersionFFX)
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
			fieldString := event.NewFieldString(charset, 0x00000008, 0x00000008, testBytes, common.GameVersionFFX)
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
			fieldString := event.NewFieldString(charset, 0x00000010, 0x00000010, testBytes, common.GameVersionFFX)
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
			fieldString1 := event.NewFieldString(charset, 0x00000010, 0x00000010, testBytes, common.GameVersionFFX)
			fieldString2 := event.NewFieldString(charset, 0x00000014, 0x00000014, testBytes, common.GameVersionFFX)

			obj.SetLocalizedContent("us", fieldString1)
			obj.SetLocalizedContent("jp", fieldString2)

			// Create target object
			target := event.NewLocalizedFieldStringObject()

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
			fieldString1 := event.NewFieldString(charset, 0x00000008, 0x00000008, testBytes, common.GameVersionFFX)
			fieldString2 := event.NewFieldString(charset, 0x0000000C, 0x0000000C, testBytes, common.GameVersionFFX)

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

	Describe("String", func() {
		It("should return string representation of default content", func() {
			fieldString := event.NewFieldString(charset, 0x00000008, 0x00000008, testBytes, common.GameVersionFFX)
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
	usMap[0x50] = 'A'
	usMap[0x51] = 'B'
	usMap[0x52] = 'C'
	usMap[0x53] = 'D'
	usMap[0x54] = 'E'
	usReverseMap['A'] = 0x50
	usReverseMap['B'] = 0x51
	usReverseMap['C'] = 0x52
	usReverseMap['D'] = 0x53
	usReverseMap['E'] = 0x54

	ffxencoding.SetCharMap(common.GameVersionFFX, "us", usMap, usReverseMap)
	ffxencoding.SetCharMap(common.GameVersionFFX, "jp", usMap, usReverseMap)
}
