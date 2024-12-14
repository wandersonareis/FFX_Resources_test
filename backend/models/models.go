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

type GameDataInfo struct {
	FilePath       string `json:"file_path"`
	ExtractedFile  string `json:"extracted_file"`
	TranslatedFile string `json:"translated_file"`
	ImportedFile   string `json:"imported_file"`
}

/* type TreeNode struct {
	Key      string       `json:"key"`
	Label    string       `json:"label"`
	Data     GameDataInfo `json:"data"`
	Icon     string       `json:"icon"`
	Children []TreeNode   `json:"children"`
} */

type Pointer struct {
	Offset int64
	Value  uint32
}

type IExtractor interface {
	Extract() error
}

type ICompressor interface {
	Compress() error
}
