package core

import (
	"encoding/json"
	"fmt"
)

type Duplicate struct {
	Data map[string][]string
}

func NewDuplicate() *Duplicate {
	return &Duplicate{
		Data: make(map[string][]string),
	}
}

func (d *Duplicate) AddFromData(data string) error {
	var dataMap map[string][]string

	err := json.Unmarshal([]byte(data), &dataMap)
	if err != nil {
		return fmt.Errorf("error when decoding json: %w", err)
	}

	for key, values := range dataMap {
		if _, exists := d.Data[key]; !exists {
			d.Data[key] = values
		}
	}

	return nil
}

func (d Duplicate) Find(key string) []string {
	if _, exists := d.Data[key]; exists {
		return d.Data[key]
	}
	return nil
}

func (d Duplicate) GetKeys() []string {
	keys := make([]string, 0, len(d.Data))
	for key := range d.Data {
		keys = append(keys, key)
	}
	return keys
}

func (d Duplicate) GetValues() [][]string {
	values := make([][]string, 0, len(d.Data))
	for _, value := range d.Data {
		values = append(values, value)
	}

	return values
}
