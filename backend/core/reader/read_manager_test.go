package reader_test

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/reader"
	"ffxresources/backend/datastore"
	"ffxresources/backend/core/encoding"
	testcommon "ffxresources/testData"
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
	Context("when testing FFX (ffx)", func() {
		BeforeEach(func() {
			common.SetCurrentGameVersion(common.GameVersionFFX)
			common.SetGameFilesRoot(filepath.Join(rootDir, "FFX", "binary"))
		})

		Context("when all required files exist", func() {
			It("should initialize all character maps for available charsets", func() {
				Expect(reader.InitializeInternals()).To(Succeed())

				// Verify that character maps were created for each charset
				for _, charset := range common.Charsets {
					byteToChar := ffxencoding.GetByteToCharMap(common.GameVersionFFX, charset)
					Expect(byteToChar).ToNot(BeNil(), "ByteToChar map should exist for charset %s", charset)
					Expect(byteToChar).ToNot(BeEmpty(), "ByteToChar map should be populated for charset %s", charset)

					charToByte := ffxencoding.GetCharToByteMap(common.GameVersionFFX, charset)
					Expect(charToByte).ToNot(BeNil(), "CharToByte map should exist for charset %s", charset)
					Expect(charToByte).ToNot(BeEmpty(), "CharToByte map should be populated for charset %s", charset)
				}
			})

			It("should prepare string macros for all localizations", func() {
				Expect(reader.InitializeInternals()).To(Succeed())

				// Verify that macros were published into the datastore
				// Since we can't access specific macros by localization directly,
				// we'll check that the datastore macros were populated
				Expect(datastore.GetMacros(common.GameVersionFFX).Count()).To(BeNumerically(">", 0), "Datastore macros should be populated")
			})
			It("should handle Korean and Chinese localizations without output", func() {
				// This test verifies that kr and ch localizations are processed
				// but with printOutput set to false
				Expect(reader.InitializeInternals()).To(Succeed())

				// We can't directly test the printOutput behavior in unit tests,
				// but we can verify the function completes without error
				Expect(datastore.GetMacros(common.GameVersionFFX)).ToNot(BeNil())
			})
		})
	})

	Context("when testing FFX-2 (ffx2)", func() {
		BeforeEach(func() {
			common.SetCurrentGameVersion(common.GameVersionFFX2)
			common.SetGameFilesRoot(filepath.Join(rootDir, "FFX-2", "binary"))
		})

		Context("when all required files exist", func() {
			It("should initialize all character maps for available charsets", func() {
				// PrepareVersion recebe a versão explícita: a versão global
				// (config do app via interactions) não controla os buckets
				// de charset, e o config do ambiente de teste é sempre FFX.
				Expect(reader.PrepareVersion(common.GameVersionFFX2)).To(Succeed())

				// Verify that character maps were created for each charset
				for _, charset := range common.Charsets {
					byteToChar := ffxencoding.GetByteToCharMap(common.GameVersionFFX2, charset)
					Expect(byteToChar).ToNot(BeNil(), "ByteToChar map should exist for charset %s", charset)
					Expect(byteToChar).ToNot(BeEmpty(), "ByteToChar map should be populated for charset %s", charset)

					charToByte := ffxencoding.GetCharToByteMap(common.GameVersionFFX2, charset)
					Expect(charToByte).ToNot(BeNil(), "CharToByte map should exist for charset %s", charset)
					Expect(charToByte).ToNot(BeEmpty(), "CharToByte map should be populated for charset %s", charset)
				}
			})

			It("should prepare string macros for all localizations", func() {
				Expect(reader.PrepareVersion(common.GameVersionFFX2)).To(Succeed())

				// Verify that macros were published into the datastore
				// Since we can't access specific macros by localization directly,
				// we'll check that the datastore macros were populated
				Expect(datastore.GetMacros(common.GameVersionFFX2).Count()).To(BeNumerically(">", 0), "Datastore macros should be populated")
			})
			It("should handle Korean and Chinese localizations without output", func() {
				// This test verifies that kr and ch localizations are processed
				// but with printOutput set to false
				Expect(reader.PrepareVersion(common.GameVersionFFX2)).To(Succeed())

				// We can't directly test the printOutput behavior in unit tests,
				// but we can verify the function completes without error
				Expect(datastore.GetMacros(common.GameVersionFFX2)).ToNot(BeNil())
			})
		})
	})

	Context("when game files are missing", func() {
		BeforeEach(func() {
			common.SetGameFilesRoot("missing_resources")
		})
		AfterEach(func() {
			common.SetGameFilesRoot(originalResourcesRoot)
		})

		It("inicializa sem erro: charsets são embutidos e macros ausentes são toleradas", func() {
			// Os charsets vivem em core/encoding/charset_tables.go — não há
			// mais I/O de tabela, então nada falha por recurso ausente aqui.
			// Macros ausentes são ignoradas silenciosamente (PrepareStringMacros).
			Expect(reader.InitializeInternals()).To(Succeed())
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
})
