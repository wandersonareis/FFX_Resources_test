package interactions

import (
	"bytes"
	"encoding/json"
	"ffxresources/backend/common"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type IGetAppConfig interface {
	GetGameVersion() common.GameVersion
	GetLocations() map[string]string
	GetLocation(name string) string
	GetEnableMods() bool
}

type ISetAppConfig interface {
	SetGameVersion(version common.GameVersion)
	SetLocation(name, path string)
	SetEnableMods(enabled bool)
}

type IAppConfig interface {
	IGetAppConfig
	ISetAppConfig
	FromJson() error
	ToJson() error
}

type AppConfig struct {
	filePath    string
	locations   map[string]string
	gameVersion common.GameVersion
	enableMods  bool
}

type appConfigJSON struct {
	Locations   map[string]string  `json:"Locations"`
	GameVersion common.GameVersion `json:"GameVersion"`
	EnableMods  *bool              `json:"EnableMods,omitempty"`
}

func NewAppConfig() *AppConfig {
	filePath := filepath.Join(common.GetExecDir(), "config", "config.json")

	c := &AppConfig{
		filePath:    filePath,
		locations:   defaultLocations(),
		gameVersion: common.GameVersionFFX,
		enableMods:  true, // mods-first é o comportamento padrão
	}

	if err := c.FromJson(); err != nil {
		common.LogVerbose("Using default config (%v)", err)
	}

	common.SetModsEnabled(c.enableMods)

	if err := c.validateConfig(); err != nil {
		common.LogVerbose("Could not persist default config (%v)", err)
	}

	return c
}

// defaultLocations monta os diretórios padrão a partir do diretório do
// executável e dos nomes em common. GameFiles = <exec>/data;
// Translate é derivado: <gamefiles>/mods/translated. Extract/Import
// são internos (ocultos do frontend) mas mantidos no config.
func defaultLocations() map[string]string {
	execDir := common.GetExecDir()
	gameDir := filepath.Join(execDir, common.DirData)
	return map[string]string{
		"GameFilesLocation": gameDir,
		"ExtractLocation":   filepath.Join(execDir, common.DirExtracted),
		"TranslateLocation": common.DefaultTranslatedDir(gameDir),
		"ImportLocation":    filepath.Join(execDir, common.DirReimported),
	}
}

// ensureTranslateDerived preenche TranslateLocation derivando do
// GameFilesLocation atual quando a chave está ausente/vazia.
// Valores customizados do usuário nunca são sobrescritos.
func ensureTranslateDerived(locations map[string]string) {
	if strings.TrimSpace(locations["TranslateLocation"]) != "" {
		return
	}
	gameDir := strings.TrimSpace(locations["GameFilesLocation"])
	if gameDir == "" {
		gameDir = filepath.Join(common.GetExecDir(), common.DirData)
	}
	locations["TranslateLocation"] = common.DefaultTranslatedDir(gameDir)
}

func (c *AppConfig) validateConfig() error {
	changed := false

	if c.locations == nil {
		c.locations = defaultLocations()
		changed = true
	} else {
		// Preenche chaves ausentes sem sobrescrever as definidas pelo usuário.
		for key, defPath := range defaultLocations() {
			if key == "TranslateLocation" {
				continue
			}
			if strings.TrimSpace(c.locations[key]) == "" {
				c.locations[key] = defPath
				changed = true
			}
		}
		// Translate deriva do gamefiles atual, não de constante fixa.
		before := strings.TrimSpace(c.locations["TranslateLocation"])
		ensureTranslateDerived(c.locations)
		if strings.TrimSpace(c.locations["TranslateLocation"]) != before {
			changed = true
		}
	}

	// Garante que os diretórios padrão existam.
	for _, dir := range c.locations {
		if strings.TrimSpace(dir) == "" {
			continue
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			common.LogVerbose("Could not create directory %s (%v)", dir, err)
		}
	}

	if changed {
		return c.ToJson()
	}

	return nil
}

func (c *AppConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(appConfigJSON{
		Locations:   c.locations,
		GameVersion: c.gameVersion,
		EnableMods:  &c.enableMods,
	})
}

func (c *AppConfig) UnmarshalJSON(data []byte) error {
	var aux appConfigJSON
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	// Mescla sobre os defaults: chaves ausentes/vazias no arquivo
	// mantêm os diretórios padrão em vez de zerarem a config.
	// Translate vazio deriva do gamefiles (<game>/mods/translated).
	merged := defaultLocations()
	// Remove o translate pré-derivado do default para que ele seja
	// recalculado a partir do GameFilesLocation efetivo (arquivo ou default).
	delete(merged, "TranslateLocation")
	for key, path := range aux.Locations {
		if key == "TranslateLocation" {
			continue
		}
		if strings.TrimSpace(path) != "" {
			merged[key] = path
		}
	}
	if p := strings.TrimSpace(aux.Locations["TranslateLocation"]); p != "" {
		merged["TranslateLocation"] = p
	} else {
		ensureTranslateDerived(merged)
	}
	c.locations = merged
	c.gameVersion = aux.GameVersion
	// EnableMods ausente no arquivo → mantém o default (true).
	if aux.EnableMods != nil {
		c.enableMods = *aux.EnableMods
	}
	return nil
}

func (c *AppConfig) ToJson() error {
	if c == nil {
		return fmt.Errorf("invalid configuration")
	}

	if err := os.MkdirAll(filepath.Dir(c.filePath), 0755); err != nil {
		return err
	}

	file, err := os.Create(c.filePath)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			common.LogVerbose("Error closing config file: %v", err)
		}
	}(file)

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(c)
	if err != nil {
		return err
	}

	return nil
}

func (c *AppConfig) FromJson() error {
	file, err := os.ReadFile(c.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return c.validateConfig()
		}
		return err
	}
	defer func() { file = file[:0] }()

	if len(bytes.TrimSpace(file)) == 0 {
		return c.validateConfig()
	}

	if err := json.Unmarshal(file, c); err != nil {
		return err
	}

	return c.validateConfig()
}

func (c *AppConfig) GetGameVersion() common.GameVersion {
	return c.gameVersion
}

func (c *AppConfig) SetGameVersion(version common.GameVersion) {
	c.gameVersion = version
	common.SetCurrentGameVersion(c.gameVersion)

	_ = c.ToJson()
}

func (c *AppConfig) GetLocations() map[string]string {
	return c.locations
}

func (c *AppConfig) GetLocation(name string) string {
	return c.locations[name]
}

func (c *AppConfig) SetLocation(name, path string) {
	if c.locations == nil {
		c.locations = make(map[string]string)
	}
	c.locations[name] = path
	_ = c.ToJson()
}

// GetEnableMods devolve a preferência mods-first (config EnableMods).
func (c *AppConfig) GetEnableMods() bool {
	return c.enableMods
}

// SetEnableMods persiste o toggle e aplica na resolução de caminhos.
func (c *AppConfig) SetEnableMods(enabled bool) {
	c.enableMods = enabled
	common.SetModsEnabled(enabled)
	_ = c.ToJson()
}
