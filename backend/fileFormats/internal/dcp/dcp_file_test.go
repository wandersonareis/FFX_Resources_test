package dcp_test

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/fileFormats/internal/dcp"
	"ffxresources/backend/fileFormats/internal/dcp/internal/integrity"
	"ffxresources/backend/fileFormats/internal/dcp/internal/lib"
	"ffxresources/backend/formatters"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
	"ffxresources/backend/logger"
	"fmt"

	"os"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("DcpFile", func() {
	var (
		interactionService *interactions.InteractionService
		extractTempPath    string
		reimportTempPath   string
		translatePath      string
		config             *interactions.FFXAppConfig
		formatter          interfaces.ITextFormatter
		fileOptions        core.IDcpFileOptions
		source             interfaces.ISource
		destination        locations.IDestination
		dcpFile            interfaces.IFileProcessor
		temp               *common.TempProvider
		gameVersionDir     string
		err                error
	)

	ginkgo.BeforeEach(func() {
		gomega.Expect(os.Setenv("APP_BASE_PATH", `F:\ffxWails\FFX_Resources\build\bin`)).To(gomega.Succeed())

		testPath := `F:\ffxWails\FFX_Resources\build\bin\data\ffx_ps2\ffx2\master\new_uspc\menu\macrodic.dcp`
		currentDir, err := filepath.Abs(".")
		gomega.Expect(err).To(gomega.BeNil())

		temp = common.NewTempProvider("", "")

		extractTempPath = filepath.Join(temp.TempFilePath, "extract")
		reimportTempPath = filepath.Join(temp.TempFilePath, "reimport")
		translatePath = filepath.Join(currentDir, "/testData/")

		gameVersionDir = "FFX-2"

		// Setup config
		config = &interactions.FFXAppConfig{
			FFXGameVersion:    2,
			GameFilesLocation: translatePath,
			ExtractLocation:   extractTempPath,
			TranslateLocation: translatePath,
			ImportLocation:    reimportTempPath,
		}

		// Initialize formatter
		formatter = &formatters.TxtFormatter{
			TargetExtension: ".txt",
			GameVersionDir:  gameVersionDir,
			GameFilesPath:   translatePath,
		}

		// Initialize file options
		fileOptions = core.NewDcpFileOptions(2)

		// Setup interaction service
		interactionService = interactions.NewInteractionServiceWithConfig(config)
		interactionService = interactions.NewInteractionWithTextFormatter(formatter)

		// Setup source and destination
		source, err = locations.NewSource(testPath)
		gomega.Expect(err).To(gomega.BeNil())

		destination = &locations.Destination{
			ExtractLocation: locations.NewExtractLocation("extracted", extractTempPath, gameVersionDir),
			TranslateLocation: locations.NewTranslateLocation("translated", translatePath, gameVersionDir),
			ImportLocation:    locations.NewImportLocation("reimported", reimportTempPath, gameVersionDir),
		}

		gomega.Expect(destination.InitializeLocations(source, formatter)).To(gomega.Succeed())

		dcpFile = dcp.NewDcpFile(source, destination)
	})

	ginkgo.AfterEach(func() {
		err = common.RemoveDir(extractTempPath)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("should have valid configuration", func() {
		gomega.Expect(config).NotTo(gomega.BeNil())
	})

	ginkgo.It("should have initialized interaction service", func() {
		gomega.Expect(interactionService).NotTo(gomega.BeNil())
		gomega.Expect(interactionService.FFXAppConfig()).NotTo(gomega.BeNil())
	})

	ginkgo.It("should have valid formatter", func() {
		gomega.Expect(formatter).NotTo(gomega.BeNil())
	})

	ginkgo.It("should have interaction service with text formatter", func() {
		gomega.Expect(interactionService.TextFormatter()).NotTo(gomega.BeNil())
	})

	ginkgo.It("should have valid destination", func() {
		gomega.Expect(destination).NotTo(gomega.BeNil())
	})

	ginkgo.It("should have correct extract location path", func() {
		expected := filepath.Join(extractTempPath, gameVersionDir, lib.DCP_PARTS_TARGET_DIR_NAME)
		expected = filepath.ToSlash(expected)

		actual := destination.Extract().GetTargetPath()
		actual = filepath.ToSlash(actual)

		gomega.Expect(actual).To(gomega.Equal(expected))
	})

	ginkgo.It("should extract the DCP file successfully", func() {
		gomega.Expect(dcpFile).NotTo(gomega.BeNil())
		gomega.Expect(dcpFile.Extract()).To(gomega.Succeed())
	})

	ginkgo.It("should have integrity text check fail", func() {
		gomega.Expect(dcpFile).NotTo(gomega.BeNil())
		gomega.Expect(dcpFile.Extract()).To(gomega.Succeed())

		dcpIntegrity := integrity.NewDcpFileExtractorIntegrity(logger.NewLoggerHandler("dcp_file_integrity_testing"))
		gomega.Expect(dcpIntegrity).NotTo(gomega.BeNil())

		targetPath := destination.Extract().GetTargetPath()
		gomega.Expect(targetPath).NotTo(gomega.BeEmpty())

		gomega.Expect(fileOptions).NotTo(gomega.BeNil())

		matchingFiles, err := filepath.Glob(filepath.Join(targetPath, "*.txt"))
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(matchingFiles).NotTo(gomega.BeEmpty())

		if len(matchingFiles) > 0 {
			randomFile := matchingFiles[1]
			fmt.Println(randomFile)
			err = os.Remove(randomFile)
			gomega.Expect(err).To(gomega.Succeed())
		}

		err = dcpIntegrity.Verify(targetPath, formatter, fileOptions)

		gomega.Expect(err).ToNot(gomega.Succeed())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("dcp file parts count mismatch: expected 6, got 7"))
	})

	ginkgo.It("should have integrity binary check fail", func() {
		gomega.Expect(dcpFile).NotTo(gomega.BeNil())
		gomega.Expect(dcpFile.Extract()).To(gomega.Succeed())

		dcpIntegrity := integrity.NewDcpFileExtractorIntegrity(logger.NewLoggerHandler("dcp_file_integrity_testing"))
		gomega.Expect(dcpIntegrity).NotTo(gomega.BeNil())

		targetPath := destination.Extract().GetTargetPath()
		gomega.Expect(targetPath).NotTo(gomega.BeEmpty())

		gomega.Expect(fileOptions).NotTo(gomega.BeNil())

		matchingFiles, err := filepath.Glob(filepath.Join(targetPath, "*.00?"))
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(matchingFiles).NotTo(gomega.BeEmpty())

		if len(matchingFiles) > 0 {
			randomFile := matchingFiles[1]
			fmt.Println(randomFile)
			err = os.Remove(randomFile)
			gomega.Expect(err).To(gomega.Succeed())
		}

		err = dcpIntegrity.Verify(targetPath, formatter, fileOptions)

		gomega.Expect(err).ToNot(gomega.Succeed())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("failed to ensure all DCP file binary parts"))
	})

	ginkgo.It("shoult have integrity check", func() {
		gomega.Expect(dcpFile).NotTo(gomega.BeNil())
		gomega.Expect(dcpFile.Extract()).To(gomega.Succeed())

		dcpIntegrity := integrity.NewDcpFileExtractorIntegrity(logger.NewLoggerHandler("dcp_file_integrity_testing"))
		gomega.Expect(dcpIntegrity).NotTo(gomega.BeNil())

		targetPath := destination.Extract().GetTargetPath()
		gomega.Expect(targetPath).NotTo(gomega.BeEmpty())

		gomega.Expect(fileOptions).NotTo(gomega.BeNil())

		gomega.Expect(dcpIntegrity.Verify(targetPath, formatter, fileOptions)).To(gomega.Succeed())
	})
})
