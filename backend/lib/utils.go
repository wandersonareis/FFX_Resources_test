package lib

import "strings"

func IsEmptyOrWhitespace(s *string) bool {
	if s == nil {
		return true
	}
	return strings.TrimSpace(*s) == ""
}
