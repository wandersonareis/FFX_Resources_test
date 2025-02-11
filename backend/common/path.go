package common

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func sanitizationPath(path string) string {
	return filepath.Clean(path)
}

// GetDir takes a file path as input, sanitizes it, and returns the directory part of the path.
//
// Parameters:
//   - path: A string representing the file path to be processed.
//
// Returns:
//   - A string representing the directory part of the sanitized file path.
//     already exists or was successfully created.
func GetDir(path string) string {
	cPath := sanitizationPath(path)
	return filepath.Dir(cPath)
}

// IsPathExists checks if the given path exists in the file system.
// It sanitizes the input path before performing the check.
//
// Parameters:
//   - path: The file system path to check.
//
// Returns:
//   - bool: true if the path exists, false otherwise.
//     already exists or was successfully created.
func IsPathExists(path string) bool {
	cPath := sanitizationPath(path)
	_, err := os.Stat(cPath)
	return !os.IsNotExist(err)
}

// EnsurePathExists ensures that the directory for the given path exists.
// If the path includes a file name, the directory containing the file will be created.
// If the directory does not exist, it will be created with the appropriate permissions.
//
// Parameters:
//
//	path - The path for which to ensure the existence of the directory.
//
// Returns:
//
//	error - An error if the directory could not be created, or nil if the directory
//	        already exists or was successfully created.
func EnsurePathExists(path string) error {
	cPath := sanitizationPath(path)

	if filepath.Ext(cPath) != "" {
		cPath = filepath.Dir(cPath)
	}

	if err := os.MkdirAll(cPath, os.ModePerm); err != nil {
		return fmt.Errorf("error when creating the destination directory: %s", err.Error())
	}

	return nil
}

func RemoveDir(path string) error {
	cPath := sanitizationPath(path)
	return os.RemoveAll(cPath)
}

// GetRelativePathFromMarker takes a file path as input and returns the relative path
// starting from a predefined marker. If the marker is not found in the path, it logs
// an error message and returns an empty string.
//
// Parameters:
//   - path: The full file path as a string.
//
// Returns:
//   - A string representing the relative path starting from the marker, or an empty
//     string if the marker is not found.
func GetRelativePathFromMarker(path string) string {
	var marker = FFX_DIR_MARKER

	index := strings.Index(path, marker)
	if index == -1 {
		log.Println("unable to find marker in path:", path)
		return ""
	}

	return path[index:]
}

// MakeRelativePath takes two file paths, 'from' and 'to', and returns a relative path from 'from' to 'to'.
// If the paths are identical, it returns an empty string.
// If 'from' starts with 'to', it trims the 'to' prefix from 'from' and returns the result.
// Otherwise, it returns the 'from' path as is.
//
// Parameters:
//   - from: The starting file path.
//   - to: The target file path.
//
// Returns:
//
//	A string representing the relative path from 'from' to 'to'.
func MakeRelativePath(fromPath, toPath string) string {
	from := filepath.Clean(fromPath)
	to := filepath.Clean(toPath)

	// Remove volume name from the paths before comparison
	vol1 := filepath.VolumeName(from)
	vol2 := filepath.VolumeName(to)
	from = strings.TrimPrefix(from, vol1)
	to = strings.TrimPrefix(to, vol2)

	// Change path values keeping the longest one in p1
	if len(from) < len(to) {
		from, to = to, from
	}

	if from == to {
		return ""
	}

	if !strings.HasPrefix(from, to) {
		return ""
	}

	result := strings.TrimPrefix(from, to)
	return strings.TrimPrefix(result, string(os.PathSeparator))
}

// ContainsNewUSPCPath checks if the provided path contains the required sequence
// for a valid Spira US path. The required sequence is "ffx_ps2/ffx2/master/new_uspc".
// If the required sequence is not found in the provided path, an error is returned.
//
// Parameters:
//   - path: The path to be checked.
//
// Returns:
//   - error: An error indicating that the path is not a valid Spira US path, or nil if the path is valid.
func ContainsNewUSPCPath(path string) error {
	cPath := filepath.Clean(path)

	requiredSequence := filepath.Join("ffx_ps2", "ffx2", "master", "new_uspc")
	requiredPath := filepath.Join(cPath, requiredSequence)

	if !IsPathExists(requiredPath) {
		return fmt.Errorf("is not a valid spira us path: %s", path)
	}

	return nil
}

// ContainsGameResourcesPath checks if the provided path contains the required game resources path.
// It constructs the required path by joining the provided path with the sequence "ffx-2_data/gamedata/ps3data".
// If the constructed path does not exist, it returns an error indicating that the path is not a valid Spira game resources US path.
// Otherwise, it returns nil.
//
// Parameters:
//   - path: The base path to check.
//
// Returns:
//   - error: An error if the required path does not exist, otherwise nil.
func ContainsGameResourcesPath(path string) error {
	cPath := filepath.Clean(path)

	requiredSequence := filepath.Join("ffx-2_data", "gamedata", "ps3data")
	requiredPath := filepath.Join(cPath, requiredSequence)

	if !IsPathExists(requiredPath) {
		return fmt.Errorf("is not a valid spira game resources us path: %s", path)
	}

	return nil
}
