package slug

import (
	"fmt"

	"github.com/gosimple/slug"
)

// Generate creates a URL-friendly slug from a title string.
func Generate(title string) string {
	return slug.Make(title) // "Rumah Mewah Jakarta" → "rumah-mewah-jakarta"
}

// GenerateUnique creates a URL-friendly slug and ensures it is unique by checking against an existence function.
func GenerateUnique(title string, existsFn func(string) bool) string {
	base := Generate(title)
	candidate := base
	for i := 2; existsFn(candidate); i++ {
		candidate = fmt.Sprintf("%s-%d", base, i)
	}
	return candidate
}
