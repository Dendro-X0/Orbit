package run

import (
	"regexp"
	"strings"
)

var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(api[_-]?key|token|secret|password|pepper)\s*[:=]\s*\S+`),
	regexp.MustCompile(`(?i)bearer\s+[a-z0-9._-]+`),
	regexp.MustCompile(`ghp_[a-zA-Z0-9]+`),
	regexp.MustCompile(`gho_[a-zA-Z0-9]+`),
}

func RedactLine(line string) string {
	out := line
	for _, re := range secretPatterns {
		out = re.ReplaceAllStringFunc(out, func(match string) string {
			parts := strings.SplitN(match, "=", 2)
			if len(parts) == 2 {
				return parts[0] + "=***"
			}
			parts = strings.SplitN(match, ":", 2)
			if len(parts) == 2 {
				return parts[0] + ": ***"
			}
			return "***"
		})
	}
	return out
}
