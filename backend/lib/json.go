package lib

import (
	"encoding/json"
	"os"
)

func SaveToJSONFile(data interface{}, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(data)
	if err != nil {
		return err
	}

	return nil
}

func LoadFromJson(data interface{}, path string) error {
	file, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(file, &data)
	return err
}
