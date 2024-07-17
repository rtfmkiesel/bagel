package router

import "github.com/microcosm-cc/bluemonday"

var (
	bluemondayPolicy = bluemonday.StrictPolicy()
)

// SanitizeHTML sanitizes the given HTML string using bluemonday's StrictPolicy
func SanitizeHTML(html string) string {
	return bluemondayPolicy.Sanitize(html)
}
