package models

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
	Lockit
	LockitParts
)

type LocationInterface interface {
	SetPath(path string)
	GetPath() string
	ProvideTargetDirectory() (string, error)
	TargetFileExists() bool
}

type Pointer struct {
	Offset int64
	Value  uint32
}

type IExtractor interface {
	Extract()
}

type ICompressor interface {
	Compress()
}
