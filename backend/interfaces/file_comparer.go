package interfaces

// IFileComparer defines an interface for comparing files.
type IFileComparer interface {
	CompareFiles() error
}
