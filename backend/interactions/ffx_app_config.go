package interactions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"ffxresources/backend/common"
	"os"
	"path/filepath"
)

type IGetAppConfig interface {
	GetGameVersion() int
	GetLocations() map[string]string
	GetLocation(name string) string
}

type ISetAppConfig interface {
	SetGameVersion(version int)
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
	gameVersion int
}

type appConfigJSON struct {
	Locations   map[string]string `json:"Locations"`
	GameVersion int               `json:"GameVersion"`
}

func NewAppConfig() *AppConfig {
	filePath := filepath.Join(common.GetExecDir(), "config", "config.json")

	c := &AppConfig{
		filePath:    filePath,
		locations:   make(map[string]string),
		gameVersion: 1,
	}

	if err := c.FromJson(); err != nil {
		return nil
	}

	return c
}

func (c *AppConfig) validateConfig() error {
	changed := false

	if c.gameVersion <= 0 {
		c.gameVersion = 1
		changed = true
	}

	if c.locations == nil {
		c.locations = make(map[string]string)
		changed = true
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
	})
}

func (c *AppConfig) UnmarshalJSON(data []byte) error {
	var aux appConfigJSON
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	c.locations = aux.Locations
	c.gameVersion = aux.GameVersion
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
		err := file.Close()
		if err != nil {
			panic(err)
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

	err = json.Unmarshal(file, c)

	return err
}

func (c *AppConfig) GetGameVersion() int {
	return c.gameVersion
}

func (c *AppConfig) SetGameVersion(version int) {
	c.gameVersion = version
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
