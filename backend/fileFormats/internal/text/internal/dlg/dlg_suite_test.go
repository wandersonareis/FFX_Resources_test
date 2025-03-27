package dlg_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDlg(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dlg Suite")
}
