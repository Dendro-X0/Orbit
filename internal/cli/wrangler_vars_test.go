package cli

import (
	"os"
	"testing"
)

func TestReadWranglerCORSOrigins(t *testing.T) {
	content := `[vars]
CORS_ORIGINS = "http://localhost:5173,https://docs.example.com"
`
	dir := t.TempDir()
	path := dir + "/wrangler.toml"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	origins := readWranglerCORSOrigins(path)
	if len(origins) != 2 {
		t.Fatalf("origins = %v", origins)
	}
}

func TestCorsOnlyLocalDev(t *testing.T) {
	if !corsOnlyLocalDev([]string{"http://localhost:5173", "http://localhost:8787"}) {
		t.Fatal("expected localhost only")
	}
	if corsOnlyLocalDev([]string{"http://localhost:5173", "https://docs.example.com"}) {
		t.Fatal("expected mixed origins")
	}
}

func TestCorsMissingOrigin(t *testing.T) {
	origins := []string{"http://localhost:5173"}
	if !corsMissingOrigin(origins, "https://docs.example.com") {
		t.Fatal("expected missing")
	}
	if corsMissingOrigin(origins, "http://localhost:5173") {
		t.Fatal("expected present")
	}
}
