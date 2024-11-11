package core

import "strings"

type Duplicate struct {
	Data map[string][]string
}

func NewDuplicate() *Duplicate {
	return &Duplicate{
		Data: make(map[string][]string),
	}
}

func (d *Duplicate) AddFromSpaceSeparatedString(key, spaceSeparatedValues string) {
	values := strings.Fields(spaceSeparatedValues)

	if _, exists := d.Data[key]; !exists {
		d.Data[key] = values
	}
}

/* func add(key string, values []string) {
	if key == "" || values == nil || len(values) == 0 {
		return
	}

	if _, exists := d.data[key]; !exists {
		data[key] = values
	}
} */

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
