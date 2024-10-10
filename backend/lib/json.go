package lib

import (
	"encoding/json"
	"os"
)

func SaveToJSONFile(data interface{}, filename string) error {
	// Criar o arquivo
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Serializar o objeto em JSON
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Formatar o JSON com identação
	err = encoder.Encode(data)
	if err != nil {
		return err
	}

	return nil
}