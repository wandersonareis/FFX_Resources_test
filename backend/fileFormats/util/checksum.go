package util

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strings"
)

type Checksum struct {}

func (c *Checksum) VerifyChecksum(file, checksum string) bool {
	data, err := c.readBytes(file)
	if err != nil {
		return false
	}
	
	hash := sha256.Sum256(data)

	return strings.ToLower(checksum) == hex.EncodeToString(hash[:])
}

func (c *Checksum) readBytes(file string) ([]byte, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	return data, nil
}