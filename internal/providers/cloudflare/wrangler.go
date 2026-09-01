package cloudflare

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	d1NameRe = regexp.MustCompile(`(?m)^\s*database_name\s*=\s*"([^"]+)"`)
	d1IDRe   = regexp.MustCompile(`(?m)^\s*database_id\s*=\s*"([^"]+)"`)
	uuidRe   = regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
)

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

func wranglerPath(root, targetPath string) string {
	return filepath.Join(root, targetPath, "wrangler.toml")
}
