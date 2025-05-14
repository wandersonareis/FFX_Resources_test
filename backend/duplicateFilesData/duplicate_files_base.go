package duplicateFilesData

type DuplicateFiles struct {
	fileDuplicationData DuplicateMap
}

func NewDuplicateFiles() *DuplicateFiles {
	return &DuplicateFiles{
		fileDuplicationData: make(DuplicateMap),
	}
}

func (d *DuplicateFiles) Add(items DuplicateMap) {
	for key, values := range items {
		if _, exists := d.fileDuplicationData[key]; !exists {
			d.fileDuplicationData[key] = values
		}
	}
}

func (d DuplicateFiles) Find(key string) []string {
	if values, exists := d.fileDuplicationData[key]; exists {
		return values
	}
	return nil
}
