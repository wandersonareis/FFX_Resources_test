package reader_test

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/reader"
	"ffxresources/backend/core/writer"
	"ffxresources/backend/datastore"
	"ffxresources/backend/fileFormats/event"
	"ffxresources/backend/core/encoding"
	"ffxresources/backend/interactions"
	"ffxresources/backend/models"
	testcommon "ffxresources/testData"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func interactionsVersion() int {
	return interactions.NewInteractionService().FFXAppConfig().GetGameVersion()
}

func TestReadManager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ReadManager Suite")
}

var _ = Describe("ReadManager", Ordered, func() {
	var (
		rootDir               string
		originalResourcesRoot string
		tempDir               string
	)

	BeforeAll(func() {
		Expect(testcommon.SetBuildBinPath()).To(Succeed())
		common.SetVerboseMode(false)
	})

	BeforeEach(func() {
		// Save original ResourcesRoot
		originalResourcesRoot = common.ResourcesRoot

		rootDir = testcommon.GetTestDataRootDirectory()
		Expect(rootDir).NotTo(BeEmpty(), "Project root directory should not be empty")
	})

	AfterEach(func() {
		// Restore original ResourcesRoot
		common.ResourcesRoot = originalResourcesRoot

		// Clear any global state that might have been set
		ffxencoding.ClearAllCharMaps()
		datastore.Instance.ClearAllMacros()
	})
	Context("when testing FFX (version 1)", func() {
		BeforeEach(func() {
			common.SetGameVersion(1)
			//common.SetGameFilesRoot(filepath.Join(rootDir, "FFX", "binary"))
		})

		Context("when all required files exist", func() {
			It("should initialize all character maps for available charsets", func() {
				Expect(reader.InitializeInternals()).To(Succeed())

				// Verify that character maps were created for each charset
				for _, charset := range common.Charsets {
					byteToChar := ffxencoding.GetByteToCharMap(models.FFX, charset)
					Expect(byteToChar).ToNot(BeNil(), "ByteToChar map should exist for charset %s", charset)
					Expect(byteToChar).ToNot(BeEmpty(), "ByteToChar map should be populated for charset %s", charset)

					charToByte := ffxencoding.GetCharToByteMap(models.FFX, charset)
					Expect(charToByte).ToNot(BeNil(), "CharToByte map should exist for charset %s", charset)
					Expect(charToByte).ToNot(BeEmpty(), "CharToByte map should be populated for charset %s", charset)
				}
			})

			It("should prepare string macros for all localizations", func() {
				Expect(reader.InitializeInternals()).To(Succeed())

				// Verify that macros were published into the datastore
				// Since we can't access specific macros by localization directly,
				// we'll check that the datastore macros were populated
				Expect(datastore.GetMacros(models.FFX).Count()).To(BeNumerically(">", 0), "Datastore macros should be populated")
			})
			It("should handle Korean and Chinese localizations without output", func() {
				// This test verifies that kr and ch localizations are processed
				// but with printOutput set to false
				Expect(reader.InitializeInternals()).To(Succeed())

				// We can't directly test the printOutput behavior in unit tests,
				// but we can verify the function completes without error
				Expect(datastore.GetMacros(models.FFX)).ToNot(BeNil())
			})
		})
	})

	Context("when testing FFX-2 (version 2)", func() {
		BeforeEach(func() {
			common.SetGameVersion(2)
			//common.SetGameFilesRoot(filepath.Join(rootDir, "FFX-2", "binary"))
		})

		Context("when all required files exist", func() {
			It("should initialize all character maps for available charsets", func() {
				Expect(reader.InitializeInternals()).To(Succeed())

				// Verify that character maps were created for each charset
				for _, charset := range common.Charsets {
					byteToChar := ffxencoding.GetByteToCharMap(models.FFX2, charset)
					Expect(byteToChar).ToNot(BeNil(), "ByteToChar map should exist for charset %s", charset)
					Expect(byteToChar).ToNot(BeEmpty(), "ByteToChar map should be populated for charset %s", charset)

					charToByte := ffxencoding.GetCharToByteMap(models.FFX2, charset)
					Expect(charToByte).ToNot(BeNil(), "CharToByte map should exist for charset %s", charset)
					Expect(charToByte).ToNot(BeEmpty(), "CharToByte map should be populated for charset %s", charset)
				}
			})

			It("should prepare string macros for all localizations", func() {
				Expect(reader.InitializeInternals()).To(Succeed())

				// Verify that macros were published into the datastore
				// Since we can't access specific macros by localization directly,
				// we'll check that the datastore macros were populated
				Expect(datastore.GetMacros(models.FFX2).Count()).To(BeNumerically(">", 0), "Datastore macros should be populated")
			})
			It("should handle Korean and Chinese localizations without output", func() {
				// This test verifies that kr and ch localizations are processed
				// but with printOutput set to false
				Expect(reader.InitializeInternals()).To(Succeed())

				// We can't directly test the printOutput behavior in unit tests,
				// but we can verify the function completes without error
				Expect(datastore.GetMacros(models.FFX2)).ToNot(BeNil())
			})
		})
	})

	Context("when charset files are missing", func() {
		BeforeEach(func() {
			common.SetGameFilesRoot("missing_resources")
		})
		AfterEach(func() {
			common.SetGameFilesRoot(originalResourcesRoot)
		})

		It("should handle missing charset files gracefully", func() {
			// This depends on how PrepareCharset handles errors
			// You might want to adjust this based on actual error handling
			err := reader.InitializeInternals()
			Expect(err).To(HaveOccurred(), "Initializing internals should return an error when resources are missing")
		})
	})

	Context("when macro dictionary files are missing", func() {
		BeforeEach(func() {
			// Remove macro dictionary files
			for loc := range common.SupportedLanguages {
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
	})

	Context("should read all event files", func() {
		It("should read event file and write event file binary successfully", func() {
			// Initialize internals to read event files
			err := reader.InitializeInternals()
			Expect(err).ToNot(HaveOccurred(), "Initializing internals should not return an error")

			eventsFolder, err := common.NewFileAccessor(common.GetPathOriginalsEvent())
			Expect(err).ToNot(HaveOccurred(), "Resolving events directory should not return an error")

			err = reader.ReadAllEvents(eventsFolder)
			Expect(err).ToNot(HaveOccurred(), "Reading all events should not return an error")

			// Verify that EventFiles map is populated
			Expect(event.HasEvents(models.GameVersion(interactionsVersion()))).To(BeTrue(), "Events should be populated in datastore")
			writer.ExportAllLocalizationsToJSON()
			Expect(reader.EditAndSaveEventJSONFiles()).To(Succeed())
		})
	})
})
