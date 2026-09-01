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

func TestExtractVercelURL(t *testing.T) {
	log := `Production: https://assess-docs.vercel.app [copied to clipboard]
Inspect: https://vercel.com/...`
	if got := ExtractVercelURL(log); got != "https://assess-docs.vercel.app" {
		t.Fatalf("got %q", got)
	}
}

func TestExtractDeployURLs(t *testing.T) {
	log := `Deployed to https://assess-api.acme.workers.dev
https://assess-docs.vercel.app`
	urls := ExtractDeployURLs(log)
	if urls.API != "https://assess-api.acme.workers.dev" {
		t.Fatalf("api %q", urls.API)
	}
	if urls.Docs != "https://assess-docs.vercel.app" {
		t.Fatalf("docs %q", urls.Docs)
	}
}

func TestProviderFromStepID(t *testing.T) {
	if got := providerFromStepID("cloudflare-whoami"); got != "cloudflare" {
		t.Fatalf("got %q", got)
	}
	if got := providerFromStepID("wire-vite-api-url"); got != "vercel" {
		t.Fatalf("got %q", got)
	}
}
