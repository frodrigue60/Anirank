package security

import (
	"github.com/microcosm-cc/bluemonday"
)

var (
	strictPolicy = bluemonday.StrictPolicy()
	ugcPolicy    = bluemonday.UGCPolicy()
)

// SanitizeStrict strips all HTML tags from the input string.
// Best for comments, user bios, and other plain-text fields.
func SanitizeStrict(input string) string {
	if input == "" {
		return ""
	}
	return strictPolicy.Sanitize(input)
}

// SanitizeHTML allows a safe subset of HTML tags (b, i, strong, em, a, br, etc).
// Best for content that specifically allows basic formatting (Anime descriptions, Announcements).
func SanitizeHTML(input string) string {
	if input == "" {
		return ""
	}
	return ugcPolicy.Sanitize(input)
}

// SanitizeStrictPtr does the same as SanitizeStrict but handles string pointers.
func SanitizeStrictPtr(input *string) *string {
	if input == nil {
		return nil
	}
	s := SanitizeStrict(*input)
	return &s
}

// SanitizeHTMLPtr does the same as SanitizeHTML but handles string pointers.
func SanitizeHTMLPtr(input *string) *string {
	if input == nil {
		return nil
	}
	s := SanitizeHTML(*input)
	return &s
}
