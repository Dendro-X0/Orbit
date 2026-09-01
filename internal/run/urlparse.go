package run

import "regexp"

var (
	workersURLRe = regexp.MustCompile(`https://[a-zA-Z0-9-]+\.[a-zA-Z0-9.-]*workers\.dev`)
	vercelURLRe  = regexp.MustCompile(`https://[a-zA-Z0-9-]+\.vercel\.app`)
	flyURLRe     = regexp.MustCompile(`https://[a-zA-Z0-9-]+\.fly\.dev`)
)

// DeployURLs holds URLs parsed from deploy logs.
type DeployURLs struct {
	API  string
	Docs string
}

// ExtractWorkersURL returns the last Workers deploy URL found in log output.
func ExtractWorkersURL(text string) string {
	matches := workersURLRe.FindAllString(text, -1)
	if len(matches) == 0 {
		return ""
	}
	return matches[len(matches)-1]
}

// ExtractVercelURL returns the last Vercel deployment URL found in log output.
func ExtractVercelURL(text string) string {
	matches := vercelURLRe.FindAllString(text, -1)
	if len(matches) == 0 {
		return ""
	}
	return matches[len(matches)-1]
}

// ExtractFlyURL returns the last Fly.io deployment URL found in log output.
func ExtractFlyURL(text string) string {
	matches := flyURLRe.FindAllString(text, -1)
	if len(matches) == 0 {
		return ""
	}
	return matches[len(matches)-1]
}

// ExtractDeployURLs parses API and docs URLs from combined deploy output.
func ExtractDeployURLs(text string) DeployURLs {
	api := ExtractWorkersURL(text)
	if api == "" {
		api = ExtractFlyURL(text)
	}
	return DeployURLs{
		API:  api,
		Docs: ExtractVercelURL(text),
	}
}
