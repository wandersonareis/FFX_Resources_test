package text_test

import (
	"ffxresources/backend/common"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/text"
	"ffxresources/backend/fileFormats/internal/text/internal/lib"
	"ffxresources/backend/fileFormats/internal/text/internal/mt2"
	"ffxresources/backend/fileFormats/internal/text/textverify"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/models"
	testcommon "ffxresources/testData"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestKrnl(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Krnl Suite")
}

var _ = Describe("KrnlFile", Ordered, func() {
	var (
		formatter             interfaces.ITextFormatter
		destination           locations.IDestination
		krnlCompressor        mt2.IKrnlCompressor
		krnlExtractor         mt2.IKrnlExtractor
		integrityVerification textverify.ITextVerificationService
		rootDir               string
		binaryPath            string
		gameVersionDir        string
		extractTempPath       string
		reimportTempPath      string
		translatePath         string
		testDataPath          string
		gameLocationPath      string
		config                *interactions.FFXAppConfig
		mockNotifierService   *testcommon.MockNotifier
		temp                  *common.TempProvider
		log                   *testcommon.MockLogHandler
	)

	BeforeAll(func() {
		Expect(testcommon.SetBuildBinPath()).To(Succeed())
		Expect(os.Setenv("FFX_GAME_VERSION", "2")).To(Succeed())

		rootDir = testcommon.GetTestDataRootDirectory()
		Expect(rootDir).NotTo(BeEmpty(), "Project root directory should not be empty")

		binaryPath = "binary"
		temp = common.NewTempProvider("", "")
		gameVersionDir = "FFX-2"

		testDataPath = filepath.Join(rootDir, gameVersionDir)
		gameLocationPath = filepath.Join(testDataPath, binaryPath)

		extractTempPath = filepath.Join(temp.TempFilePath, "extract")
		reimportTempPath = filepath.Join(temp.TempFilePath, "reimport")
		translatePath = filepath.Join(testDataPath, "translated")

		config = &interactions.FFXAppConfig{
			FFXGameVersion:    2,
			GameFilesLocation: gameLocationPath,
			ExtractLocation:   extractTempPath,
			TranslateLocation: translatePath,
			ImportLocation:    reimportTempPath,
		}

		formatter = &formatters.TxtFormatter{
			TargetExtension: ".txt",
			GameVersionDir:  gameVersionDir,
			GameFilesPath:   translatePath,
		}

		interactions.NewInteractionServiceWithConfig(config)
		interactions.NewInteractionWithTextFormatter(formatter)

		log = testcommon.NewLogHandlerMock()

		// Setup destination
		destination = &locations.Destination{
			ExtractLocation:   locations.NewExtractLocation("extracted", extractTempPath, gameVersionDir),
			TranslateLocation: locations.NewTranslateLocation("translated", translatePath, gameVersionDir),
			ImportLocation:    locations.NewImportLocation("reimported", reimportTempPath, gameVersionDir),
		}
	})

	AfterEach(func() {
		Expect(common.RemoveDir(temp.TempFilePath)).To(Succeed())
	})

	Context("Kernel file Extract Functionality", func() {
		BeforeEach(func() {
			krnlExtractor = mt2.NewKrnlExtractor(log)
			Expect(krnlExtractor).NotTo(BeNil())

			krnlCompressor = nil
			integrityVerification = nil
		})

		AfterEach(func() {
			Expect(common.RemoveDir(temp.TempFilePath)).To(Succeed())
		})

		var extract = func(fileRelPath string) {
			testFilePath := filepath.Join(gameLocationPath, fileRelPath)
			Expect(common.CheckPathExists(testFilePath)).To(Succeed())

			source, err := locations.NewSource(testFilePath)
			Expect(err).To(BeNil())
			Expect(source).NotTo(BeNil())

			Expect(destination).NotTo(BeNil())
			Expect(destination.InitializeLocations(source, formatter)).To(Succeed())

			Expect(krnlExtractor.Extract(source, destination)).To(Succeed())
		}

		It("should extract the lm_accesary.bin successfully", func() {
			extract(`ffx_ps2\ffx2\master\new_uspc\lastmiss\kernel\lm_accesary.bin`)
		})

		It("should extract the a_ability.bin successfully", func() {
			extract(`ffx_ps2\ffx2\master\new_uspc\battle\kernel\a_ability.bin`)
		})

		/* It("should throw error", func() {
			file := `ffx_ps2\ffx2\master\new_uspc\battle\btl\bika07_235\bika07_235.bin`
			testFilePath := filepath.Join(gameLocationPath, file)
			Expect(common.CheckPathExists(testFilePath)).To(Succeed())

			source, err := locations.NewSource(testFilePath)
			Expect(err).To(BeNil())
			Expect(source).NotTo(BeNil())

			Expect(destination).NotTo(BeNil())

			Expect(destination.InitializeLocations(source, formatter)).To(Succeed())

			m := new(testcommon.MockDlgDecoder)
			m.On("Decoder", source, destination, mock.Anything).Return(fmt.Errorf("mock error"))

			mockDlgDecoder := dlg.DialogExtractor{
				DialogDecoder: m,
				Logger:        log,
			}

			err = mockDlgDecoder.Extract(source, destination)
			Expect(err).To(HaveOccurred())
		}) */
	})

	Context("Kernel file Compress Functionality", func() {
		BeforeEach(func() {
			krnlCompressor = mt2.NewKrnlCompressor(log)
			Expect(krnlCompressor).NotTo(BeNil())
			Expect(destination).NotTo(BeNil())

			krnlExtractor = nil
		})

		AfterEach(func() {
			Expect(common.RemoveDir(temp.TempFilePath)).To(Succeed())
		})

		var compress = func(relPath string) {
			testFilePath := filepath.Join(gameLocationPath, relPath)
			Expect(common.CheckPathExists(testFilePath)).To(Succeed())

			source, err := locations.NewSource(testFilePath)
			Expect(err).To(BeNil())

			Expect(destination.InitializeLocations(source, formatter)).To(Succeed())

			Expect(krnlCompressor.Compress(source, destination)).To(Succeed())
			Expect(common.CheckPathExists(destination.Import().GetTargetFile())).To(Succeed())
		}

		It("should compress the lm_accesary.bin successfully", func() {
			compress(`ffx_ps2\ffx2\master\new_uspc\lastmiss\kernel\lm_accesary.bin`)
		})

		It("should compress the a_ability.bin successfully", func() {
			compress(`ffx_ps2\ffx2\master\new_uspc\battle\kernel\a_ability.bin`)
		})
	})

	Context("Verification Integrity Functionality", func() {
		BeforeEach(func() {
			krnlExtractor = nil
			krnlCompressor = nil
			integrityVerification = textverify.NewTextVerificationService()
			Expect(integrityVerification).NotTo(BeNil())
			mockNotifierService = testcommon.NewMockNotifier()
			Expect(mockNotifierService).NotTo(BeNil())
		})

		AfterEach(func() {
			Expect(common.RemoveDir(temp.TempFilePath)).To(Succeed())
		})

		It("should verify file integrity binary fail", func() {
			testPath := `binary\ffx_ps2\ffx2\master\new_uspc\lastmiss\kernel\lm_accesary.bin`
			sourceFile := filepath.Join(testDataPath, testPath)
			Expect(common.CheckPathExists(sourceFile)).To(Succeed())

			source, err := locations.NewSource(sourceFile)
			Expect(err).To(BeNil())

			Expect(destination).NotTo(BeNil())

			Expect(destination.InitializeLocations(source, formatter)).To(Succeed())

			krnlFile := text.NewKernel(source, destination)
			Expect(krnlFile).NotTo(BeNil())
			Expect(krnlFile.Extract()).To(Succeed())

			Expect(testcommon.RemoveFirstNLines(destination.Extract().GetTargetFile(), 4)).To(Succeed())

			err = integrityVerification.Verify(source, destination, textverify.NewTextExtractionVerificationStrategy())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("source and target segments count mismatch"))

			err = common.CheckPathExists(destination.Extract().GetTargetFile())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("path does not exist:"))
		})
	})

	Context("KernelFile Functionality", func() {
		BeforeEach(func() {
			krnlExtractor = nil
			krnlCompressor = nil
			integrityVerification = textverify.NewTextVerificationService()
			Expect(integrityVerification).NotTo(BeNil())
			mockNotifierService = testcommon.NewMockNotifier()
			Expect(mockNotifierService).NotTo(BeNil())
		})

		AfterEach(func() {
			Expect(common.RemoveDir(temp.TempFilePath)).To(Succeed())
		})

		It("should extract kernelFile lm_accesary.bin successfully", func() {
			testPath := `ffx_ps2\ffx2\master\new_uspc\lastmiss\kernel\lm_accesary.bin`
			sourceFile := filepath.Join(gameLocationPath, testPath)
			Expect(common.CheckPathExists(sourceFile)).To(Succeed(), "File should exist: %s", sourceFile)

			source, err := locations.NewSource(sourceFile)
			Expect(err).To(BeNil())
			Expect(source).NotTo(BeNil())

			Expect(destination).NotTo(BeNil())
			Expect(destination.InitializeLocations(source, formatter)).To(Succeed())

			krnlFile := text.NewKernel(source, destination)
			Expect(krnlFile).NotTo(BeNil())
			Expect(krnlFile.Extract()).To(Succeed())

			Expect(integrityVerification.Verify(source, destination, textverify.NewTextExtractionVerificationStrategy())).To(Succeed())
			Expect(mockNotifierService.Notifications).To(HaveLen(0))
		})

		It("should compress kernelFile a_ability.bin successfully", func() {
			testPath := `ffx_ps2\ffx2\master\new_uspc\battle\kernel\a_ability.bin`
			sourceFile := filepath.Join(gameLocationPath, testPath)
			Expect(common.CheckPathExists(sourceFile)).To(Succeed(), "File should exist: %s", sourceFile)

			source, err := locations.NewSource(sourceFile)
			Expect(err).To(BeNil())
			Expect(source).NotTo(BeNil())

			Expect(destination).NotTo(BeNil())
			Expect(destination.InitializeLocations(source, formatter)).To(Succeed())

			krnlFile := text.NewKernel(source, destination)
			Expect(krnlFile).NotTo(BeNil())
			Expect(krnlFile.Extract()).To(Succeed())
			Expect(krnlFile.Compress()).To(Succeed())

			Expect(integrityVerification.Verify(source, destination, textverify.NewTextCompressionVerificationStrategy())).To(Succeed())
			Expect(mockNotifierService.Notifications).To(HaveLen(0))
			Expect(common.CheckPathExists(destination.Import().GetTargetFile())).To(Succeed(), "File should exist: %s", destination.Import().GetTargetFile())
			Expect(common.CheckPathExists(destination.Extract().GetTargetFile())).To(Succeed(), "File should exist: %s", destination.Extract().GetTargetFile())
		})
	})

	Context("Kernel Count Functionality", func() {
		It("should be dialogs count", func() {
			testPath := `ffx_ps2\ffx2\master\new_uspc\lastmiss\kernel\lm_accesary.bin`
			sourceFile := filepath.Join(gameLocationPath, testPath)
			Expect(common.CheckPathExists(sourceFile)).To(Succeed())

			source, err := locations.NewSource(sourceFile)
			Expect(err).To(BeNil())
			Expect(source).NotTo(BeNil())

			encoding := ffxencoding.NewFFXTextEncodingFactory().CreateFFXTextDlgEncoding(source.GetType())
			defer encoding.Dispose()

			gameVersion := interactions.NewInteractionService().FFXGameVersion().GetGameVersion()
			Expect(gameVersion).To(Equal(models.FFX2))

			count, err := lib.TextSegmentsCounter(sourceFile, source.GetType(), gameVersion)
			Expect(err).To(BeNil())
			Expect(count).To(Equal(328))
		})
	})
})
