package utils

import (
	"strings"
	"unicode"
)

// ToCamelCase converts snake_case, kebab-case, or other delimited formats to CamelCase.
// It handles delimiters such as '_', '-', and '#'. Acronyms are preserved in uppercase.
func ToCamelCase(input string, pascalCase bool) string {
	words := strings.FieldsFunc(input, func(r rune) bool {
		return r == '_' || r == '-' || r == ' ' || r == '#'
	})

	for i := 0; i < len(words); i++ {
		if i == 0 && !pascalCase {
			// If the first word is already in camelCase, leave it as is
			if len(words[i]) > 0 && unicode.IsLower([]rune(words[i])[0]) {
				continue
			}
			// Lowercase the first word for camelCase
			words[i] = strings.ToLower(words[i])
		} else {
			// Capitalize the first letter of each subsequent word for both camelCase and PascalCase
			runes := []rune(words[i])
			if len(runes) > 0 {
				runes[0] = unicode.ToUpper(runes[0])
			}
			words[i] = string(runes)
		}
	}

	return strings.Join(words, "")
}

// isAcronym checks if all characters in the slice are uppercase letters.
// This helps in preserving acronyms like "API", "ID", etc.
func isAcronym(runes []rune) bool {
	if len(runes) <= 1 {
		return false
	}
	for _, r := range runes {
		if !unicode.IsUpper(r) {
			return false
		}
	}
	return true
}

// SanitizeString removes or replaces unwanted characters from a string.
func SanitizeString(input string) string {
	// Remove leading/trailing spaces and replace multiple spaces with a single space
	trimmed := strings.TrimSpace(input)
	sanitized := strings.Join(strings.Fields(trimmed), " ")
	return sanitized
}

// ToSnakeCase converts CamelCase or PascalCase strings to snake_case.
func ToSnakeCase(input string) string {
	var result []rune
	runes := []rune(input)

	for i, r := range runes {
		if unicode.IsUpper(r) {
			// Insert underscore if it's not the first character
			// and the previous character is lowercase
			// or the next character is lowercase (handling acronyms).
			if i > 0 && (unicode.IsLower(runes[i-1]) || (i+1 < len(runes) && unicode.IsLower(runes[i+1]))) {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}

	return string(result)
}

// Wrap strings.Contains for checks if the substring is present within the main string.
func Contains(mainStr, substr string) bool {
	return strings.Contains(mainStr, substr)
}

// Wrap strings.ReplaceAll to replace all instances of old with new in the input string.
func ReplaceAll(input, old, new string) string {
	return strings.ReplaceAll(input, old, new)
}

// Reverse returns the reverse of the input string.
// Handles multi-byte Unicode characters.
func Reverse(input string) string {
	runes := []rune(input)
	n := len(runes)
	for i := 0; i < n/2; i++ {
		runes[i], runes[n-1-i] = runes[n-1-i], runes[i]
	}
	return string(runes)
}

// IsEmpty checks if the input string is empty or contains only whitespace.
// It returns true if the string is empty or consists solely of whitespace characters.
func IsEmpty(input string) bool {
	return len(strings.TrimSpace(input)) == 0
}
