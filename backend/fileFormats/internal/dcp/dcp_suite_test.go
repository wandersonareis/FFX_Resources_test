package dcp_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDcp(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dcp Suite")
}
