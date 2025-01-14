package interactions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

type (
	ConfigField string

	FFXAppConfig struct {
		filePath          string
		FFXGameVersion    int    `json:"FFXGameVersion"`
		GameFilesLocation string `json:"GameFilesLocation"`
		ExtractLocation   string `json:"ExtractLocation"`
		TranslateLocation string `json:"TranslateLocation"`
		ImportLocation    string `json:"ImportLocation"`
	}

	IFFXAppConfig interface {
		ToJson() error
		FromJson() error
		GetField(field ConfigField) (interface{}, error)
		UpdateField(field ConfigField, value interface{}) error
	}
)

const (
	ConfigGameVersion       ConfigField = "FFXGameVersion"
	ConfigGameFilesLocation ConfigField = "GameFilesLocation"
	ConfigExtractLocation   ConfigField = "ExtractLocation"
	ConfigTranslateLocation ConfigField = "TranslateLocation"
	ConfigImportLocation    ConfigField = "ImportLocation"
)

func newAppConfig(filePath string) *FFXAppConfig {
	ffxAppConfig := &FFXAppConfig{
		filePath: filePath,
	}
	err := ffxAppConfig.FromJson()
	if err != nil {
		return nil
	}

	return ffxAppConfig
}

func (c *FFXAppConfig) validateConfig() error {
	changed := false

	if c.FFXGameVersion <= 0 {
		c.FFXGameVersion = 1
		changed = true
	}
	if c.GameFilesLocation == "" {
		c.GameFilesLocation = NewInteractionService().GameLocation.GetTargetDirectory()
		changed = true
	}
	if c.ExtractLocation == "" {
		c.ExtractLocation = NewInteractionService().ExtractLocation.GetTargetDirectory()
		changed = true
	}
	if c.TranslateLocation == "" {
		c.TranslateLocation = NewInteractionService().TranslateLocation.GetTargetDirectory()
		changed = true
	}
	if c.ImportLocation == "" {
		c.ImportLocation = NewInteractionService().ImportLocation.GetTargetDirectory()
		changed = true
	}

	if changed {
		return c.ToJson()
	}

	return nil
}

func (c *FFXAppConfig) ToJson() error {
	if c == nil {
		return fmt.Errorf("%s", "invalid configuration")
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

func (c *FFXAppConfig) FromJson() error {
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

func (c *FFXAppConfig) GetField(field ConfigField) (interface{}, error) {
	if err := c.validateConfig(); err != nil {
		return nil, err
	}

	switch field {
	case ConfigGameVersion:
		return c.FFXGameVersion, nil
	case ConfigGameFilesLocation:
		return c.GameFilesLocation, nil
	case ConfigExtractLocation:
		return c.ExtractLocation, nil
	case ConfigTranslateLocation:
		return c.TranslateLocation, nil
	case ConfigImportLocation:
		return c.ImportLocation, nil
	default:
		return nil, fmt.Errorf("%s", "invalid field: "+string(field))
	}
}

func (c *FFXAppConfig) UpdateField(field ConfigField, value interface{}) error {
	changed := false

	switch field {
	case ConfigGameVersion:
		v, ok := value.(int)
		if !ok {
			return fmt.Errorf("incompatible value type for gamepart field")
		}
		c.FFXGameVersion = v
		changed = true
	case ConfigGameFilesLocation:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("incompatible value type for GameFilesLocation field")
		}
		c.GameFilesLocation = v
		changed = true
	case ConfigExtractLocation:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("incompatible value type for Extractlocation field")
		}
		c.ExtractLocation = v
		changed = true
	case ConfigTranslateLocation:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("incompatible value type for TranslateLocation field")
		}
		c.TranslateLocation = v
		changed = true
	case ConfigImportLocation:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("incompatible value type for Importlocation field")
		}
		c.ImportLocation = v
		changed = true
	default:
		return fmt.Errorf("%s", "invalid field: "+string(field))
	}

	if changed {
		if err := c.ToJson(); err != nil {
			return err
		}
	}

	return nil
}
