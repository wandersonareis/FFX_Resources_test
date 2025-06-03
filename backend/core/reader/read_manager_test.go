package reader_test

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/reader"
	"ffxresources/backend/core/writer"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestReadManager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ReadManager Suite")
}

var _ = Describe("ReadManager", func() {
	var (
		originalResourcesRoot string
		tempDir               string
	)

	BeforeEach(func() {
		// Save original ResourcesRoot
		originalResourcesRoot = components.ResourcesRoot

		// Create temporary directory structure that mimics real structure
		var err error
		tempDir, err = os.MkdirTemp("", "ffx_test")
		Expect(err).ToNot(HaveOccurred())

		// Set ResourcesRoot to our test directory
		components.ResourcesRoot = tempDir

		// Create test directory structure
		createTestDirectories(tempDir)
		createTestFiles(tempDir)
	})

	AfterEach(func() {
		// Clean up test directory first
		if tempDir != "" {
			os.RemoveAll(tempDir)
		}

		// Restore original ResourcesRoot
		components.ResourcesRoot = originalResourcesRoot

		// Clear any global state that might have been set
		// Reset character maps
		components.ByteToCharMaps = make(map[string]map[uint]rune)
		components.CharToByteMaps = make(map[string]map[rune]uint)
		// Reset macro lookup
		components.MacroLookup = make(map[int]*components.LocalizedMacroStringObject)
	})

	Context("when all required files exist", func() {
		It("should initialize all character maps for available charsets", func() {
			reader.InitializeInternals()

			// Verify that character maps were created for each charset
			for _, charset := range common.Charsets {
				byteToChar, exists := components.ByteToCharMaps[charset]
				Expect(exists).To(BeTrue(), "ByteToChar map should exist for charset %s", charset)
				Expect(byteToChar).ToNot(BeEmpty(), "ByteToChar map should be populated for charset %s", charset)

				charToByte, exists := components.CharToByteMaps[charset]
				Expect(exists).To(BeTrue(), "CharToByte map should exist for charset %s", charset)
				Expect(charToByte).ToNot(BeEmpty(), "CharToByte map should be populated for charset %s", charset)
			}
		})

		It("should prepare string macros for all localizations", func() {
			reader.InitializeInternals()

			// Verify that macro lookups were created
			// Since we can't access specific macros by localization directly,
			// we'll check that the global MacroLookup was populated
			Expect(components.MacroLookup).ToNot(BeEmpty(), "MacroLookup should be populated")
		})
		It("should handle Korean and Chinese localizations without output", func() {
			// This test verifies that kr and ch localizations are processed
			// but with printOutput set to false
			reader.InitializeInternals()

			// We can't directly test the printOutput behavior in unit tests,
			// but we can verify the function completes without error
			Expect(components.MacroLookup).ToNot(BeNil())
		})
	})

	Context("when charset files are missing", func() {
		BeforeEach(func() {
			// Remove charset files to test error handling
			encodingDir := filepath.Join(tempDir, "ffx_encoding")
			os.RemoveAll(encodingDir)
		})

		It("should handle missing charset files gracefully", func() {
			// This depends on how PrepareCharset handles errors
			// You might want to adjust this based on actual error handling
			Expect(func() {
				reader.InitializeInternals()
			}).ToNot(Panic())
		})
	})

	Context("when macro dictionary files are missing", func() {
		BeforeEach(func() {
			// Remove macro dictionary files
			for loc := range common.Localizations {
				locDir := filepath.Join(tempDir, "new_"+loc+"pc")
				os.RemoveAll(locDir)
			}
		})

		It("should handle missing macro dictionary files gracefully", func() {
			// This depends on how PrepareStringMacros handles errors
			// You might want to adjust this based on actual error handling
			Expect(func() {
				reader.InitializeInternals()
			}).ToNot(Panic())
		})
	}) // Note: GetLocalizationRoot tests are skipped because they require
	// modifying a constant (common.PathFfxRoot) which cannot be changed in tests.
	// This function is tested indirectly through integration tests.

	Context("should read all event files", func() {
		/* It("should read all event files and populate the EventFiles map", func() {
			// Initialize internals to read event files
			reader.InitializeInternals()

			err := reader.ReadAllEvents(false)
			Expect(err).ToNot(HaveOccurred(), "Reading all events should not return an error")

			// Verify that EventFiles map is populated
			Expect(components.EVENTS).ToNot(BeEmpty(), "EventFiles should be populated")
			writer.WriteEventFileForAllLocalizations(true)
			reader.EditAndSaveEventCsv(true)

			// Check if specific event files exist
			for _, loc := range common.Localizations {
				eventFilePath := filepath.Join(tempDir, "new_"+loc+"pc", "menu", "event.dcp")
				Expect(components.EVENTS).To(HaveKey(eventFilePath), "Event file for %s should exist", loc)
			}
		}) */

		It("should read event file and write event file binary successfully", func() {
			// Initialize internals to read event files
			reader.InitializeInternals()

			err := reader.ReadAllEvents(false)
			Expect(err).ToNot(HaveOccurred(), "Reading all events should not return an error")

			// Verify that EventFiles map is populated
			Expect(components.EVENTS).ToNot(BeEmpty(), "EventFiles should be populated")
			writer.WriteEventFileForAllLocalizationsJSON(true)
			Expect(reader.EditAndSaveEventJSONFiles(true)).To(Succeed())

			// Check if specific event files exist
			for _, loc := range common.Localizations {
				eventFilePath := filepath.Join(tempDir, "new_"+loc+"pc", "menu", "event.dcp")
				Expect(components.EVENTS).To(HaveKey(eventFilePath), "Event file for %s should exist", loc)
			}
		})
	})
})

// Helper functions for test setup
func createTestDirectories(tempDir string) {
	// Create ffx_encoding directory
	encodingDir := filepath.Join(tempDir, "ffx_encoding")
	os.MkdirAll(encodingDir, 0755)

	// Create localization directories
	for loc := range common.Localizations {
		locDir := filepath.Join(tempDir, "new_"+loc+"pc", "menu")
		os.MkdirAll(locDir, 0755)
	}
}

func createTestFiles(tempDir string) {
	// Create test charset files
	for _, charset := range common.Charsets {
		charsetFile := filepath.Join(tempDir, "ffx_encoding", "ffxsjistbl_"+charset+".bin")
		// Create a simple test charset file with some sample data
		testData := []byte("テストデータ") // Some test Japanese characters
		os.WriteFile(charsetFile, testData, 0644)
	}

	// Create test macro dictionary files
	for loc := range common.Localizations {
		macroFile := filepath.Join(tempDir, "new_"+loc+"pc", "menu", "macrodic.dcp")
		// Create a simple test macro file
		testMacroData := []byte{0x00, 0x01, 0x02, 0x03} // Simple test data
		os.WriteFile(macroFile, testMacroData, 0644)
	}
}
