package cloudflare

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	d1NameRe      = regexp.MustCompile(`(?m)^\s*database_name\s*=\s*"([^"]+)"`)
	d1IDRe        = regexp.MustCompile(`(?m)^\s*database_id\s*=\s*"([^"]+)"`)
	uuidRe        = regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
	whoAmIEmailRe = regexp.MustCompile(`associated with the email\s+(\S+)`)
)

type wranglerWhoAmI struct {
	LoggedIn bool   `json:"loggedIn"`
	Email    string `json:"email"`
	Accounts []struct {
		Name string `json:"name"`
	} `json:"accounts"`
}

type d1Config struct {
	DatabaseName string
	DatabaseID   string
}

func readD1Config(wranglerPath string) (d1Config, error) {
	content, err := os.ReadFile(wranglerPath)
	if err != nil {
		return d1Config{}, err
	}
	text := string(content)
	nameMatch := d1NameRe.FindStringSubmatch(text)
	idMatch := d1IDRe.FindStringSubmatch(text)
	if len(nameMatch) < 2 {
		return d1Config{}, fmt.Errorf("database_name not found in %s", wranglerPath)
	}
	cfg := d1Config{DatabaseName: nameMatch[1]}
	if len(idMatch) >= 2 {
		cfg.DatabaseID = idMatch[1]
	}
	return cfg, nil
}

func patchDatabaseID(wranglerPath, databaseID string) error {
	content, err := os.ReadFile(wranglerPath)
	if err != nil {
		return err
	}
	text := string(content)
	if !d1IDRe.MatchString(text) {
		return fmt.Errorf("database_id not found in %s", wranglerPath)
	}
	updated := d1IDRe.ReplaceAllString(text, fmt.Sprintf(`database_id = "%s"`, databaseID))
	return os.WriteFile(wranglerPath, []byte(updated), 0o644)
}

func parseDatabaseIDFromCreateOutput(output string) (string, error) {
	match := uuidRe.FindString(output)
	if match == "" {
		return "", fmt.Errorf("could not parse database_id from wrangler output")
	}
	return match, nil
}

func needsD1Link(databaseID string) bool {
	id := strings.TrimSpace(databaseID)
	return id == "" || strings.Contains(id, "REPLACE")
}

// parseWhoAmI extracts login status and a short account label from wrangler whoami output.
func parseWhoAmI(output string) (loggedIn bool, account string) {
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return false, ""
	}

	var parsed wranglerWhoAmI
	if err := json.Unmarshal([]byte(trimmed), &parsed); err == nil {
		if !parsed.LoggedIn {
			return false, ""
		}
		if parsed.Email != "" {
			return true, parsed.Email
		}
		if len(parsed.Accounts) > 0 && parsed.Accounts[0].Name != "" {
			return true, parsed.Accounts[0].Name
		}
		return true, "logged in"
	}

	if match := whoAmIEmailRe.FindStringSubmatch(trimmed); len(match) >= 2 {
		return true, strings.TrimSuffix(match[1], ".")
	}
	if strings.Contains(strings.ToLower(trimmed), "not logged in") {
		return false, ""
	}
	// Legacy fallback: treat non-empty output as logged in but keep it short.
	if len(trimmed) > 120 {
		return true, "logged in"
	}
	return true, trimmed
}

func wranglerPath(root, targetPath string) string {
	return filepath.Join(root, targetPath, "wrangler.toml")
}
