package credentials

import (
	"testing"

	"github.com/zalando/go-keyring"
)

func TestMain(m *testing.M) {
	keyring.MockInit()
	m.Run()
}

func TestSetGetDelete(t *testing.T) {
	if err := Delete("cloudflare"); err != nil {
		t.Fatal(err)
	}

	if err := Set("cloudflare", "test-token"); err != nil {
		t.Fatal(err)
	}
	if !Has("cloudflare") {
		t.Fatal("expected credential to exist")
	}

	got, err := Get("cloudflare")
	if err != nil {
		t.Fatal(err)
	}
	if got != "test-token" {
		t.Fatalf("got %q", got)
	}

	env, err := Env("cloudflare")
	if err != nil {
		t.Fatal(err)
	}
	if len(env) != 1 || env[0] != "CLOUDFLARE_API_TOKEN=test-token" {
		t.Fatalf("env = %#v", env)
	}

	if err := Delete("cloudflare"); err != nil {
		t.Fatal(err)
	}
	if Has("cloudflare") {
		t.Fatal("expected credential to be deleted")
	}
}

func TestSetUnsupportedProvider(t *testing.T) {
	if err := Set("railway", "token"); err == nil {
		t.Fatal("expected error for unsupported provider")
	}
}

func TestSetFlyToken(t *testing.T) {
	if err := Delete("fly"); err != nil {
		t.Fatal(err)
	}
	if err := Set("fly", "fly-token"); err != nil {
		t.Fatal(err)
	}
	env, err := Env("fly")
	if err != nil {
		t.Fatal(err)
	}
	if len(env) != 1 || env[0] != "FLY_API_TOKEN=fly-token" {
		t.Fatalf("env = %#v", env)
	}
	if err := Delete("fly"); err != nil {
		t.Fatal(err)
	}
}

func TestSetNetlifyToken(t *testing.T) {
	if err := Delete("netlify"); err != nil {
		t.Fatal(err)
	}
	if err := Set("netlify", "netlify-token"); err != nil {
		t.Fatal(err)
	}
	env, err := Env("netlify")
	if err != nil {
		t.Fatal(err)
	}
	if len(env) != 1 || env[0] != "NETLIFY_AUTH_TOKEN=netlify-token" {
		t.Fatalf("env = %#v", env)
	}
	if err := Delete("netlify"); err != nil {
		t.Fatal(err)
	}
}
