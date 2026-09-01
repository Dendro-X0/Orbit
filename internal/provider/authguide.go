package provider

// AuthGuide documents how to manually obtain an API token for a provider.
// Orbit cannot automate provider OAuth; it guides users to the right page instead.
type AuthGuide struct {
	ProviderID  string
	TokenLabel  string
	CreateURL   string
	DocsURL     string
	Permissions string
	Steps       []string
}

var authGuides = map[string]AuthGuide{}

// RegisterAuthGuide registers token setup instructions for a provider.
func RegisterAuthGuide(guide AuthGuide) {
	authGuides[guide.ProviderID] = guide
}

// AuthGuideFor returns setup instructions for a provider, if registered.
func AuthGuideFor(providerID string) (AuthGuide, bool) {
	g, ok := authGuides[providerID]
	return g, ok
}
