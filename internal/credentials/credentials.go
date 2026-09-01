package credentials

import (
	"errors"
	"fmt"

	"github.com/zalando/go-keyring"
)

const service = "orbit"

var ErrNotFound = errors.New("credential not found")

var envKeys = map[string]string{
	"cloudflare": "CLOUDFLARE_API_TOKEN",
	"vercel":     "VERCEL_TOKEN",
}

// Set stores a provider API token in the OS keychain.
func Set(provider, token string) error {
	if token == "" {
		return fmt.Errorf("token is empty")
	}
	if envKeys[provider] == "" {
		return fmt.Errorf("unsupported provider %q", provider)
	}
	return keyring.Set(service, provider, token)
}

// Get returns a stored provider API token.
func Get(provider string) (string, error) {
	token, err := keyring.Get(service, provider)
	if errors.Is(err, keyring.ErrNotFound) {
		return "", ErrNotFound
	}
	return token, err
}

// Delete removes a stored provider API token.
func Delete(provider string) error {
	err := keyring.Delete(service, provider)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}

// Has reports whether a token is stored for the provider.
func Has(provider string) bool {
	_, err := Get(provider)
	return err == nil
}

// Env returns subprocess environment entries for a stored token.
func Env(provider string) ([]string, error) {
	key, ok := envKeys[provider]
	if !ok {
		return nil, nil
	}
	token, err := Get(provider)
	if errors.Is(err, ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return []string{key + "=" + token}, nil
}
