package cli

import "testing"

func TestVersionDefaults(t *testing.T) {
	if Version == "" {
		t.Fatal("Version should not be empty")
	}
}
