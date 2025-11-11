package helpers

import "strings"

// FuzzyMatch performs fuzzy matching on a string
// Returns true if all characters in query appear in target in order (case-insensitive)
func FuzzyMatch(query, target string) bool {
	if query == "" {
		return true
	}

	// Convert to lowercase for case-insensitive matching
	query = strings.ToLower(query)
	target = strings.ToLower(target)

	queryIdx := 0
	for _, char := range target {
		if queryIdx < len(query) && rune(query[queryIdx]) == char {
			queryIdx++
		}
		if queryIdx == len(query) {
			return true
		}
	}
	return queryIdx == len(query)
}

// ToLower converts a string to lowercase (simple ASCII version for performance)
func ToLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + ('a' - 'A')
		} else {
			result[i] = c
		}
	}
	return string(result)
}

// Trim removes leading and trailing whitespace
func Trim(s string) string {
	start := 0
	end := len(s)

	// Trim leading whitespace
	for start < end {
		c := s[start]
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			break
		}
		start++
	}

	// Trim trailing whitespace
	for end > start {
		c := s[end-1]
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			break
		}
		end--
	}

	return s[start:end]
}

// Contains checks if a substring exists in a string
func Contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(substr) > len(s) {
		return false
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// ToUpperFirst converts the first word of a string to uppercase
func ToUpperFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	// Find first word (up to first space or entire string)
	endIdx := 0
	for endIdx < len(s) && s[endIdx] != ' ' && s[endIdx] != '\t' && s[endIdx] != '\n' {
		endIdx++
	}
	firstWord := s[:endIdx]

	// Convert to uppercase
	result := make([]byte, len(firstWord))
	for i := 0; i < len(firstWord); i++ {
		c := firstWord[i]
		if c >= 'a' && c <= 'z' {
			result[i] = c - ('a' - 'A')
		} else {
			result[i] = c
		}
	}
	return string(result)
}

// StartsWithAny checks if a string starts with any of the given prefixes (case-sensitive)
func StartsWithAny(s string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if len(s) >= len(prefix) {
			match := true
			for i := 0; i < len(prefix); i++ {
				if s[i] != prefix[i] {
					match = false
					break
				}
			}
			if match {
				return true
			}
		}
	}
	return false
}
