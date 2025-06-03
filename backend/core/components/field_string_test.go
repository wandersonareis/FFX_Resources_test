package components_test

import (
	"ffxresources/backend/core/components"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("FieldString", func() {
	var (
		testBytes []byte
		charset   string
	)

	BeforeEach(func() {
		// Setup test data
		charset = "us"
		setupFieldStringCharMaps()

		// Create test byte data that mimics the field string format
		// Header format: count (2 bytes) + headers (8 bytes each) + string data
		testBytes = []byte{
			// Count: 2 strings (16 bytes total for headers)
			0x10, 0x00, // count = 16 (0x0010) -> 16/8 = 2 strings
			0x00, 0x00, // padding

			// First string header (8 bytes)
			0x14, 0x00, 0x00, 0x00, // regular: offset=20 (0x14), flags=0, choices=0
			0x18, 0x00, 0x01, 0x02, // simplified: offset=24 (0x18), flags=1, choices=2

			// Second string header (8 bytes)
			0x1C, 0x00, 0x00, 0x00, // regular: offset=28 (0x1C), flags=0, choices=0
			0x1C, 0x00, 0x00, 0x00, // simplified: offset=28 (same as regular)

			// String data starting at offset 16 (0x10)
			0x41, 0x42, 0x43, 0x00, // "ABC" + null terminator at offset 20 (0x14)
			0x44, 0x45, 0x00, // "DE" + null terminator at offset 24 (0x18)
			0x00, // null terminator at offset 28 (0x1C) - empty string
		}
	})

	Describe("NewFieldString", func() {
		Context("when creating a new FieldString", func() {
			It("should correctly parse header data", func() {
				regularHeader := 0x02010014    // offset=20, flags=1, choices=2
				simplifiedHeader := 0x03020018 // offset=24, flags=2, choices=3

				fs := components.NewFieldString(charset, regularHeader, simplifiedHeader, testBytes)

				Expect(fs.Charset).To(Equal(charset))
				Expect(fs.RegularOffset).To(Equal(20))
				Expect(fs.RegularFlags).To(Equal(1))
				Expect(fs.RegularChoices).To(Equal(2))
				Expect(fs.SimplifiedOffset).To(Equal(24))
				Expect(fs.SimplifiedFlags).To(Equal(2))
				Expect(fs.SimplifiedChoices).To(Equal(3))
			})

			It("should extract correct byte sequences", func() {
				regularHeader := 0x00000014    // offset=20
				simplifiedHeader := 0x00000018 // offset=24

				fs := components.NewFieldString(charset, regularHeader, simplifiedHeader, testBytes)

				// Should extract "ABC" and "DE"
				Expect(fs.RegularBytes).To(Equal([]byte{0x41, 0x42, 0x43}))
				Expect(fs.SimplifiedBytes).To(Equal([]byte{0x44, 0x45}))
			})

			It("should handle same offset for regular and simplified", func() {
				regularHeader := 0x00000014    // offset=20
				simplifiedHeader := 0x00000014 // same offset=20

				fs := components.NewFieldString(charset, regularHeader, simplifiedHeader, testBytes)

				// Both should point to same byte sequence
				Expect(fs.RegularBytes).To(Equal(fs.SimplifiedBytes))
			})
		})
	})

	Describe("FromFieldStringData", func() {
		Context("when parsing field string data", func() {
			It("should correctly parse multiple field strings", func() {
				strings, err := components.FromFieldStringData(testBytes, false, charset)

				Expect(err).ToNot(HaveOccurred())
				Expect(strings).To(HaveLen(2))

				// First string
				Expect(strings[0].RegularOffset).To(Equal(20))
				Expect(strings[0].SimplifiedOffset).To(Equal(24))
				Expect(strings[0].SimplifiedFlags).To(Equal(1))
				Expect(strings[0].SimplifiedChoices).To(Equal(2))

				// Second string
				Expect(strings[1].RegularOffset).To(Equal(28))
				Expect(strings[1].SimplifiedOffset).To(Equal(28))
			})

			It("should handle empty byte arrays", func() {
				strings, err := components.FromFieldStringData([]byte{}, false, charset)

				Expect(err).ToNot(HaveOccurred())
				Expect(strings).To(BeEmpty())
			})

			It("should support print option", func() {
				// This mainly tests that print doesn't cause errors
				strings, err := components.FromFieldStringData(testBytes, true, charset)

				Expect(err).ToNot(HaveOccurred())
				Expect(strings).To(HaveLen(2))
			})
		})
	})

	Describe("Header conversion methods", func() {
		var fs *components.FieldString

		BeforeEach(func() {
			fs = &components.FieldString{
				RegularOffset:     0x1234,
				RegularFlags:      0x56,
				RegularChoices:    0x78,
				SimplifiedOffset:  0x9ABC,
				SimplifiedFlags:   0xDE,
				SimplifiedChoices: 0xF0,
			}
		})

		It("should convert regular header to bytes correctly", func() {
			result := fs.ToRegularHeaderBytes()
			expected := 0x78561234 // choices=0x78, flags=0x56, offset=0x1234
			Expect(result).To(Equal(expected))
		})

		It("should convert simplified header to bytes correctly", func() {
			result := fs.ToSimplifiedHeaderBytes()
			expected := 0xF0DE9ABC // choices=0xF0, flags=0xDE, offset=0x9ABC
			Expect(result).To(Equal(expected))
		})
	})

	Describe("String methods", func() {
		var fs *components.FieldString

		BeforeEach(func() {
			fs = components.NewFieldString(charset, 0x00000014, 0x00000018, testBytes)
		})

		It("should return regular string", func() {
			result := fs.GetRegularString()
			Expect(result).To(Equal("ABC"))
		})

		It("should return simplified string", func() {
			result := fs.GetSimplifiedString()
			Expect(result).To(Equal("DE"))
		})

		It("should detect distinct simplified strings", func() {
			Expect(fs.HasDistinctSimplified()).To(BeTrue())
		})

		It("should detect same simplified strings", func() {
			sameFs := components.NewFieldString(charset, 0x00000014, 0x00000014, testBytes)
			Expect(sameFs.HasDistinctSimplified()).To(BeFalse())
		})

		It("should format string representation correctly", func() {
			result := fs.String()
			Expect(result).To(Equal("ABC (Simplified: DE)"))
		})

		It("should format string representation without simplified when same", func() {
			sameFs := components.NewFieldString(charset, 0x00000014, 0x00000014, testBytes)
			result := sameFs.String()
			Expect(result).To(Equal("ABC"))
		})

		It("should detect empty strings", func() {
			emptyFs := components.NewFieldString(charset, 0x0000001C, 0x0000001C, testBytes)
			Expect(emptyFs.IsEmpty()).To(BeTrue())
		})
	})

	Describe("String modification methods", func() {
		var fs *components.FieldString

		BeforeEach(func() {
			fs = components.NewFieldString(charset, 0x00000014, 0x00000018, testBytes)
		})

		It("should set regular string", func() {
			fs.SetRegularString("NEW")
			Expect(fs.GetRegularString()).To(Equal("NEW"))
		})

		It("should set simplified string", func() {
			fs.SetSimplifiedString("SIMPLE")
			Expect(fs.GetSimplifiedString()).To(Equal("SIMPLE"))
		})

		It("should sync simplified when setting regular on synced strings", func() {
			// Create a synced field string
			syncedFs := components.NewFieldString(charset, 0x00000014, 0x00000014, testBytes)
			Expect(syncedFs.HasDistinctSimplified()).To(BeFalse())

			syncedFs.SetRegularString("SYNCED")
			Expect(syncedFs.GetRegularString()).To(Equal("SYNCED"))
			Expect(syncedFs.GetSimplifiedString()).To(Equal("SYNCED"))
		})

		It("should update charset", func() {
			fs.SetCharset("jp")
			Expect(fs.Charset).To(Equal("jp"))
		})

		It("should not update charset when same", func() {
			originalCharset := fs.Charset
			fs.SetCharset(originalCharset)
			Expect(fs.Charset).To(Equal(originalCharset))
		})

		It("should not update charset when empty", func() {
			originalCharset := fs.Charset
			fs.SetCharset("")
			Expect(fs.Charset).To(Equal(originalCharset))
		})
	})

	Describe("RebuildFieldStrings", func() {
		Context("when rebuilding field strings", func() {
			var fieldStrings []*components.FieldString

			BeforeEach(func() {
				// Create test field strings
				fs1 := &components.FieldString{
					Charset: charset,
				}
				fs1.SetRegularString("Hello")
				fs1.SetSimplifiedString("Hi")

				fs2 := &components.FieldString{
					Charset: charset,
				}
				fs2.SetRegularString("World")
				fs2.SetSimplifiedString("World") // Same as regular

				fieldStrings = []*components.FieldString{fs1, fs2}
			})

			It("should rebuild strings into byte format", func() {
				result := components.RebuildFieldStrings(fieldStrings, charset, false)

				// Should contain the string bytes
				Expect(result).ToNot(BeEmpty())
				// The exact content depends on StringToBytes implementation
			})

			It("should handle empty field strings list", func() {
				result := components.RebuildFieldStrings([]*components.FieldString{}, charset, false)
				Expect(result).To(BeEmpty())
			})
		})
	})

	Describe("LocalizedFieldStringObject", func() {
		var (
			localizedObj *components.LocalizedFieldStringObject
			testFS1      *components.FieldString
			testFS2      *components.FieldString
		)

		BeforeEach(func() {
			setupFieldStringCharMaps()
			localizedObj = components.NewLocalizedFieldStringObject()

			// Create test field strings
			testFS1 = &components.FieldString{
				Charset:      "us",
				RegularBytes: []byte{0x41, 0x42, 0x43}, // "ABC"
			}
			testFS2 = &components.FieldString{
				Charset:      "jp",
				RegularBytes: []byte{0x44, 0x45, 0x46}, // "DEF"
			}
		})

		Describe("NewLocalizedFieldStringObject", func() {
			It("should create an empty object", func() {
				obj := components.NewLocalizedFieldStringObject()
				Expect(obj).ToNot(BeNil())
				Expect(obj.Contents).ToNot(BeNil())
				Expect(len(obj.Contents)).To(Equal(0))
			})

			It("should create an object with initial content", func() {
				obj := components.NewLocalizedFieldStringObjectWithContent("us", testFS1)
				Expect(obj).ToNot(BeNil())
				Expect(len(obj.Contents)).To(Equal(1))
				Expect(obj.GetLocalizedContent("us")).To(Equal(testFS1))
			})
		})

		Describe("SetLocalizedContent", func() {
			Context("when setting new content", func() {
				It("should add content for a localization", func() {
					localizedObj.SetLocalizedContent("us", testFS1)
					Expect(localizedObj.GetLocalizedContent("us")).To(Equal(testFS1))
				})

				It("should update existing content", func() {
					localizedObj.SetLocalizedContent("us", testFS1)
					localizedObj.SetLocalizedContent("us", testFS2)
					Expect(localizedObj.GetLocalizedContent("us")).To(Equal(testFS2))
				})
			})

			Context("when setting empty content", func() {
				BeforeEach(func() {
					localizedObj.SetLocalizedContent("us", testFS1)
				})

				It("should not overwrite existing content with empty content", func() {
					emptyFS := &components.FieldString{
						Charset:      "us",
						RegularBytes: []byte{},
					}
					localizedObj.SetLocalizedContent("us", emptyFS)
					Expect(localizedObj.GetLocalizedContent("us")).To(Equal(testFS1))
				})
			})
		})

		Describe("ReadAndSetLocalizedContent", func() {
			It("should create and set content from byte data", func() {
				testBytes := []byte{
					0x41, 0x42, 0x43, 0x00, // "ABC" + null
				}
				regularHeader := 0x00000000
				simplifiedHeader := 0x00000000

				localizedObj.ReadAndSetLocalizedContent("us", testBytes, regularHeader, simplifiedHeader)

				content := localizedObj.GetLocalizedContent("us")
				Expect(content).ToNot(BeNil())
				Expect(content.Charset).To(Equal("us"))
			})

			It("should handle nil bytes gracefully", func() {
				localizedObj.ReadAndSetLocalizedContent("us", nil, 0, 0)
				Expect(localizedObj.GetLocalizedContent("us")).To(BeNil())
			})
		})

		Describe("GetLocalizedContent and GetLocalizedString", func() {
			BeforeEach(func() {
				localizedObj.SetLocalizedContent("us", testFS1)
				localizedObj.SetLocalizedContent("jp", testFS2)
			})

			It("should return the correct content for a localization", func() {
				Expect(localizedObj.GetLocalizedContent("us")).To(Equal(testFS1))
				Expect(localizedObj.GetLocalizedContent("jp")).To(Equal(testFS2))
			})

			It("should return nil for non-existent localization", func() {
				Expect(localizedObj.GetLocalizedContent("fr")).To(BeNil())
			})

			It("should return the string representation", func() {
				usString := localizedObj.GetLocalizedString("us")
				jpString := localizedObj.GetLocalizedString("jp")

				Expect(usString).ToNot(BeEmpty())
				Expect(jpString).ToNot(BeEmpty())
			})

			It("should return empty string for non-existent localization", func() {
				Expect(localizedObj.GetLocalizedString("fr")).To(Equal(""))
			})
		})

		Describe("GetDefaultContent", func() {
			It("should return US localization as default", func() {
				localizedObj.SetLocalizedContent("us", testFS1)
				localizedObj.SetLocalizedContent("jp", testFS2)

				defaultContent := localizedObj.GetDefaultContent()
				Expect(defaultContent).To(Equal(testFS1))
			})

			It("should return nil if no US content exists", func() {
				localizedObj.SetLocalizedContent("jp", testFS2)
				defaultContent := localizedObj.GetDefaultContent()
				Expect(defaultContent).To(BeNil())
			})
		})

		Describe("CopyInto", func() {
			It("should copy all content to another object", func() {
				localizedObj.SetLocalizedContent("us", testFS1)
				localizedObj.SetLocalizedContent("jp", testFS2)

				otherObj := components.NewLocalizedFieldStringObject()
				localizedObj.CopyInto(otherObj)

				Expect(otherObj.GetLocalizedContent("us")).To(Equal(testFS1))
				Expect(otherObj.GetLocalizedContent("jp")).To(Equal(testFS2))
			})
		})

		Describe("WriteAllContent", func() {
			It("should format all content with localization labels", func() {
				localizedObj.SetLocalizedContent("us", testFS1)
				localizedObj.SetLocalizedContent("jp", testFS2)

				result := localizedObj.WriteAllContent()
				Expect(result).To(ContainSubstring("["))
				Expect(result).To(ContainSubstring("]"))
			})
		})

		Describe("WriteAllContentToCsv", func() {
			It("should generate CSV format with headers", func() {
				localizedObj.SetLocalizedContent("us", testFS1)

				csv := localizedObj.WriteAllContentToCsv()

				Expect(csv).To(ContainSubstring("\"string index\""))
				Expect(csv).To(ContainSubstring("\n"))
				Expect(csv).To(ContainSubstring("\"us\""))
			})

			It("should handle empty content", func() {
				csv := localizedObj.WriteAllContentToCsv()
				Expect(csv).ToNot(BeEmpty())
				Expect(csv).To(ContainSubstring("\"string index\""))
			})
		})

		Describe("String", func() {
			It("should return default content string representation", func() {
				localizedObj.SetLocalizedContent("us", testFS1)
				result := localizedObj.String()
				Expect(result).ToNot(BeEmpty())
			})

			It("should return empty string when no default content", func() {
				result := localizedObj.String()
				Expect(result).To(Equal(""))
			})
		})
	})
})

// Helper function to setup basic character maps for FieldString testing
func setupFieldStringCharMaps() {
	// Basic ASCII character mapping for testing
	usMap := make(map[uint]rune)
	usReverseMap := make(map[rune]uint)

	// Add some basic characters
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789 "
	for i, char := range chars {
		byteVal := uint(0x30 + i) // Start from 0x30 to avoid command bytes
		usMap[byteVal] = char
		usReverseMap[char] = byteVal
	}

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
}
