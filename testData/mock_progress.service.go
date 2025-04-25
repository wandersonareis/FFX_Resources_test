package testcommon

import "ffxresources/backend/services"

type MockProgressService struct {
	Max         int
	Steps       int
	Files       []string
	Started     bool
	Stopped     bool
	ProgressLog []services.ProgressService
}

func NewMockProgressService() *MockProgressService {
	return &MockProgressService{
		Files:       []string{},
		ProgressLog: []services.ProgressService{},
	}
}

func (m *MockProgressService) SetMax(max int) {
	m.Max = max
}

func (m *MockProgressService) Step() {
	m.Steps++
}

func (m *MockProgressService) StepFile(file string) {
	m.Files = append(m.Files, file)
	m.Steps++
}

func (m *MockProgressService) Start() {
	m.Started = true
}

func (m *MockProgressService) Stop() {
	m.Stopped = true
}
