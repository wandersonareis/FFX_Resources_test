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
	DcpParts
)

type FileInfo struct {
	Name              string            `json:"name"`
	NamePrefix        string            `json:"name_prefix"`
	Size              int64             `json:"size"`
	Type              NodeType          `json:"type"`
	Extension         string            `json:"extension"`
	Parent            string            `json:"parent"`
	IsDir             bool              `json:"is_dir"`
	AbsolutePath      string            `json:"absolute_path"`
	RelativePath      string            `json:"relative_path"`
	ExtractLocation   ExtractLocation   `json:"extract_location"`
	TranslateLocation TranslateLocation `json:"translate_location"`
	ImportLocation    ImportLocation    `json:"import_location"`
}

type IFileProcessor interface {
	GetFileInfo() *FileInfo
	Extract()
	Compress()
}

type ITextFormatter interface {
	ReadFile(fileInfo *FileInfo, targetDirectory string) (string, string)
	WriteFile(fileInfo *FileInfo, targetDirectory string) (string, string)
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
	Data     *FileInfo  `json:"data"`
	Icon     string     `json:"icon"`
	Children []TreeNode `json:"children"`
}

/* type LocationBase struct {
	IsExist             bool
	TargetFile          string
	TargetPath          string
	TargetFileName      string
	TargetDirectory     string
	TargetDirectoryName string
}

func (lb *LocationBase) SetPath(path string) {
	if path == "" {
		return
	}

	lb.TargetDirectory = path
}

func (lb *LocationBase) GetPath() string {
	return lb.TargetDirectory
} */

type Pointer struct {
	Offset int64
	Value  uint32
}