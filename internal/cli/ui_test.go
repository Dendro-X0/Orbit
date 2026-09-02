package cli

import (
	"os"
	"strings"
	"testing"
)

func TestHighlightCmdLine(t *testing.T) {
	got := highlightCmdLine("orbit login vercel")
	if !strings.Contains(got, "orbit") {
		t.Fatalf("got %q", got)
	}
}

func TestStyledURL(t *testing.T) {
	os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")
	got := styledURL("https://example.workers.dev")
	if got != "https://example.workers.dev" {
		t.Fatalf("NO_COLOR should pass through: %q", got)
	}
}

func TestColorsEnabledRespectsNoColor(t *testing.T) {
	os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")
	if colorsEnabled() {
		t.Fatal("expected colors disabled with NO_COLOR")
	}
}
