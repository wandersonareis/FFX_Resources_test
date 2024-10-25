package common

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