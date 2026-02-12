package helpers

import "strings"

// MaskString masks a string with asterisks, leaving cursor position visible
func MaskString(s string, cursorPos int) string {
	runes := []rune(s)
	masked := make([]rune, len(runes))

	for i := range runes {
		if i == cursorPos {
			masked[i] = '█' // Cursor
		} else {
			masked[i] = '*'
		}
	}

	return string(masked)
}

// InsertRune inserts a rune at the specified position in a string
func InsertRune(s string, r rune, pos int) string {
	runes := []rune(s)
	if pos < 0 {
		pos = 0
	}
	if pos > len(runes) {
		pos = len(runes)
	}

	result := make([]rune, len(runes)+1)
	copy(result[:pos], runes[:pos])
	result[pos] = r
	copy(result[pos+1:], runes[pos:])

	return string(result)
}

// DeleteRune deletes a rune at the specified position (backspace)
func DeleteRune(s string, pos int) (string, int) {
	runes := []rune(s)
	if pos <= 0 || len(runes) == 0 {
		return s, pos
	}

	result := make([]rune, len(runes)-1)
	copy(result[:pos-1], runes[:pos-1])
	copy(result[pos-1:], runes[pos:])

	return string(result), pos - 1
}

// FilterStrings filters a slice of strings based on a query
func FilterStrings(items []string, query string) []string {
	if query == "" {
		return items
	}

	query = strings.ToLower(query)
	filtered := make([]string, 0)

	for _, item := range items {
		if strings.Contains(strings.ToLower(item), query) {
			filtered = append(filtered, item)
		}
	}

	return filtered
}
