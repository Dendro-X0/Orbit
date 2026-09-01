package cloudflare

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSecretNames(t *testing.T) {
	path := filepath.Join(t.TempDir(), "wrangler.toml")
	content := `name = "test"
# Secrets (set via wrangler secret put):
# API_KEY_PEPPER, GITHUB_TOKEN, POLAR_WEBHOOK_SECRET
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	names, err := ParseSecretNames(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 3 {
		t.Fatalf("names = %#v", names)
	}
}

func TestParseSecretNamesEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "wrangler.toml")
	if err := os.WriteFile(path, []byte(`name = "test"`), 0o644); err != nil {
		t.Fatal(err)
	}
	names, err := ParseSecretNames(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 0 {
		t.Fatalf("names = %#v", names)
	}
}
