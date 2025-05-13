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

// IsValidFilePath checks if the given file path is valid.
// It returns false if the base name of the path is empty or starts with a dot.
// Otherwise, it returns true.
func IsValidFilePath(path string) bool {
	base := filepath.Base(path)

	if base == "" || strings.HasPrefix(base, ".") {
		return false
	}
	return true
}

// RemoveDir removes the directory at the specified path along with all its contents.
// The function returns an error if the removal process fails.
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

func hasExactComponent(path, component string) bool {
	clean := filepath.Clean(path)
	parts := strings.Split(clean, string(os.PathSeparator))
	return slices.Contains(parts, component)
}

func checkPS2Version1(path string) (int, string, bool) {
	p1 := "ffx_ps2"
	p2 := "ffx"
	p3 := "master"

	if hasExactComponent(path, p1) && hasExactComponent(path, p2) && hasExactComponent(path, p3) {
		return 1, p1, true
	}
	return 0, "", false
}

func checkPS2Version2(path string) (int, string, bool) {
	p1 := filepath.Join("ffx_ps2")
	p2 := filepath.Join("ffx2")
	p3 := filepath.Join("master")
	if hasExactComponent(path, p1) && hasExactComponent(path, p2) && hasExactComponent(path, p3) {
		return 2, p1, true
	}
	return 0, "", false
}

func checkDataVersion1(path string) (int, string, bool) {
	p1 := filepath.Join("ffx_data")
	p2 := filepath.Join("gamedata")
	p3 := filepath.Join("ps3data")
	if hasExactComponent(path, p1) && hasExactComponent(path, p2) && hasExactComponent(path, p3) {
		return 1, p1, true
	}
	return 0, "", false
}

func checkDataVersion2(path string) (int, string, bool) {
	p1 := filepath.Join("ffx-2_data")
	p2 := filepath.Join("gamedata")
	p3 := filepath.Join("ps3data")
	if hasExactComponent(path, p1) && hasExactComponent(path, p2) && hasExactComponent(path, p3) {
		return 2, p1, true
	}
	return 0, "", false
}

// CheckFFXPath resolves the given path to its absolute form and then validates it
// by running a series of checks for known PS2 and Data version patterns. If one of
// the checks succeeds, the function returns the detected version number along with
// a nil error. If no check passes, it returns 0 and an error indicating that the
// supplied path does not conform to a valid spira us path.
func CheckFFXPath(path string) (int, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return 0, fmt.Errorf("error getting absolute path: %s", err.Error())
	}

	checks := []func(string) (int, string, bool){
		checkPS2Version1,
		checkPS2Version2,
		checkDataVersion1,
		checkDataVersion2,
	}
	for _, check := range checks {
		if version, _, ok := check(absPath); ok {
			return version, nil
		}
	}
	return 0, fmt.Errorf("not a valid spira us path: %s", path)
}

// RelativePathFromMatch converts the given file path into its absolute path form,
// then inspects it against a series of predefined checks (such as checkPS2Version1,
// checkPS2Version2, checkDataVersion1, and checkDataVersion2) to determine if it contains
// a valid segment indicative of a recognized spira us path.
// 
// If a matching segment is found, the function extracts and returns the portion of the absolute path
// starting from this segment (including the trailing path separator). Otherwise, it returns an error
// indicating that the provided path is not a valid spira us path.
// 
// Parameters:
//   path - A string representing the file system path to be processed.
// 
// Returns:
//   A string containing the relative path starting from the matched segment, or an error if the
//   path does not conform to any of the expected spira us path formats.
func RelativePathFromMatch(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	checks := []func(string) (int, string, bool){
		checkPS2Version1,
		checkPS2Version2,
		checkDataVersion1,
		checkDataVersion2,
	}
	for _, check := range checks {
		if _, seg, ok := check(absPath); ok && seg != "" {
			idx := strings.Index(absPath, seg+string(os.PathSeparator))
			if idx != -1 {
				return absPath[idx:], nil
			}
		}
	}
	return "", fmt.Errorf("not a valid spira us path: %s", path)
}
