package testcommon

import (
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/interfaces"

	"github.com/stretchr/testify/mock"
)

// Mock para IDlgDecoder
type MockDlgDecoder struct {
	mock.Mock
}

func (m *MockDlgDecoder) Decoder(source interfaces.ISource, destination locations.IDestination, textEncoding ffxencoding.IFFXTextDlgEncoding) error {
	args := m.Called(source, destination, textEncoding)
	return args.Error(0)
}
