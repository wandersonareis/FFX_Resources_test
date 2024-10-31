package common

import "os"

func WriteStringToFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

func WriteBytesToFile(file string, data []byte) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write(data)

	return err
}