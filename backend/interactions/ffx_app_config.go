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

func (c *FFXAppConfig) createConfig() error {
	c.FFXGameVersion = 1
	return c.ToJson()
}

func (c *FFXAppConfig) ToJson() error {
	if c == nil {
		return fmt.Errorf("%s", "configuração inválida")
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
			return c.createConfig()
		}
		return err
	}

	if len(bytes.TrimSpace(file)) == 0 {
		return c.createConfig()
	}

	err = json.Unmarshal(file, c)
	return err
}

func (c *FFXAppConfig) GetField(field ConfigField) (interface{}, error) {
	err := c.FromJson()
	if err != nil {
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
		return nil, fmt.Errorf("%s", "campo inválido: "+string(field))
	}
}

func (c *FFXAppConfig) UpdateField(field ConfigField, value interface{}) error {
	err := c.FromJson()
	if err != nil {
		return err
	}

	switch field {
	case ConfigGameVersion:
		v, ok := value.(int)
		if !ok {
			return fmt.Errorf("tipo de valor incompatível para o campo GamePart")
		}
		c.FFXGameVersion = v
	case ConfigGameFilesLocation:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("tipo de valor incompatível para o campo GameFilesLocation")
		}
		c.GameFilesLocation = v
	case ConfigExtractLocation:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("tipo de valor incompatível para o campo ExtractLocation")
		}
		c.ExtractLocation = v
	case ConfigTranslateLocation:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("tipo de valor incompatível para o campo TranslateLocation")
		}
		c.TranslateLocation = v
	case ConfigImportLocation:
		v, ok := value.(string)
		if !ok {
			return fmt.Errorf("tipo de valor incompatível para o campo ReimportLocation")
		}
		c.ImportLocation = v
	default:
		return fmt.Errorf("%s", "campo inválido: "+string(field))
	}

	err = c.ToJson()
	if err != nil {
		return err
	}
	return nil
}
