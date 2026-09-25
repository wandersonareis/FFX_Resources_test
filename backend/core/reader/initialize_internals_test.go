package reader_test

import (
	"path/filepath"

	"ffxresources/backend/common"
	"ffxresources/backend/core/converter"
	"ffxresources/backend/core/encoding"
	"ffxresources/backend/core/reader"
	"ffxresources/backend/datastore"
	testcommon "ffxresources/testData"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Specs de inicialização: existência dos mapas de charset + round-trip de
// codificação string → bytes → string com tags, usando os dados de teste do
// repositório. Rodam no suite "ReadManager" (runner em read_manager_test.go).
// A versão é passada explícita a InitializeInternals: sem versão global nem
// singleton de interactions no caminho.
var _ = Describe("InitializeInternals", Ordered, func() {
	var rootDir string
	var originalResourcesRoot string

	BeforeAll(func() {
		Expect(testcommon.SetBuildBinPath()).To(Succeed())
		common.SetVerboseMode(false)
	})

	BeforeEach(func() {
		originalResourcesRoot = common.ResourcesRoot

		rootDir = testcommon.GetTestDataRootDirectory()
		Expect(rootDir).NotTo(BeEmpty(), "Project root directory should not be empty")
		common.SetGameFilesRoot(filepath.Join(rootDir, "FFX", "binary"))
	})

	AfterEach(func() {
		common.ResourcesRoot = originalResourcesRoot
		ffxencoding.ClearAllCharMaps()
		datastore.Instance.ClearAllMacros()
	})

	It("inicializa os charsets e publica os macros da versão indicada", func() {
		Expect(reader.InitializeInternals(common.GameVersionFFX)).To(Succeed())

		for _, charset := range common.Charsets {
			Expect(ffxencoding.GetByteToCharMap(common.GameVersionFFX, charset)).ToNot(BeEmpty(),
				"ByteToChar map should exist for charset %s", charset)
			Expect(ffxencoding.GetCharToByteMap(common.GameVersionFFX, charset)).ToNot(BeEmpty(),
				"CharToByte map should exist for charset %s", charset)
		}
		Expect(datastore.GetMacros(common.GameVersionFFX).Count()).To(BeNumerically(">", 0),
			"Datastore macros should be populated")
	})

	It("faz round-trip de textos com tags: string → bytes → string", func() {
		Expect(reader.InitializeInternals(common.GameVersionFFX)).To(Succeed())

		texts := []string{
			"Hello{PAUSE}World{TEXT_NEWLINE}Line2",
			"{TEXT_ITALIC}Italic{TEXT_NORMAL}Normal",
			"{CHOICE:00}Yes{CHOICE:01}No{CHOICE-END}",
			"Cid's Ãã",
		}
		for _, text := range texts {
			encoded, err := converter.StringToStoredBytes(text, "us", common.GameVersionFFX)
			Expect(err).ToNot(HaveOccurred(), "encode falhou para %q", text)

			decoded := converter.BytesToString(encoded, "us", common.GameVersionFFX)
			Expect(decoded).To(Equal(text), "round-trip falhou para %q", text)
		}
	})

	It("codifica com os slots reais da tabela us", func() {
		Expect(reader.InitializeInternals(common.GameVersionFFX)).To(Succeed())

		encoded, err := converter.StringToBytes("A{PAUSE}B", "us", common.GameVersionFFX)
		Expect(err).ToNot(HaveOccurred())
		Expect(encoded).To(Equal([]byte{0x50, 0x01, 0x51}))
	})
})
