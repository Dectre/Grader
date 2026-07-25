package utils

import "strings"

func RemoveBOM(s string) string {
	if strings.HasPrefix(s, "\xef\xbb\xbf") {
		return s[3:]
	}
	return s
}

func CleanString(s string) string {
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "\u200c", "")
	s = strings.ReplaceAll(s, "_", "")
	return strings.ToLower(s)
}

func CleanID(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, ".0") {
		s = strings.TrimSuffix(s, ".0")
	}
	return s
}