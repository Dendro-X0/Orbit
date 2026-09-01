package provider

import "fmt"

var registry = map[string]Provider{}

func Register(p Provider) {
	registry[p.ID()] = p
}

func Get(id string) (Provider, error) {
	p, ok := registry[id]
	if !ok {
		return nil, fmt.Errorf("unknown provider %q", id)
	}
	return p, nil
}

func All() []Provider {
	out := make([]Provider, 0, len(registry))
	for _, p := range registry {
		out = append(out, p)
	}
	return out
}

func IDs() []string {
	out := make([]string, 0, len(registry))
	for id := range registry {
		out = append(out, id)
	}
	return out
}
