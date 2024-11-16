package util

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strings"
)

type Checksum struct {
	Checksum string
}

func (c *Checksum) IsValid(file string) bool {
	data, err := c.readBytes(file)
	if err != nil {
		return false
	}
	
	hash := sha256.Sum256(data)

	return c.Checksum == hex.EncodeToString(hash[:])
}

func (c *Checksum) SetChecksumString(sum string) {
	c.Checksum = strings.ToLower(sum)
}

func (c *Checksum) readBytes(file string) ([]byte, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	return data, nil
}