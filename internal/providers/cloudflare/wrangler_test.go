package cloudflare

import (
	"os"
	"testing"
)

func TestParseDatabaseIDFromCreateOutput(t *testing.T) {
	out := `Created your new D1 database.

[[d1_databases]]
binding = "DB"
database_name = "assess-db"
database_id = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
`
	id, err := parseDatabaseIDFromCreateOutput(out)
	if err != nil {
		t.Fatal(err)
	}
	if id != "a1b2c3d4-e5f6-7890-abcd-ef1234567890" {
		t.Fatalf("got %q", id)
	}
}

func TestPatchDatabaseID(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/wrangler.toml"
	original := `[[d1_databases]]
binding = "DB"
database_name = "assess-db"
database_id = "REPLACE_AFTER_CREATE"
`
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := patchDatabaseID(path, "a1b2c3d4-e5f6-7890-abcd-ef1234567890"); err != nil {
		t.Fatal(err)
	}
	cfg, err := readD1Config(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DatabaseID != "a1b2c3d4-e5f6-7890-abcd-ef1234567890" {
		t.Fatalf("got %q", cfg.DatabaseID)
	}
}

func TestNeedsD1Link(t *testing.T) {
	if !needsD1Link("REPLACE_AFTER_CREATE") {
		t.Fatal("expected true")
	}
	if needsD1Link("a1b2c3d4-e5f6-7890-abcd-ef1234567890") {
		t.Fatal("expected false")
	}
}

func TestParseWhoAmIJSON(t *testing.T) {
	out := `{"loggedIn":true,"email":"user@example.com","accounts":[{"name":"User Account"}]}`
	loggedIn, account := parseWhoAmI(out)
	if !loggedIn || account != "user@example.com" {
		t.Fatalf("got loggedIn=%v account=%q", loggedIn, account)
	}
}

func TestParseWhoAmINotLoggedIn(t *testing.T) {
	out := `{"loggedIn":false}`
	loggedIn, account := parseWhoAmI(out)
	if loggedIn || account != "" {
		t.Fatalf("got loggedIn=%v account=%q", loggedIn, account)
	}
}

func TestParseWhoAmIPlainText(t *testing.T) {
	out := "You are logged in with an OAuth Token, associated with the email user@example.com."
	loggedIn, account := parseWhoAmI(out)
	if !loggedIn || account != "user@example.com" {
		t.Fatalf("got loggedIn=%v account=%q", loggedIn, account)
	}
}
