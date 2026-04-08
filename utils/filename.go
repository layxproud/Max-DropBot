package utils

import "strings"

func SanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, " ", "_")

	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)

	return replacer.Replace(name)
}
