package components

import (
	"ffxresources/backend/interfaces"
)

// IList is an alias for interfaces.IList - kept for backward compatibility
type IList[T any] = interfaces.IList[T]

// IMap is an alias for interfaces.IMap - kept for backward compatibility
type IMap[K comparable, V any] = interfaces.IMap[K, V]

// IFileComparer is an alias for interfaces.IFileComparer - kept for backward compatibility
type IFileComparer = interfaces.IFileComparer
