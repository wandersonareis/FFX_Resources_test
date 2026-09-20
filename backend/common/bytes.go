package common

import "bytes"

// GetStringBytesAtLookupOffset retrieves a null-terminated string from a byte table
// starting at the specified offset. The function reads bytes from the offset position
// until it encounters a null byte (0x00) or reaches the end of the table.
//
// Parameters:
//   - table: The byte slice containing the string data
//   - offset: The starting position in the table to begin reading from
//
// Returns:
//   - []byte: The string bytes without the null terminator, or nil if offset is invalid
//   - If no null terminator is found, returns all bytes from offset to end of table
//   - Returns nil if offset is negative or beyond the table length
//   - For directories: Recursively processes all non-hidden files in sorted order
//   - For files: Resolves path, reads bytes, and parses as string data using appropriate charset
func GetStringBytesAtLookupOffset(table []byte, offset int) []byte {
	if offset < 0 || offset >= len(table) {
		return nil
	}

	end := bytes.IndexByte(table[offset:], 0x00)
	if end == -1 {
		return table[offset:]
	}
	subArray := table[offset : offset+end]
	var newArray = make([]byte, len(subArray))
	copy(newArray, subArray)
	return newArray
}
