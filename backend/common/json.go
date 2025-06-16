package common

import (
	"encoding/json"
	"os"
)

// ReadJsonFile reads and parses a JSON file into the specified type T.
// It takes a filename as input and returns the parsed data of type T and an error.
// The function uses generics to allow parsing into any type that can be unmarshaled from JSON.
// If the file cannot be opened or the JSON cannot be decoded, an error is returned.
//
// Parameters:
//   - filename: the path to the JSON file to read
//
// Returns:
//   - T: the parsed data of the specified type
//   - error: nil on success, or an error if file reading or JSON parsing fails
func ReadJsonFile[T any](filename string) (T, error) {
	var data T
	file, err := OpenFile(filename)
	if err != nil {
		return data, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return data, err
	}

	return data, nil
}

// SaveAsJSON marshals the provided data to JSON format with indentation and writes it to the specified file path.
// The JSON output is formatted with 2-space indentation for readability.
//
// Parameters:
//   - data: any type of data that can be marshaled to JSON
//   - filePath: the target file path where the JSON data will be written
//
// Returns:
//   - error: nil if successful, otherwise returns the error from JSON marshaling or file writing operations
//
// The function creates or overwrites the target file with permissions 0644 (read-write for owner, read-only for group and others).
func SaveAsJSON(data any, filePath string) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}
