package vercel

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindVercelTargets(t *testing.T) {
	root := t.TempDir()
	docs := filepath.Join(root, "apps", "docs")
	if err := os.MkdirAll(docs, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(docs, "vercel.json"), []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(docs, "vite.config.ts"), []byte(`export default {}`), 0o644); err != nil {
		t.Fatal(err)
	}

	targets, err := findVercelTargets(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 1 {
		t.Fatalf("targets = %d, want 1", len(targets))
	}
	if targets[0].ID != "docs" || targets[0].Kind != "vite" {
		t.Fatalf("got %+v", targets[0])
	}
}

func TestRequiredViteEnvVars(t *testing.T) {
	root := t.TempDir()
	example := `VITE_API_URL=http://localhost:8787
VITE_POLAR_CHECKOUT_URL=
# comment
OTHER=1
`
	if err := os.WriteFile(filepath.Join(root, ".env.example"), []byte(example), 0o644); err != nil {
		t.Fatal(err)
	}
	vars := requiredViteEnvVars(root, ".")
	if len(vars) != 2 {
		t.Fatalf("vars = %v", vars)
	}
}

func TestParseEnvList(t *testing.T) {
	out := ` name               value               environments
 VITE_API_URL       https://api.example  Production
`
	found := parseEnvList(out)
	if !found["VITE_API_URL"] {
		t.Fatal("expected VITE_API_URL")
	}
}

func TestReadProjectLink(t *testing.T) {
	root := t.TempDir()
	vercelDir := filepath.Join(root, ".vercel")
	if err := os.MkdirAll(vercelDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(vercelDir, "project.json"), []byte(`{"projectId":"prj_123","orgId":"team_456"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	link, ok, err := readProjectLink(root, ".")
	if err != nil || !ok || link.ProjectID != "prj_123" {
		t.Fatalf("link=%+v ok=%v err=%v", link, ok, err)
	}
}
