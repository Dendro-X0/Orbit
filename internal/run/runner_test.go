package run

import (
	"fmt"
	"testing"
)

func TestClassifyErrorAuthWithProvider(t *testing.T) {
	hint := classifyError(fmt.Errorf("wrangler whoami: not logged in"), "cloudflare-whoami")
	if hint.Code != "auth.required" {
		t.Fatalf("code %q", hint.Code)
	}
	if hint.Action != "Run: orbit login cloudflare" {
		t.Fatalf("action %q", hint.Action)
	}
}
