package generator

import "strings"

// JSONCamelCase converts a snake_case identifier to a camelCase identifier,
// according to the protobuf JSON specification.
//
// Copied from: google.golang.org/protobuf/internal/strs
func JSONCamelCase(s string) string {
	var b []byte
	var wasUnderscore bool
	for i := 0; i < len(s); i++ { // proto identifiers are always ASCII
		c := s[i]
		if c != '_' {
			if wasUnderscore && isASCIILower(c) {
				c -= 'a' - 'A' // convert to uppercase
			}
			b = append(b, c)
		}
		wasUnderscore = c == '_'
	}
	return string(b)
}

func isASCIILower(c byte) bool {
	return 'a' <= c && c <= 'z'
}

// Takes a service name or method name in the form SomeMethod or HTTPMethod and
// return someMethod or httpMethod.
func functionCase(s string) string {
	if len(s) == 0 {
		return s
	}

	// Find the position of the first non-uppercase letter after the initial uppercase sequence
	firstLowerPos := -1
	for i := 0; i < len(s); i++ {
		if isASCIILower(s[i]) {
			firstLowerPos = i
			break
		}
	}

	switch firstLowerPos {
	case -1:
		// If no lowercase letter is found, return the string in lowercase
		return strings.ToLower(s)

	case 0:
		// If the first letter is lowercase, return the string as is.
		return s

	case 1:
		// If only the first letter is upper, we want to lowercase it.
		return strings.ToLower(s[:1]) + s[1:]

	default:
		// If multiple letters are upper case, we want all but the last one.
		return strings.ToLower(s[:firstLowerPos-1]) + s[firstLowerPos-1:]
	}
}
