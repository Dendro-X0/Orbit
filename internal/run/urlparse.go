package run

import "regexp"

var workersURLRe = regexp.MustCompile(`https://[a-zA-Z0-9-]+\.[a-zA-Z0-9.-]*workers\.dev`)

// ExtractWorkersURL returns the last Workers deploy URL found in log output.
func ExtractWorkersURL(text string) string {
	matches := workersURLRe.FindAllString(text, -1)
	if len(matches) == 0 {
		return ""
	}
	return matches[len(matches)-1]
}
