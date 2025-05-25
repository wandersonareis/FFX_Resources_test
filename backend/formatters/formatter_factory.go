package formatters

import (
	"ffxresources/backend/interfaces"
	"fmt"
)

type IFormatterFactory interface {
	CreateFormatter(ext string) (interfaces.ITextFormatter, error)
}

type formatterFactory struct{}

func NewFormatterFactory() IFormatterFactory {
	return &formatterFactory{}
}

func (df *formatterFactory) CreateFormatter(ext string) (interfaces.ITextFormatter, error) {
	switch ext {
	case ".txt":
		return NewTxtFormatter(), nil
	default:
		return nil, fmt.Errorf("formatter not available for extension %s", ext)
	}
}