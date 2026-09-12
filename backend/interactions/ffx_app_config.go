package interactions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"ffxresources/backend/common"
	"os"
	"path/filepath"
	"strings"
)

type IGetAppConfig interface {
	GetGameVersion() common.GameVersion
	GetLocations() map[string]string
	GetLocation(name string) string
}

type ISetAppConfig interface {
	SetGameVersion(version common.GameVersion)
	SetLocation(name, path string)
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
}

type appConfigJSON struct {
	Locations   map[string]string `json:"Locations"`
	GameVersion common.GameVersion `json:"GameVersion"`
}

func NewAppConfig() *AppConfig {
	filePath := filepath.Join(common.GetExecDir(), "config", "config.json")

	c := &AppConfig{
		filePath:    filePath,
		locations:   defaultLocations(),
		gameVersion: common.GameVersionFFX,
	}

	// FromJson mescla o arquivo sobre os defaults; em qualquer falha
	// mantemos o config padrão válido em vez de retornar nil.
	if err := c.FromJson(); err != nil {
		common.LogVerbose("Using default config (%v)", err)
	}

	if err := c.validateConfig(); err != nil {
		common.LogVerbose("Could not persist default config (%v)", err)
	}

	return c
}

// defaultLocations monta os diretórios padrão a partir do diretório do
// executável e dos nomes em common. Usado quando não há config.json
// ou quando chaves estão ausentes no arquivo.
func defaultLocations() map[string]string {
	execDir := common.GetExecDir()
	return map[string]string{
		"GameFilesLocation": filepath.Join(execDir, common.DirData),
		"ExtractLocation":   filepath.Join(execDir, common.DirExtracted),
		"TranslateLocation": filepath.Join(execDir, common.DirTranslated),
		"ImportLocation":    filepath.Join(execDir, common.DirReimported),
	}
}

func (c *AppConfig) validateConfig() error {
	changed := false

	if !c.gameVersion.Normalize().IsValid() {
		c.gameVersion = common.GameVersionFFX
		changed = true
	}

	if c.locations == nil {
		c.locations = defaultLocations()
		changed = true
	} else {
		// Preenche chaves ausentes sem sobrescrever as definidas pelo usuário.
		for key, defPath := range defaultLocations() {
			if strings.TrimSpace(c.locations[key]) == "" {
				c.locations[key] = defPath
				changed = true
			}
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
		GameVersion: c.gameVersion.Normalize(),
	})
}

func (c *AppConfig) UnmarshalJSON(data []byte) error {
	var aux appConfigJSON
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	// Mescla sobre os defaults: chaves ausentes/vazias no arquivo
	// mantêm os diretórios padrão em vez de zerarem a config.
	merged := defaultLocations()
	for key, path := range aux.Locations {
		if strings.TrimSpace(path) != "" {
			merged[key] = path
		}
	}
	c.locations = merged
	c.gameVersion = aux.GameVersion.Normalize()
	if !c.gameVersion.IsValid() {
		c.gameVersion = common.GameVersionFFX
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
	return c.gameVersion.Normalize()
}

func (c *AppConfig) SetGameVersion(version common.GameVersion) {
	c.gameVersion = version.Normalize()
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
