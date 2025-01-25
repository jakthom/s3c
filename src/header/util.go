package header

import (
	"fmt"
	"strings"
)

// StripETagQuotes removes leading and trailing quotes in a string if they exist. Used for ETags.
func StripETagQuotes(s string) string {
	if strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
		return strings.Trim(s, "\"")
	}
	return s
}

// AddETagQuotes ensures that an etag string has leading and trailing quotes. Used for ETags.
func AddETagQuotes(s string) string {
	if !strings.HasPrefix(s, "\"") {
		return fmt.Sprintf("\"%s\"", s)
	}
	return s
}
