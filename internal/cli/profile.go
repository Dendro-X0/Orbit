package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// projectProfile summarizes what kind of project this repo looks like.
type projectProfile struct {
	Type        string
	Title       string
	Summary     string
	SuggestedID string
	Reason      string
	Signals     []string
}

func collectSignals(root string) []string {
	var signals []string
	add := func(s string) {
		if s != "" {
			signals = append(signals, s)
		}
	}

	if runFileExists(filepath.Join(root, signetConfig)) || runFileExists(filepath.Join(root, legacySignetCfg)) {
		add("signet.toml")
	}
	if runFileExists(filepath.Join(root, "apps", "api", "wrangler.toml")) {
		add("monorepo:api")
	}
	if runFileExists(filepath.Join(root, "wrangler.toml")) {
		add("wrangler.toml")
	}
	if runFileExists(filepath.Join(root, "apps", "docs", "vercel.json")) {
		add("monorepo:docs")
	}
	if runFileExists(filepath.Join(root, "vercel.json")) {
		add("vercel.json")
	}
	if runFileExists(filepath.Join(root, "fly.toml")) {
		add("fly.toml")
	}
	if runFileExists(filepath.Join(root, "netlify.toml")) {
		add("netlify.toml")
	}
	if runFileExists(filepath.Join(root, "Dockerfile")) {
		add("dockerfile")
	}
	if hasMCPServerSignal(root) {
		add("mcp:package")
	}
	if signalContains(signals, "monorepo:api") && signalContains(signals, "monorepo:docs") {
		add("monorepo:api+docs")
	}
	return signals
}

func hasMCPServerSignal(root string) bool {
	paths := []string{
		filepath.Join(root, "package.json"),
		filepath.Join(root, "apps", "api", "package.json"),
	}
	for _, path := range paths {
		b, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		lower := strings.ToLower(string(b))
		if strings.Contains(lower, "@modelcontextprotocol") ||
			strings.Contains(lower, "mcp-server") ||
			strings.Contains(lower, `"mcp"`) {
			return true
		}
	}
	return false
}

func detectProjectProfile(root string, components []deployComponent) projectProfile {
	signals := collectSignals(root)
	profile := projectProfile{Signals: signals}

	switch {
	case signalContains(signals, "monorepo:api+docs"):
		profile.Type = "full_web_product"
		profile.Title = "Full web product"
		profile.Summary = "API backend with a separate docs or frontend app"
		profile.SuggestedID = "full_stack"
		profile.Reason = "Found apps/api and apps/docs — deploy together or separately"
	case signalContains(signals, "signet.toml") && len(cloudComponents(components)) == 0:
		profile.Type = "desktop_app"
		profile.Title = "Desktop / mobile app"
		profile.Summary = "Installable artifacts — use Signet to build and release"
		profile.Reason = "Found signet.toml"
	case signalContains(signals, "mcp:package"):
		profile.Type = "mcp_server"
		profile.Title = "MCP server"
		profile.Summary = "Model Context Protocol server to host"
		profile.SuggestedID = suggestComponentID(components, "app:", "api:")
		profile.Reason = "package.json references MCP"
	case len(apiComponents(components)) == 1 && len(frontendComponents(components)) == 0:
		c := apiComponents(components)[0]
		profile.Type = "api_backend"
		profile.Title = "API / backend"
		profile.Summary = c.Description
		profile.SuggestedID = "api_backend"
		profile.Reason = "Found Workers or container API configuration"
	case len(frontendComponents(components)) == 1 && len(apiComponents(components)) == 0:
		c := frontendComponents(components)[0]
		profile.Type = "web_frontend"
		profile.Title = "Web frontend"
		profile.Summary = c.Description
		profile.SuggestedID = "web_frontend"
		profile.Reason = "Found frontend or static site configuration"
	case len(apiComponents(components)) > 0 && len(frontendComponents(components)) > 0:
		profile.Type = "full_web_product"
		profile.Title = "Full web product"
		profile.Summary = "Backend and frontend can deploy together"
		profile.SuggestedID = "full_stack"
		profile.Reason = "Detected both API and frontend — deploy together or separately"
	default:
		if len(components) > 0 {
			profile.Type = "multi_target"
			profile.Title = "Multi-part project"
			profile.Summary = "Several deployable parts found"
			intents := detectDeployIntents(components)
			if len(intents) > 0 {
				profile.SuggestedID = intents[0].ID
			}
			profile.Reason = "Multiple targets detected — pick the part you need"
		}
	}

	if profile.SuggestedID != "" && !hasIntentID(components, profile.SuggestedID) {
		profile.SuggestedID = ""
	}
	return profile
}

func hasIntentID(components []deployComponent, id string) bool {
	for _, intent := range detectDeployIntents(components) {
		if intent.ID == id {
			return true
		}
	}
	return false
}

func apiComponents(parts []deployComponent) []deployComponent {
	var out []deployComponent
	for _, c := range parts {
		if strings.HasPrefix(c.ID, "api:") || strings.HasPrefix(c.ID, "app:") {
			out = append(out, c)
		}
	}
	return out
}

func frontendComponents(parts []deployComponent) []deployComponent {
	var out []deployComponent
	for _, c := range parts {
		if strings.HasPrefix(c.ID, "frontend:") || strings.HasPrefix(c.ID, "site:") {
			out = append(out, c)
		}
	}
	return out
}

func cloudComponents(parts []deployComponent) []deployComponent {
	var out []deployComponent
	for _, c := range parts {
		if strings.HasPrefix(c.ID, "bundle:") {
			continue
		}
		out = append(out, c)
	}
	return out
}

func suggestComponentID(parts []deployComponent, prefixes ...string) string {
	for _, p := range prefixes {
		for _, c := range parts {
			if strings.HasPrefix(c.ID, p) {
				return c.ID
			}
		}
	}
	if len(parts) > 0 {
		return parts[0].ID
	}
	return ""
}

func hasComponentID(parts []deployComponent, id string) bool {
	for _, c := range parts {
		if c.ID == id {
			return true
		}
	}
	return false
}

func signalContains(signals []string, want string) bool {
	for _, s := range signals {
		if s == want {
			return true
		}
	}
	return false
}

func runFileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func reorderSuggestedFirst(parts []deployComponent, suggestedID string) []deployComponent {
	if suggestedID == "" {
		return parts
	}
	var suggested, rest []deployComponent
	for _, c := range parts {
		if c.ID == suggestedID {
			suggested = append(suggested, c)
		} else {
			rest = append(rest, c)
		}
	}
	if len(suggested) == 0 {
		return parts
	}
	return append(suggested, rest...)
}

func printProjectProfile(profile projectProfile, components []deployComponent) {
	if profile.Title == "" {
		return
	}
	fmt.Printf("%s %s\n", ui.label.Render("Project type:"), ui.title.Render(profile.Title))
	if profile.Summary != "" {
		fmt.Printf("  %s\n", ui.dim.Render(profile.Summary))
	}
	if profile.SuggestedID != "" && profile.Reason != "" {
		fmt.Println()
		label := suggestionLabel(profile, components)
		if profile.SuggestedID == "full_stack" {
			fmt.Printf("  %s %s\n", ui.info.Render("For a full release →"), ui.value.Render(label))
		} else {
			fmt.Printf("  %s %s\n", ui.info.Render("Suggested →"), ui.value.Render(label))
		}
		fmt.Printf("  %s\n", ui.dim.Render(profile.Reason))
	}
	if profile.Type == "desktop_app" {
		fmt.Println()
		fmt.Printf("  %s\n", ui.dim.Render("Use Signet for this repo: signet doctor → signet build"))
	}
}

func suggestionLabel(profile projectProfile, components []deployComponent) string {
	for _, intent := range detectDeployIntents(components) {
		if intent.ID == profile.SuggestedID {
			return intent.Label
		}
	}
	return profile.SuggestedID
}
