package textVerifier

import (
	"bytes"
	"fmt"
	"os"
)

type IComparer interface {
	CompareTextPartsContents(fromFile, toFile string) error
}

type Comparer struct{}

type FileInfo struct {
	Filename              string
	HeaderElements        int
	HeaderElementsNonZero int
}

func newPartComparer() IComparer {
	return &Comparer{}
}

func (pc Comparer) CompareTextPartsContents(fromFile, toFile string) error {
	if err := pc.compare(fromFile, toFile); err != nil {
		return err
	}

	return nil
}

func (pc Comparer) compare(fromFile, toFile string) error {
	newExtractedPartData, err := os.ReadFile(fromFile)
	if err != nil {
		return fmt.Errorf("error: %w when reading file: %s", err, fromFile)
	}

	importedPartData, err := os.ReadFile(toFile)
	if err != nil {
		return fmt.Errorf("error: %w when reading file: %s", err, toFile)
	}

	if !bytes.Equal(newExtractedPartData, importedPartData) {
		return fmt.Errorf("files are different: %s and %s", fromFile, toFile)
	}

	return nil
}
