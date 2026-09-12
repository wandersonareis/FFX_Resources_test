package event_test

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/reader"
	"ffxresources/backend/fileFormats/event"
	testcommon "ffxresources/testData"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestEventsJSONReaders(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Events JSON Readers Suite")
}

var _ = Describe("Events JSON Readers", Ordered, func() {
	var (
		originalResourcesRoot string
		originalGameFilesRoot string
		tempEditDir           string
	)

	BeforeAll(func() {
		Expect(testcommon.SetBuildBinPath()).To(Succeed())
		common.SetCurrentGameVersion(common.GameVersionFFX)
		common.SetVerboseMode(false)
		Expect(reader.InitializeInternals()).To(Succeed())
	})

	BeforeEach(func() {
		// Save original paths
		originalResourcesRoot = common.ResourcesRoot
		originalGameFilesRoot = common.GameFilesRoot

		// Create temporary edit directory for test JSON files
		tempEditDir = filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")
		Expect(os.MkdirAll(tempEditDir, 0755)).To(Succeed())
	})

	AfterEach(func() {
		// Restore original paths
		common.ResourcesRoot = originalResourcesRoot
		common.GameFilesRoot = originalGameFilesRoot

		// Clean up test files
		if tempEditDir != "" {
			os.RemoveAll(tempEditDir)
		}
	})

	Context("ProcessEventsJsonFile", func() {
		It("should return error when JSON file does not exist", func() {
			// Don't create the JSON file
			err := event.ProcessEventsJsonFile(common.GameVersionFFX)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("JSON file not found"))
		})
	})

	Context("ProcessEventJsonFile", func() {
		It("should return error when JSON file does not exist", func() {
			// Don't create the JSON file
			err := event.ProcessEventJsonFile(common.GameVersionFFX, "ev001")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("JSON file not found"))
		})
	})
})
