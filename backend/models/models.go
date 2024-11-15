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
