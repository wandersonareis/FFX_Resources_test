package common

import (
	"sync"

	"github.com/google/uuid"
)

var (
	sessionUUID string
	sessionOnce sync.Once
)

func GetUUID() string {
	sessionOnce.Do(func() {
		sessionUUID = generateUUID()
	})

	return sessionUUID
}

func generateUUID() string {
	return uuid.New().String()
}
