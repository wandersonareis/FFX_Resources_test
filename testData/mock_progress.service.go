package testcommon

import (
	"ffxresources/backend/services"
	"fmt"
)

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

	m.checkSteps()
}

func (m *MockProgressService) StepFile(file string) {
	m.Files = append(m.Files, file)
	m.Steps++

	m.checkSteps()
}

func (m *MockProgressService) Start() {
	m.Started = true
}

func (m *MockProgressService) Stop() {
	m.Stopped = true
}

func (m *MockProgressService) logMsg(msg string) {
	fmt.Printf("[PROGRESSMOCK] %s\n", msg)
}

func (m *MockProgressService) checkSteps() {
	if m.Steps > m.Max {
		m.logMsg(fmt.Sprintf("Error on progress: Steps %d greater %d expected steps", m.Steps, m.Max))
	}

	if m.Steps == m.Max {
		m.logMsg(fmt.Sprintf("Progress completed: %d of %d", m.Steps, m.Max))
	}
}
