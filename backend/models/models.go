package models

type NodeType int

const (
	None NodeType = iota
	Drive
	File
	Folder
	Dialogs
	DialogsSpecial
	Tutorial
	Kernel
	Dcp
	DcpParts
	Lockit
	LockitParts
)

type GameVersion int

const (
	FFX GameVersion = iota + 1
	FFX2
)

func (gv GameVersion) String() string {
	switch gv {
	case FFX:
		return "FFX"
	case FFX2:
		return "FFX-2"
	default:
		return "Unknown"
	}
}

type GameDataInfo struct {
	FilePath       string `json:"file_path"`
	ExtractedFile  string `json:"extracted_file"`
	TranslatedFile string `json:"translated_file"`
	ImportedFile   string `json:"imported_file"`
}

type Pointer struct {
	Offset int64
	Value  uint32
}
