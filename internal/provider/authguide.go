package provider

// AuthGuide documents how to authenticate with a provider.
// OAuthSteps describe the browser flow; Steps/CreateURL are for manual API tokens (--guide).
type AuthGuide struct {
	ProviderID  string
	TokenLabel  string
	CreateURL   string
	DocsURL     string
	Permissions string
	Steps       []string
	OAuthSteps  []string
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
