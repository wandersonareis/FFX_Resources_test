package testcommon

import (
	"ffxresources/backend/services"
	"fmt"
)

type MockNotifier struct {
    Notifications []services.Notification
}

func NewMockNotifier() *MockNotifier {
    return &MockNotifier{Notifications: []services.Notification{}}
}

func (m *MockNotifier) Notify(notification services.Severity, message string) {
    n := services.Notification{
        Message:  message,
        Severity: notification.String(),
    }
    m.Notifications = append(m.Notifications, n)
    fmt.Printf("[NOTIFIERMOCK] %s: %s\n", n.Severity, n.Message)
}

func (m *MockNotifier) NotifyError(err error) {
    m.Notify(services.SeverityError, err.Error())
}

func (m *MockNotifier) NotifyInfo(message string) {
    m.Notify(services.SeverityInfo, message)
}

func (m *MockNotifier) NotifyWarn(message string) {
    m.Notify(services.SeverityWarn, message)
}

func (m *MockNotifier) NotifySuccess(message string) {
    m.Notify(services.SeveritySuccess, message)
}