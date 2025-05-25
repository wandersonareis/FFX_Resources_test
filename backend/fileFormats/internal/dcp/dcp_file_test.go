package dcp_test

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/dcp"
	"ffxresources/backend/fileFormats/internal/dcp/internal/integrity"
	"ffxresources/backend/fileFormats/internal/dcp/internal/lib"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/models"
	testcommon "ffxresources/testData"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDcp(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dcp Suite")
}

var _ = Describe("DcpFile", Ordered, func() {
	var (
		rootDir           string
		binaryPath        string
		gameVersionDir    string
		extractTempPath   string
		reimportTempPath  string
		translatePath     string
		testFilePath      string
		testDataPath      string
		gameLocationPath  string
		globalSource      interfaces.ISource
		globalDestination locations.IDestination
		formatter         interfaces.ITextFormatter
		dcpFileProperties models.IDcpFileProperties
		destination       locations.IDestination
		dcpFile           interfaces.IFileProcessor
		dcpFileReader     dcp.IDcpFileExtractor
		dcpFileWriter     dcp.IDcpFileCompressor
		verifyService     components.IVerificationService
		config            *interactions.FFXAppConfig
		temp              *common.TempProvider
		log               *testcommon.MockLogHandler
	)

	var runCommonTests = func(expectedTextMsg, expectedBinaryMsg string) {
		It("should have correct extract location path", func() {
			expected := filepath.Join(extractTempPath, gameVersionDir, lib.DCP_PARTS_TARGET_DIR_NAME)
			expected = filepath.ToSlash(expected)
			actual := destination.Extract().GetTargetPath()
			actual = filepath.ToSlash(actual)
			Expect(actual).To(Equal(expected))
		})

		It("should have correct extract file", func() {
			Expect(dcpFileReader).NotTo(BeNil(), "DCP file extractor should not be nil")
			Expect(dcpFileReader.Extract()).To(Succeed())
		})

		It("should extract the DCP file successfully", func() {
			Expect(dcpFile).NotTo(BeNil(), "DCP file should not be nil")
			Expect(dcpFile.Extract()).To(Succeed())
		})

		It("should compress the DCP file successfully", func() {
			Expect(dcpFile).NotTo(BeNil(), "DCP file should not be nil")
			Expect(dcpFile.Extract()).To(Succeed())
			Expect(dcpFile.Compress()).To(Succeed())
		})

		It("should have integrity text check fail", func() {
			Expect(dcpFile).NotTo(BeNil(), "DCP file should not be nil")
			Expect(dcpFile.Extract()).To(Succeed())
			Expect(verifyService).NotTo(BeNil(), "DCP integrity service should not be nil")

			targetPath := destination.Extract().GetTargetPath()
			Expect(targetPath).NotTo(BeEmpty())
			Expect(dcpFileProperties).NotTo(BeNil())

			// Localiza os arquivos de texto e remove um arquivo aleatório
			path := filepath.Join(targetPath, "*"+formatter.GetTargetExtension())
			matchingFiles, err := filepath.Glob(path)
			Expect(err).To(BeNil())
			Expect(matchingFiles).NotTo(BeEmpty())

			if len(matchingFiles) > 0 {
				randIdx := rand.Intn(dcpFileProperties.GetPartsLength())
				randomFile := matchingFiles[randIdx]
				Expect(common.RemoveFileWithRetries(randomFile, 3)).To(Succeed())
			}

			Expect(globalSource).NotTo(BeNil(), "Global source should not be nil")
			Expect(globalDestination).NotTo(BeNil(), "Global destination should not be nil")
			//err = dcpIntegrity.Verify(targetPath, formatter, dcpFileProperties)
			err = verifyService.Verify(globalSource, globalDestination, integrity.NewDcpExtractionVerificationStrategy())
			Expect(err).To(HaveOccurred())
			// Usamos ContainSubstring para flexibilizar a verificação da mensagem
			Expect(err.Error()).To(ContainSubstring(expectedTextMsg))
		})

		It("should have integrity binary check fail", func() {
			Expect(dcpFile).NotTo(BeNil())
			Expect(dcpFile.Extract()).To(Succeed())
			Expect(verifyService).NotTo(BeNil())

			targetPath := destination.Extract().GetTargetPath()
			Expect(targetPath).NotTo(BeEmpty())
			Expect(dcpFileProperties).NotTo(BeNil())

			// Localiza os arquivos binários e remove um arquivo aleatório
			matchingFiles, err := filepath.Glob(filepath.Join(targetPath, "*.0??"))
			Expect(err).To(BeNil())
			Expect(matchingFiles).NotTo(BeEmpty())

			if len(matchingFiles) > 0 {
				randIdx := rand.Intn(dcpFileProperties.GetPartsLength())
				randomFile := matchingFiles[randIdx]
				Expect(common.RemoveFileWithRetries(randomFile, 3)).To(Succeed())
			}

			Expect(globalSource).NotTo(BeNil(), "Global source should not be nil")
			Expect(globalDestination).NotTo(BeNil(), "Global destination should not be nil")
			//err = dcpIntegrity.Verify(targetPath, formatter, dcpFileProperties)
			err = verifyService.Verify(globalSource, globalDestination, integrity.NewDcpExtractionVerificationStrategy())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(expectedBinaryMsg))
		})

		It("should have integrity check", func() {
			Expect(dcpFile).NotTo(BeNil())
			Expect(dcpFile.Extract()).To(Succeed())
			Expect(verifyService).NotTo(BeNil())

			targetPath := destination.Extract().GetTargetPath()
			Expect(targetPath).NotTo(BeEmpty())
			Expect(dcpFileProperties).NotTo(BeNil())

			Expect(globalSource).NotTo(BeNil(), "Global source should not be nil")
			Expect(globalDestination).NotTo(BeNil(), "Global destination should not be nil")
			Expect(verifyService.Verify(globalSource, globalDestination, integrity.NewDcpExtractionVerificationStrategy())).To(Succeed())
		})
	}

	// Testes para a versão FFX-2
	Context("Test FFX-2 dcp file test", Ordered, func() {
		BeforeEach(func() {
			Expect(testcommon.SetBuildBinPath()).To(Succeed())
			Expect(os.Setenv("FFX_GAME_VERSION", "2")).To(Succeed())

			rootDir = testcommon.GetTestDataRootDirectory()
			Expect(rootDir).NotTo(BeEmpty(), "Project root directory should not be empty")

			binaryPath = "binary"
			temp = common.NewTempProvider("", "")
			gameVersionDir = "FFX-2"

			testDataPath = filepath.Join(rootDir, gameVersionDir)
			gameLocationPath = filepath.Join(testDataPath, binaryPath)

			file := `ffx_ps2\ffx2\master\new_uspc\menu\macrodic.dcp`
			testFilePath = filepath.Join(gameLocationPath, file)
			Expect(common.CheckPathExists(testFilePath)).To(Succeed(), "Test file path should exist: %s", testFilePath)

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
				GameVersionDir: gameVersionDir,
				GameFilesPath:  translatePath,
			}

			dcpFileProperties = models.NewDcpFileOptions(models.GameVersion(config.FFXGameVersion))

			interactions.NewInteractionServiceWithConfig(config)
			interactions.NewInteractionWithTextFormatter(formatter)

			destination = &locations.Destination{
				ExtractLocation:   locations.NewExtractLocation("extracted", extractTempPath, gameVersionDir),
				TranslateLocation: locations.NewTranslateLocation("translated", translatePath, gameVersionDir),
				ImportLocation:    locations.NewImportLocation("reimported", reimportTempPath, gameVersionDir),
			}

			log = testcommon.NewLogHandlerMock()
			Expect(log).NotTo(BeNil())

			// Inicializa os objetos de extração/compressão
			source, err := locations.NewSource(testFilePath)
			Expect(err).To(BeNil())
			Expect(source).NotTo(BeNil())
			globalSource = source

			Expect(destination).NotTo(BeNil(), "Destination should not be nil")
			Expect(destination.InitializeLocations(source, formatter)).To(Succeed())
			globalDestination = destination

			dcpFileReader = dcp.NewDcpFileExtractor(source, destination, formatter, log)
			Expect(dcpFileReader).NotTo(BeNil())

			dcpFileWriter = dcp.NewDcpFileCompressor(source, destination, formatter, log)
			Expect(dcpFileWriter).NotTo(BeNil())

			dcpFile = dcp.NewDcpFile(source, destination)
			Expect(dcpFile).NotTo(BeNil())

			verifyService = components.NewVerificationService()
			Expect(verifyService).NotTo(BeNil())
		})

		AfterEach(func() {
			Expect(common.RemoveDir(extractTempPath)).To(Succeed())
		})

		// Para FFX-2 esperamos erro de integridade com quantidade: expected 6, got 7
		runCommonTests("expected 6, got 7", "expected 6, got 7")
	})

	// Testes para a versão FFX
	Context("Test FFX dcp file test", Ordered, func() {
		BeforeEach(func() {
			Expect(testcommon.SetBuildBinPath()).To(Succeed())
			Expect(os.Setenv("FFX_GAME_VERSION", "1")).To(Succeed())

			rootDir = testcommon.GetTestDataRootDirectory()
			Expect(rootDir).NotTo(BeEmpty(), "Project root directory should not be empty")

			binaryPath = "binary"
			temp = common.NewTempProvider("", "")
			gameVersionDir = "FFX"

			testDataPath = filepath.Join(rootDir, gameVersionDir)
			gameLocationPath = filepath.Join(testDataPath, binaryPath)

			file := `ffx_ps2\ffx\master\new_uspc\menu\macrodic.dcp`
			testFilePath = filepath.Join(gameLocationPath, file)
			Expect(common.CheckPathExists(testFilePath)).To(Succeed(), "Test file path should exist: %s", testFilePath)

			extractTempPath = filepath.Join(temp.TempFilePath, "extract")
			reimportTempPath = filepath.Join(temp.TempFilePath, "reimport")
			translatePath = filepath.Join(testDataPath, "translated")

			config = &interactions.FFXAppConfig{
				FFXGameVersion:    1,
				GameFilesLocation: gameLocationPath,
				ExtractLocation:   extractTempPath,
				TranslateLocation: translatePath,
				ImportLocation:    reimportTempPath,
			}

			formatter = &formatters.TxtFormatter{
				GameVersionDir: gameVersionDir,
				GameFilesPath:  translatePath,
			}

			dcpFileProperties = models.NewDcpFileOptions(1)

			interactions.NewInteractionServiceWithConfig(config)
			interactions.NewInteractionWithTextFormatter(formatter)

			destination = &locations.Destination{
				ExtractLocation:   locations.NewExtractLocation("extracted", extractTempPath, gameVersionDir),
				TranslateLocation: locations.NewTranslateLocation("translated", translatePath, gameVersionDir),
				ImportLocation:    locations.NewImportLocation("reimported", reimportTempPath, gameVersionDir),
			}

			log = testcommon.NewLogHandlerMock()
			Expect(log).NotTo(BeNil())

			// Inicializa os objetos de extração/compressão
			source, err := locations.NewSource(testFilePath)
			Expect(err).To(BeNil())
			Expect(source).NotTo(BeNil())

			Expect(destination.InitializeLocations(source, formatter)).To(Succeed())

			dcpFileReader = dcp.NewDcpFileExtractor(source, destination, formatter, log)
			Expect(dcpFileReader).NotTo(BeNil())

			dcpFileWriter = dcp.NewDcpFileCompressor(source, destination, formatter, log)
			Expect(dcpFileWriter).NotTo(BeNil())

			dcpFile = dcp.NewDcpFile(source, destination)
			Expect(dcpFile).NotTo(BeNil())

			verifyService = components.NewVerificationService()
			Expect(verifyService).NotTo(BeNil())
		})

		AfterEach(func() {
			Expect(common.RemoveDir(extractTempPath)).To(Succeed())
		})

		runCommonTests("expected 4, got 5", "expected 4, got 5")
	})
})
