package utils

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	regexpSpecialChars = regexp.MustCompile(`[^a-z0-9-]+`)
	regexpMultipleDash = regexp.MustCompile(`-+`)
)

// Slugify converts a string into a URL-friendly slug
func Slugify(text string) string {
	slug := strings.ToLower(strings.TrimSpace(text))
	slug = regexpSpecialChars.ReplaceAllString(slug, "-")
	slug = regexpMultipleDash.ReplaceAllString(slug, "-")
	return strings.Trim(slug, "-")
}

// GenerateUniqueSlug produces a unique slug using a provided existence checker
func GenerateUniqueSlug(baseText string, exists func(string) bool) string {
	baseSlug := Slugify(baseText)
	if baseSlug == "" {
		baseSlug = "user"
	}

	slug := baseSlug
	counter := 1

	for exists(slug) {
		counter++
		slug = fmt.Sprintf("%s-%d", baseSlug, counter)
		
		// Safety break to prevent infinite loops in extreme cases
		if counter > 1000 {
			break
		}
	}

	return slug
}
