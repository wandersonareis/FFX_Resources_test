package helpfile

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/datastore"
)

// LocalizedHelpStringObject é um segmento (por índice de ponteiro) com
// conteúdo por localização: cada localização carrega o HelpSegment do seu
// próprio arquivo (os textos divergem entre localizações, e os blocos de
// dados também). Espelha LocalizedFieldStringObject do event.
type LocalizedHelpStringObject struct {
	Index    int
	Contents map[string]*HelpSegment
}

func NewLocalizedHelpStringObject(index int) *LocalizedHelpStringObject {
	return &LocalizedHelpStringObject{
		Index:    index,
		Contents: make(map[string]*HelpSegment),
	}
}

func (o *LocalizedHelpStringObject) SetLocalizedContent(localization string, seg *HelpSegment) {
	if localization == "" || seg == nil {
		return
	}
	o.Contents[localization] = seg
}

func (o *LocalizedHelpStringObject) GetLocalizedContent(localization string) *HelpSegment {
	return o.Contents[localization]
}

func (o *LocalizedHelpStringObject) GetLocalizedString(localization string) string {
	if seg := o.GetLocalizedContent(localization); seg != nil {
		return seg.Text
	}
	return ""
}

func (o *LocalizedHelpStringObject) GetDefaultContent() *HelpSegment {
	return o.GetLocalizedContent(common.DefaultLocalization)
}

func (o *LocalizedHelpStringObject) String() string {
	if seg := o.GetDefaultContent(); seg != nil {
		return seg.Text
	}
	return ""
}

// HelpKeyedStringFile é um painel de ajuda (um .sps2) no ciclo de vida do
// datastore. Implementa IGlobalLocalizedTextObject (mesma forma do
// EventKeyedStringFile): o arquivo de cada localização guarda o esqueleto
// estrutural (header, sub-header, bloco de páginas, footer) e os segmentos
// daquela localização.
type HelpKeyedStringFile struct {
	Name  string
	Files map[string]*HelpBinaryFile
}

func NewHelpKeyedStringFile(name string) *HelpKeyedStringFile {
	return &HelpKeyedStringFile{
		Name:  name,
		Files: make(map[string]*HelpBinaryFile),
	}
}

// SegmentCount devolve a contagem de segmentos (igual entre localizações).
func (h *HelpKeyedStringFile) SegmentCount() int {
	for _, f := range h.Files {
		return len(f.Segments)
	}
	return 0
}

// LocalizedStrings devolve os segmentos fundidos por índice (um objeto por
// ponteiro, com conteúdo por localização).
func (h *HelpKeyedStringFile) LocalizedStrings() []*LocalizedHelpStringObject {
	count := h.SegmentCount()
	out := make([]*LocalizedHelpStringObject, count)
	for i := range out {
		out[i] = NewLocalizedHelpStringObject(i)
	}
	for loc, f := range h.Files {
		for i, seg := range f.Segments {
			if i < count {
				out[i].SetLocalizedContent(loc, seg)
			}
		}
	}
	return out
}

func (h *HelpKeyedStringFile) GetName(string) string { return "" }

func (h *HelpKeyedStringFile) GetKeyedString(string) datastore.IGlobalLocalizedKeyedStringObject {
	return nil
}

func (h *HelpKeyedStringFile) GetLocalizedKeyedStrings(string) []datastore.IGlobalKeyedString {
	return nil
}

func (h *HelpKeyedStringFile) SetLocalizations(datastore.IGlobalLocalizationSetter) {}

func (h *HelpKeyedStringFile) GetTextObject() datastore.IGlobalLocalizedTextObject {
	return h
}

func (h *HelpKeyedStringFile) GetHeaderLength() int {
	return h.SegmentCount() * 4
}

// Save grava os arquivos das localizações com segmentos alterados na árvore
// mods/ (mods/<localizationRoot>/help/<dir>/<name>.sps2, padrão dos demais
// formatos). Sem alteração, nenhum arquivo é tocado.
func (h *HelpKeyedStringFile) Save() error {
	for loc, f := range h.Files {
		if !f.IsDirty() {
			continue
		}
		if err := f.Save(); err != nil {
			return fmt.Errorf("helpfile: falha ao salvar %s/%s: %w", loc, h.Name, err)
		}
	}
	return nil
}

// ToBytes reconstrói o binário do painel para a localização pedida.
func (h *HelpKeyedStringFile) ToBytes(languageCode string) ([]byte, error) {
	f := h.Files[languageCode]
	if f == nil {
		return nil, fmt.Errorf("helpfile: %s sem arquivo para %s", h.Name, languageCode)
	}
	return f.Rebuild()
}

func (h *HelpKeyedStringFile) ToString(languageCode string) string {
	f := h.Files[languageCode]
	if f == nil {
		return ""
	}
	parts := make([]string, 0, len(f.Segments))
	for _, seg := range f.Segments {
		if seg.Text != "" {
			parts = append(parts, seg.Text)
		}
	}
	return strings.Join(parts, "\n")
}

func (h *HelpKeyedStringFile) String() string {
	return h.ToString(common.DefaultLocalization)
}

// sortedSupportedLocalizations devolve as localizações suportadas com o
// idioma default primeiro (ordem determinística de carga).
func sortedSupportedLocalizations() []string {
	keys := make([]string, 0, len(common.SupportedLanguages))
	for loc := range common.SupportedLanguages {
		keys = append(keys, loc)
	}
	sort.Strings(keys)
	out := make([]string, 0, len(keys)+1)
	out = append(out, common.DefaultLocalization)
	for _, k := range keys {
		if k != common.DefaultLocalization {
			out = append(out, k)
		}
	}
	return out
}

// HelpStringsBinaryFile implementa datastore.IBinaryFile para o kind help:
// carrega todos os painéis (FFX-only — as árvores ffx2/lastmiss não têm a
// pasta help/) e salva de volta os que têm segmentos alterados.
type HelpStringsBinaryFile struct {
	Version common.GameVersion
	Objects components.IList[datastore.IGlobalLocalizedTextObject]
}

func NewHelpStringsBinaryFile(version common.GameVersion) *HelpStringsBinaryFile {
	return &HelpStringsBinaryFile{
		Version: version,
		Objects: components.NewEmptyList[datastore.IGlobalLocalizedTextObject](),
	}
}

// LoadFromBinary carrega os painéis para todas as localizações que têm o
// arquivo (resolução mods-first). Localizações sem o arquivo são puladas sem
// erro — a régua de versões fica no chamador (ListEntries). A ordem dos
// objetos segue o registro (HelpEntries), determinística.
func (b *HelpStringsBinaryFile) LoadFromBinary() error {
	if err := ffxencoding.PrepareVersionCharsets(b.Version); err != nil {
		return fmt.Errorf("helpfile: charset maps não carregados: %w", err)
	}

	loaded := make(map[string]*HelpKeyedStringFile)

	for _, loc := range sortedSupportedLocalizations() {
		for _, entry := range HelpEntries {
			f, err := ReadHelpFile(b.Version, loc, entry.Name)
			if err != nil {
				// Arquivo ausente é esperado (ex.: ffx2); demais falhas logam.
				common.LogVerbose("helpfile: %s/%s indisponível: %v", loc, entry.Name, err)
				continue
			}
			obj, ok := loaded[entry.Name]
			if !ok {
				obj = NewHelpKeyedStringFile(entry.Name)
				loaded[entry.Name] = obj
			}
			obj.Files[loc] = f
		}
	}

	for _, entry := range HelpEntries {
		if obj := loaded[entry.Name]; obj != nil {
			b.Objects.Add(obj)
		}
	}

	common.LogInfo("helpfile: painéis carregados: %d", b.Objects.Len())
	return nil
}

// SaveToBinary reconstrói e grava os painéis na árvore mods/. filePath vazio =
// todos; senão, só o painel cujo id (stem) casar. Apenas localizações com
// segmentos alterados são regravadas (o original do gamefiles permanece
// intacto).
func (b *HelpStringsBinaryFile) SaveToBinary(filePath string) error {
	target := ""
	if filePath != "" {
		target = trimExt(filepath.Base(filePath))
	}
	for i := 0; i < b.Objects.Len(); i++ {
		obj, ok := b.Objects.Get(i).(*HelpKeyedStringFile)
		if !ok {
			continue
		}
		if target != "" && !strings.EqualFold(obj.Name, target) {
			continue
		}
		if err := obj.Save(); err != nil {
			return err
		}
	}
	return nil
}

func trimExt(name string) string {
	return strings.TrimSuffix(name, filepath.Ext(name))
}

func (b *HelpStringsBinaryFile) GetObjects() components.IList[datastore.IGlobalLocalizedTextObject] {
	if b.Objects == nil {
		b.Objects = components.NewEmptyList[datastore.IGlobalLocalizedTextObject]()
	}
	return b.Objects
}

// ---- store (padrão event.SetEvent) ------------------------------------------

var (
	helpStoreMu sync.Mutex
	helpStore   = map[common.GameVersion]map[string]*HelpKeyedStringFile{}
)

// GetHelp devolve o painel carregado para a versão (nil se ausente).
func GetHelp(version common.GameVersion, name string) *HelpKeyedStringFile {
	helpStoreMu.Lock()
	defer helpStoreMu.Unlock()
	if byName, ok := helpStore[version]; ok {
		return byName[name]
	}
	return nil
}

// SetHelp registra o painel no store da versão.
func SetHelp(version common.GameVersion, file *HelpKeyedStringFile) {
	if file == nil {
		return
	}
	helpStoreMu.Lock()
	defer helpStoreMu.Unlock()
	byName, ok := helpStore[version]
	if !ok {
		byName = make(map[string]*HelpKeyedStringFile)
		helpStore[version] = byName
	}
	byName[file.Name] = file
}

// HasHelp informa se a versão tem painéis carregados.
func HasHelp(version common.GameVersion) bool {
	helpStoreMu.Lock()
	defer helpStoreMu.Unlock()
	return len(helpStore[version]) > 0
}

// ClearHelp limpa o store da versão (higiene de testes).
func ClearHelp(version common.GameVersion) {
	helpStoreMu.Lock()
	defer helpStoreMu.Unlock()
	delete(helpStore, version)
}

// ClearAllHelp limpa o store de todas as versões (troca de GameFilesLocation:
// os painéis carregados pertencem à árvore anterior).
func ClearAllHelp() {
	helpStoreMu.Lock()
	defer helpStoreMu.Unlock()
	helpStore = map[common.GameVersion]map[string]*HelpKeyedStringFile{}
}

// EnsureHelpLoaded garante o store da versão (carga única, lazily).
func EnsureHelpLoaded(version common.GameVersion) error {
	if HasHelp(version) {
		return nil
	}
	binFile := NewHelpStringsBinaryFile(version)
	if err := binFile.LoadFromBinary(); err != nil {
		return fmt.Errorf("helpfile: falha ao carregar painéis de %s: %w", version, err)
	}
	names := make([]string, 0, binFile.GetObjects().Len())
	for i := 0; i < binFile.GetObjects().Len(); i++ {
		if obj, ok := binFile.GetObjects().Get(i).(*HelpKeyedStringFile); ok {
			SetHelp(version, obj)
			names = append(names, obj.Name)
		}
	}
	// Vigia de frescura: carimbo físico de cada painel — binário modificado
	// no disco (tradução copiada em mods/) recarrega na próxima leitura
	// (EnsureHelpFresh).
	StampAllPanels(version, names)
	return nil
}
