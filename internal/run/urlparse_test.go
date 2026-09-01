package run

import "testing"

func TestExtractWorkersURL(t *testing.T) {
	log := `Uploaded assess-api
Deployed to https://assess-api.acme.workers.dev
Current Version ID: abc`
	if got := ExtractWorkersURL(log); got != "https://assess-api.acme.workers.dev" {
		t.Fatalf("got %q", got)
	}
}

func TestExtractWorkersURLLastMatch(t *testing.T) {
	log := `preview https://old.dev.workers.dev
final https://assess-api.acme.workers.dev`
	if got := ExtractWorkersURL(log); got != "https://assess-api.acme.workers.dev" {
		t.Fatalf("got %q", got)
	}
}
