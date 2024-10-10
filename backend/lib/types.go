package lib

type NodeType int

const (
	None NodeType = iota
	Drive
	File
	Folder
	Dialogs
	Tutorial
	Kernel
	Dcp
)

type FileInfo struct {
	Name           string   `json:"name"`
	Size           int64    `json:"size"`
	Type           NodeType `json:"type"`
	Extension      string   `json:"extension"`
	Parent         string   `json:"parent"`
	IsDir          bool     `json:"is_dir"`
	AbsolutePath   string   `json:"absolute_path"`
	RelativePath   string   `json:"relative_path"`
	ExtractedFile  string   `json:"extracted_file"`
	ExtractedPath  string   `json:"extracted_path"`
	TranslatedFile string   `json:"translated_file"`
	TranslatedPath string   `json:"translated_path"`

	ExtractLocation ExtractLocation `json:"extract_location"`
}

type IFileProcessor interface {
	GetFileInfo() FileInfo
	Extract()
	Compress()
}

type ITextFormatter interface {
	Write(fileInfo FileInfo, targetDirectory string) (string, string)
}

type IExtractor interface {
	Extract()
}

type ICompressor interface {
	Compress()
}

type TreeNode struct {
	Key      string     `json:"key"`
	Label    string     `json:"label"`
	Data     FileInfo   `json:"data"`
	Icon     string     `json:"icon"`
	Children []TreeNode `json:"children"`
}
