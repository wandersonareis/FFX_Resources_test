package objectsfile_test

import (
	"crypto/sha256"
	"ffxresources/backend/common"
	"ffxresources/backend/core/reader"
	"ffxresources/backend/datastore"
	"ffxresources/backend/fileFormats/objectsfile"
	"ffxresources/backend/interactions"
	testcommon "ffxresources/testData"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBinaryFile(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BinaryFile Integrity Suite")
}

type integrityCase struct {
	pattern string
	read    func(string) datastore.IBinaryFile
}

// Arquivos exercitados por examples/main.go (runFFXExamples).
var ffxIntegrityCases = []integrityCase{
	{pattern: "battle/kernel/important.bin", read: objectsfile.ReadCommandLocalizations},
	{pattern: "battle/kernel/command.bin", read: objectsfile.ReadCommandLocalizations},
	{pattern: "battle/kernel/a_ability.bin", read: objectsfile.ReadCommandLocalizations},
	{pattern: "battle/kernel/item.bin", read: objectsfile.ReadCommandLocalizations},
	{pattern: "battle/kernel/arms_txt.bin", read: objectsfile.ReadCommandLocalizations},
	{pattern: "battle/kernel/config_txt.bin", read: objectsfile.ReadCommandLocalizations},
	{pattern: "battle/kernel/item_txt.bin", read: objectsfile.ReadCommandLocalizations},
	{pattern: "battle/kernel/mmain_txt.bin", read: objectsfile.ReadCommandLocalizations},
	{pattern: "battle/kernel/panel.bin", read: objectsfile.ReadCommandLocalizations},
	{pattern: "battle/kernel/status_txt.bin", read: objectsfile.ReadCommandLocalizations},
	{pattern: "battle/kernel/summon_txt.bin", read: objectsfile.ReadCommandLocalizations},
	{pattern: "battle/kernel/btl_txt.bin", read: objectsfile.ReadNameOnlyLocalizations},
	{pattern: "battle/kernel/btlend_txt.bin", read: objectsfile.ReadNameOnlyLocalizations},
	{pattern: "battle/kernel/monmagic1.bin", read: objectsfile.ReadNameOnlyLocalizations},
	{pattern: "battle/kernel/monmagic2.bin", read: objectsfile.ReadNameOnlyLocalizations},
	{pattern: "battle/kernel/build_txt.bin", read: objectsfile.ReadNameOnlyLocalizations},
	{pattern: "battle/kernel/ply_rom.bin", read: objectsfile.ReadCommandLocalizations},
	{pattern: "battle/kernel/ply_save.bin", read: objectsfile.ReadNameOnlyV2Localizations},
	{pattern: "battle/kernel/sphere.bin", read: objectsfile.ReadNameOnlyLocalizations},
	{pattern: "battle/kernel/save_txt.bin", read: objectsfile.ReadNameOnlyLocalizations},
	{pattern: "battle/kernel/name_txt.bin", read: objectsfile.ReadNameOnlyLocalizations},
	{pattern: "battle/kernel/monster1.bin", read: objectsfile.ReadNameSensorScanLocalizations},
	{pattern: "battle/kernel/monster2.bin", read: objectsfile.ReadNameSensorScanLocalizations},
	{pattern: "battle/kernel/monster3.bin", read: objectsfile.ReadNameSensorScanLocalizations},
	{pattern: "battle/kernel/w_name.bin", read: objectsfile.ReadWeaponNamesLocalizations},
}

var _ = Describe("BinaryFile Integrity", Ordered, func() {
	var (
		tmpRootFFX        string
		tmpRootFFX2       string
		originalGameFiles string
		originalResources string
		creator           func([]byte, []byte, int, string) (datastore.IGlobalLocalizedTextObject, error)
	)

	setupVersion := func(version int, srcTree string) string {
		common.SetGameVersion(version)

		tmpRoot, err := os.MkdirTemp("", "binaryfile_integrity")
		Expect(err).ToNot(HaveOccurred())

		gameDir := filepath.Join(tmpRoot, "game")
		Expect(os.CopyFS(gameDir, os.DirFS(srcTree))).To(Succeed())

		// Directories are swapped exclusively through interactions.
		config := interactions.NewAppConfig()
		Expect(config).NotTo(BeNil())
		config.SetGameVersion(version)
		config.SetLocation("GameFilesLocation", gameDir)
		config.SetLocation("ExtractLocation", filepath.Join(tmpRoot, "extracted"))
		config.SetLocation("TranslateLocation", filepath.Join(tmpRoot, "translated"))
		config.SetLocation("ImportLocation", filepath.Join(tmpRoot, "reimported"))

		interactions.NewInteractionServiceWithConfig(config)
		Expect(common.GameFilesRoot).To(Equal(gameDir))

		Expect(reader.InitializeInternals()).To(Succeed())

		return tmpRoot
	}

	// Ciclo de integridade espelhando examples/main.go:
	// Read -> ExportToJson -> ImportFromJson -> SaveToBinary -> hash.
	integrityCycle := func(tmpRoot string, tc integrityCase) {
		binFile := tc.read(tc.pattern)
		Expect(binFile).NotTo(BeNil())
		Expect(binFile.GetObjects()).NotTo(BeNil())
		Expect(binFile.GetObjects().Len()).To(BeNumerically(">", 0))

		jsonName := strings.ReplaceAll(strings.ReplaceAll(tc.pattern, "/", "_"), ".bin", "_integrity.json")
		Expect(binFile.ExportToJson(jsonName)).To(Succeed())
		Expect(binFile.ImportFromJson(jsonName)).To(Succeed())

		outPath := filepath.Join(tmpRoot, "reimported", strings.ReplaceAll(tc.pattern, "/", "_"))
		Expect(binFile.SaveToBinary(outPath)).To(Succeed())

		info, err := os.Stat(outPath)
		Expect(err).ToNot(HaveOccurred())
		Expect(info.Size()).To(BeNumerically(">", 0))

		// gameDir é resolvido através do interaction service.
		gameDir := interactions.NewInteractionService().GameLocation.GetTargetDirectory()
		rel := filepath.Join(common.GetLocalizationRoot(common.DefaultLocalization), tc.pattern)
		origBytes, err := os.ReadFile(filepath.Join(gameDir, rel))
		Expect(err).ToNot(HaveOccurred())
		outBytes, err := os.ReadFile(outPath)
		Expect(err).ToNot(HaveOccurred())

		Expect(sha256.Sum256(outBytes)).To(Equal(sha256.Sum256(origBytes)))
	}

	BeforeAll(func() {
		Expect(testcommon.SetBuildBinPath()).To(Succeed())
		common.SetVerboseMode(false)

		originalGameFiles = common.GameFilesRoot
		originalResources = common.ResourcesRoot

		creator = func(data []byte, stringBytes []byte, headerLength int, localization string) (datastore.IGlobalLocalizedTextObject, error) {
			gameVersion := interactions.NewInteractionService().FFXAppConfig().GetGameVersion()
			return objectsfile.NewNameOnlyTextObject(data, stringBytes, headerLength, localization, gameVersion)
		}
	})

	AfterAll(func() {
		common.GameFilesRoot = originalGameFiles
		common.ResourcesRoot = originalResources
		if tmpRootFFX != "" {
			os.RemoveAll(tmpRootFFX)
		}
		if tmpRootFFX2 != "" {
			os.RemoveAll(tmpRootFFX2)
		}
	})

	Context("FFX (v1) - examples/main.go runFFXExamples", func() {
		BeforeAll(func() {
			srcTree := filepath.Join(testcommon.GetTestDataRootDirectory(), "FFX", "binary")
			tmpRootFFX = setupVersion(1, srcTree)
		})

		for _, tc := range ffxIntegrityCases {
			tc := tc
			It("should complete integrity cycle with matching hash for "+tc.pattern, func() {
				integrityCycle(tmpRootFFX, tc)
			})
		}
	})

	Context("FFX-2 (v2) - examples/main.go runFFX2Examples", func() {
		BeforeAll(func() {
			srcTree := filepath.Join(testcommon.GetTestDataRootDirectory(), "FFX-2", "binary")
			tmpRootFFX2 = setupVersion(2, srcTree)
		})

		// testData/FFX-2/binary contém apenas battle/kernel/a_ability.bin;
		// os demais arquivos do runFFX2Examples não existem nos dados de teste.
		It("should complete integrity cycle with matching hash for battle/kernel/a_ability.bin", func() {
			integrityCycle(tmpRootFFX2, integrityCase{pattern: "battle/kernel/a_ability.bin", read: objectsfile.ReadCommandLocalizations})
		})
	})

	Context("Access errors", func() {
		It("should fail LoadFromBinary when file does not exist", func() {
			version := interactions.NewInteractionService().FFXAppConfig().GetGameVersion()
			binFile := objectsfile.NewBinaryFile("battle/kernel/does_not_exist.bin", creator, common.DefaultLocalization, version)
			err := binFile.LoadFromBinary()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("file does not exist"))
		})

		It("should fail LoadFromBinary when path is a directory", func() {
			gameDir := interactions.NewInteractionService().GameLocation.GetTargetDirectory()
			dirAsFile := filepath.Join(gameDir, common.GetLocalizationRoot(common.DefaultLocalization), "battle", "kernel", "dir_as_file.bin")
			Expect(os.MkdirAll(dirAsFile, 0755)).To(Succeed())

			binFile := objectsfile.NewBinaryFile("battle/kernel/dir_as_file.bin", creator, common.DefaultLocalization, interactions.NewInteractionService().FFXAppConfig().GetGameVersion())
			err := binFile.LoadFromBinary()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to read file"))
		})

		It("should fail ExportToJson when file path is empty", func() {
			binFile := objectsfile.NewBinaryFile("battle/kernel/command.bin", creator, common.DefaultLocalization, interactions.NewInteractionService().FFXAppConfig().GetGameVersion())
			err := binFile.ExportToJson("")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("json file not configured"))
		})

		It("should fail ImportFromJson when file path is empty", func() {
			binFile := objectsfile.NewBinaryFile("battle/kernel/command.bin", creator, common.DefaultLocalization, interactions.NewInteractionService().FFXAppConfig().GetGameVersion())
			err := binFile.ImportFromJson("")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("json file not configured"))
		})

		It("should fail ImportFromJson when json file does not exist", func() {
			binFile := objectsfile.NewBinaryFile("battle/kernel/command.bin", creator, common.DefaultLocalization, interactions.NewInteractionService().FFXAppConfig().GetGameVersion())
			err := binFile.ImportFromJson("binary_integrity_missing.json")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("JSON file not found"))
		})
	})
})
