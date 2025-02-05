package components

import (
	"bytes"
	"ffxresources/backend/common"
	"fmt"
	"io"
	"os"
)

func BinaryCompareByPath(fromFile, toFile string) error {
	f1, err := os.Open(fromFile)
	if err != nil {
		return fmt.Errorf("error opening source file: %s", common.GetFileName(fromFile))
	}
	defer f1.Close()

	f2, err := os.Open(toFile)
	if err != nil {
		return fmt.Errorf("error opening destination file: %s", common.GetFileName(toFile))
	}
	defer f2.Close()

	return BinaryCompare(f1, f2)
}
func BinaryCompare(fromFile, toFile *os.File) error {

	buf1 := make([]byte, 4096)
	buf2 := make([]byte, 4096)

	for {
		n1, e1 := fromFile.Read(buf1)
		n2, e2 := toFile.Read(buf2)

		if n1 != n2 || !bytes.Equal(buf1[:n1], buf2[:n2]) {
			return fmt.Errorf("files differ: %s", fromFile.Name())
		}

		if e1 == io.EOF && e2 == io.EOF {
			break
		}

		if e1 != nil && e1 != io.EOF {
			return fmt.Errorf("error reading source file: %s", fromFile.Name())
		}

		if e2 != nil && e2 != io.EOF {
			return fmt.Errorf("error reading destination file: %s", toFile.Name())
		}
	}
	return nil
}
