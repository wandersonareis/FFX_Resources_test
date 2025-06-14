package testcommon

import (
	"ffxresources/backend/loggingService"
	"fmt"
)

type LoggedMessage struct {
	Type    string
	Message string
}

type MockLogHandler struct {
	Messages []LoggedMessage
}

func NewLogHandlerMock() *MockLogHandler {
	return &MockLogHandler{
		Messages: []LoggedMessage{},
	}
}

func (m *MockLogHandler) GetLogger() *loggingService.LoggerService {
	fmt.Println("[LOGMOCK] GetLogger chamado")
	return nil
}

func (m *MockLogHandler) Info(message string, args ...interface{}) {
	var msg string
	if len(args) > 0 {
		msg = fmt.Sprintf(message, args...)
		fmt.Printf("[LOGMOCK][INFO] "+message+"\n", args...)
	} else {
		msg = message
		fmt.Printf("[LOGMOCK][INFO] %s\n", message)
	}
	m.Messages = append(m.Messages, LoggedMessage{
		Type:    "info",
		Message: msg,
	})
}

func (m *MockLogHandler) Error(err error, message string, args ...interface{}) {
	var msg string
	if err != nil {
		if len(args) > 0 {
			msg = fmt.Sprintf(message, args...)
			fmt.Printf("[LOGMOCK][ERROR] %s | erro: %v\n", msg, err)
		} else {
			msg = message
			fmt.Printf("[LOGMOCK][ERROR] %s | erro: %v\n", message, err)
		}
	} else {
		if len(args) > 0 {
			msg = fmt.Sprintf(message, args...)
			fmt.Printf("[LOGMOCK][ERROR] "+message+"\n", args...)
		} else {
			msg = message
			fmt.Printf("[LOGMOCK][ERROR] %s\n", message)
		}
	}
	m.Messages = append(m.Messages, LoggedMessage{
		Type:    "error",
		Message: msg,
	})
}
