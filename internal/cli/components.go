package cli

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/Dendro-X0/Orbit/internal/provider"
	"github.com/Dendro-X0/Orbit/internal/state"
	"github.com/charmbracelet/huh"
)

// deployComponent is a detected deploy target mapped to one provider.
type deployComponent struct {
	ID          string
	Label       string
	Description string
	Providers   []string
	Kind        string
}

// deployIntent is a user-facing project type (step 1 of ship).
type deployIntent struct {
	ID          string
	Label       string
	Description string
}

// fullStackPairing maps a full-stack choice to one or two providers (step 2).
type fullStackPairing struct {
	ID        string
	Label     string
	Providers []string
}

func detectDeployComponents(ctx context.Context, root string) []deployComponent {
	stack := detectStack(ctx, root)
	var parts []deployComponent

	for _, id := range stack {
		p, err := provider.Get(id)
		if err != nil {
			continue
		}
		det, err := p.Detect(ctx, root)
		if err != nil || !det.Supported {
			continue
		}
		for _, t := range det.Targets {
			part := componentForTarget(id, t)
			if part.ID != "" {
				parts = append(parts, part)
			}
		}
		if len(det.Targets) == 0 {
			if part := componentForProvider(id); part.ID != "" {
				parts = append(parts, part)
			}
		}
	}
	return parts
}

func componentForTarget(providerID string, t provider.Target) deployComponent {
	path := t.Path
	if path == "." {
		path = "project root"
	} else {
		path = strings.ReplaceAll(path, "\\", "/")
	}

	switch providerID {
	case "cloudflare":
		return deployComponent{
			ID:          "api:" + t.ID,
			Label:       "API / backend",
			Description: fmt.Sprintf("Edge API + database at %s", path),
			Providers:   []string{"cloudflare"},
			Kind:        t.Kind,
		}
	case "vercel":
		label, desc := vercelComponentLabel(t.Kind, path)
		return deployComponent{
			ID:          "frontend:" + t.ID,
			Label:       label,
			Description: desc,
			Providers:   []string{"vercel"},
			Kind:        t.Kind,
		}
	case "fly":
		label := "Container app"
		if t.Kind == "container" {
			label = "Container app / API"
		}
		return deployComponent{
			ID:          "app:" + t.ID,
			Label:       label,
			Description: fmt.Sprintf("Fly.io app at %s", path),
			Providers:   []string{"fly"},
			Kind:        t.Kind,
		}
	case "netlify":
		return deployComponent{
			ID:          "site:" + t.ID,
			Label:       "Static site / frontend",
			Description: fmt.Sprintf("Netlify site at %s", path),
			Providers:   []string{"netlify"},
			Kind:        t.Kind,
		}
	default:
		p, _ := provider.Get(providerID)
		name := providerID
		if p != nil {
			name = p.DisplayName()
		}
		return deployComponent{
			ID:          providerID + ":" + t.ID,
			Label:       name,
			Description: path,
			Providers:   []string{providerID},
			Kind:        t.Kind,
		}
	}
}

func componentForProvider(providerID string) deployComponent {
	p, err := provider.Get(providerID)
	if err != nil {
		return deployComponent{}
	}
	return deployComponent{
		ID:        providerID + ":root",
		Label:     p.DisplayName(),
		Providers: []string{providerID},
	}
}

func vercelComponentLabel(kind, path string) (label, desc string) {
	switch kind {
	case "nextjs":
		return "Web application (Next.js)", fmt.Sprintf("SSR app at %s", path)
	case "vite":
		return "Docs / web frontend", fmt.Sprintf("Vite app at %s", path)
	default:
		return "Static site / frontend", fmt.Sprintf("Site at %s", path)
	}
}

func detectDeployIntents(components []deployComponent) []deployIntent {
	apis := apiComponents(components)
	fronts := frontendComponents(components)
	var intents []deployIntent

	if len(apis) > 0 {
		intents = append(intents, deployIntent{
			ID:    "api_backend",
			Label: "API / backend",
		})
	}
	if len(fronts) > 0 {
		label := "Static site / docs"
		if fronts[0].Kind == "nextjs" {
			label = "Web application"
		}
		intents = append(intents, deployIntent{
			ID:    "web_frontend",
			Label: label,
		})
	}
	if len(apis) > 0 && len(fronts) > 0 {
		intents = append(intents, deployIntent{
			ID:    "full_stack",
			Label: "Full-stack",
		})
	}
	return enrichDeployIntents(components, intents)
}

func enrichDeployIntents(components []deployComponent, intents []deployIntent) []deployIntent {
	out := make([]deployIntent, len(intents))
	for i, intent := range intents {
		intent.Description = intentScopeDescription(intent.ID, components)
		out[i] = intent
	}
	return out
}

func intentScopeDescription(id string, components []deployComponent) string {
	switch id {
	case "api_backend":
		apis := apiComponents(components)
		if len(apis) == 0 {
			return "One provider"
		}
		if len(apis) == 1 {
			return fmt.Sprintf("One provider — %s only", providerDisplayName(apis[0].Providers[0]))
		}
		return fmt.Sprintf("One provider — pick from %d API target(s)", len(apis))
	case "web_frontend":
		fronts := frontendComponents(components)
		if len(fronts) == 0 {
			return "One provider"
		}
		if len(fronts) == 1 {
			return fmt.Sprintf("One provider — %s only", providerDisplayName(fronts[0].Providers[0]))
		}
		return fmt.Sprintf("One provider — pick from %d frontend target(s)", len(fronts))
	case "full_stack":
		pairings := buildFullStackPairings(components)
		if len(pairings) == 1 {
			n := len(pairings[0].Providers)
			if n == 1 {
				return "One provider — " + pairings[0].Label
			}
			return fmt.Sprintf("%d providers — %s", n, pairings[0].Label)
		}
		return "Multiple providers — backend and frontend together"
	default:
		return ""
	}
}

func orderDeployIntents(intents []deployIntent) []deployIntent {
	priority := map[string]int{
		"api_backend":  0,
		"web_frontend": 1,
		"full_stack":   2,
	}
	ordered := make([]deployIntent, len(intents))
	copy(ordered, intents)
	sort.Slice(ordered, func(i, j int) bool {
		return priority[ordered[i].ID] < priority[ordered[j].ID]
	})
	return ordered
}

func buildFullStackPairings(components []deployComponent) []fullStackPairing {
	apis := apiComponents(components)
	fronts := frontendComponents(components)

	// Next.js on Vercel with no separate API — one provider covers full-stack.
	if len(apis) == 0 {
		for _, front := range fronts {
			if front.Kind == "nextjs" {
				return []fullStackPairing{{
					ID:        "vercel:nextjs",
					Label:     "Vercel (Next.js full-stack)",
					Providers: []string{"vercel"},
				}}
			}
		}
	}

	var pairings []fullStackPairing
	for _, api := range apis {
		for _, front := range fronts {
			providers := appendUnique(nil, api.Providers[0], front.Providers[0])
			pairings = append(pairings, fullStackPairing{
				ID:        api.ID + "+" + front.ID,
				Label:     pairingLabel(api, front),
				Providers: providers,
			})
		}
	}
	return pairings
}

func pairingLabel(api, front deployComponent) string {
	bName := providerDisplayName(api.Providers[0])
	fName := providerDisplayName(front.Providers[0])
	return fmt.Sprintf("%s (API) + %s (frontend)", bName, fName)
}

func providerDisplayName(id string) string {
	p, err := provider.Get(id)
	if err != nil || p == nil {
		return id
	}
	return p.DisplayName()
}

func appendUnique(dst []string, ids ...string) []string {
	seen := make(map[string]bool, len(dst))
	for _, id := range dst {
		seen[id] = true
	}
	for _, id := range ids {
		if !seen[id] {
			dst = append(dst, id)
			seen[id] = true
		}
	}
	return dst
}

type providerOption struct {
	ID       string
	Label    string
	Provider string
}

func providerOptionsFromComponents(components []deployComponent) []providerOption {
	opts := make([]providerOption, 0, len(components))
	for _, c := range components {
		pid := c.Providers[0]
		label := fmt.Sprintf("%s on %s", c.Label, providerDisplayName(pid))
		if c.Description != "" {
			label += " — " + c.Description
		}
		opts = append(opts, providerOption{
			ID:       c.ID,
			Label:    label,
			Provider: pid,
		})
	}
	return opts
}

func reorderIntentsSuggestedFirst(intents []deployIntent, suggestedID string) []deployIntent {
	ordered := orderDeployIntents(intents)
	if suggestedID == "" {
		return ordered
	}
	var suggested, rest []deployIntent
	for _, intent := range ordered {
		if intent.ID == suggestedID {
			suggested = append(suggested, intent)
		} else {
			rest = append(rest, intent)
		}
	}
	if len(suggested) == 0 {
		return ordered
	}
	return append(suggested, rest...)
}

type shipSelection struct {
	Providers []string
	IntentID  string
	Label     string
}

// runShipDeploySelect: step 1 project type, step 2 provider(s).
func runShipDeploySelect(ctx context.Context, root string) (shipSelection, error) {
	components := detectDeployComponents(ctx, root)
	if len(components) == 0 {
		return shipSelection{}, fmt.Errorf("nothing deployable detected in this project")
	}

	profile := detectProjectProfile(root, components)
	printProjectProfile(profile, components)
	fmt.Println()

	intents := detectDeployIntents(components)
	if len(intents) == 0 {
		return shipSelection{}, fmt.Errorf("nothing deployable detected in this project")
	}

	intent, err := selectDeployIntent(intents, profile)
	if err != nil {
		return shipSelection{}, err
	}

	providers, detail, err := selectProvidersForIntent(intent, components)
	if err != nil {
		return shipSelection{}, err
	}

	label := formatScopeLabel(intent, providers, detail)
	fmt.Printf("\n%s %s\n", ui.label.Render("Deploying:"), ui.title.Render(label))
	printProviderAuthStatus(ctx, providers)
	fmt.Println()

	return shipSelection{
		Providers: providers,
		IntentID:  intent.ID,
		Label:     label,
	}, nil
}

func formatScopeLabel(intent deployIntent, providers []string, detail string) string {
	if detail != "" {
		return intent.Label + " — " + detail
	}
	if len(providers) == 1 {
		return intent.Label + " — " + providerDisplayName(providers[0])
	}
	if len(providers) > 1 {
		return intent.Label + " — " + providerListLabel(providers)
	}
	return intent.Label
}

func saveShipScope(root string, st *state.Project, sel shipSelection) error {
	st.SetShipScope(sel.IntentID, sel.Label, sel.Providers)
	return state.Save(statePath(root), *st)
}

func selectDeployIntent(intents []deployIntent, profile projectProfile) (deployIntent, error) {
	if len(intents) == 1 {
		fmt.Printf("%s %s\n", ui.label.Render("Project type:"), ui.title.Render(intents[0].Label))
		if intents[0].Description != "" {
			fmt.Printf("  %s\n", ui.dim.Render(intents[0].Description))
		}
		fmt.Println()
		return intents[0], nil
	}

	printDetectedComponents(intents, profile)

	ordered := orderDeployIntents(intents)
	choice := ordered[0].ID

	options := make([]huh.Option[string], 0, len(ordered))
	for _, intent := range ordered {
		label := intent.Label
		if intent.Description != "" {
			label += " — " + intent.Description
		}
		if intent.ID == profile.SuggestedID {
			label += " (suggested for full release)"
		}
		options = append(options, huh.NewOption(label, intent.ID))
	}

	desc := "API-only and frontend-only each need one provider. Full-stack needs two."
	if profile.SuggestedID == "full_stack" {
		desc = "Pick API / backend if you only need the API. Full-stack is for releasing both parts."
	}

	if err := huh.NewSelect[string]().
		Title("What are you deploying right now?").
		Description(desc).
		Options(options...).
		Value(&choice).
		Run(); err != nil {
		return deployIntent{}, err
	}

	for _, intent := range intents {
		if intent.ID == choice {
			return intent, nil
		}
	}
	return deployIntent{}, fmt.Errorf("unknown project type %q", choice)
}

func printDetectedComponents(intents []deployIntent, profile projectProfile) {
	fmt.Println(ui.label.Render("This repo can deploy:"))
	for _, intent := range orderDeployIntents(intents) {
		line := ui.value.Render(intent.Label)
		if intent.Description != "" {
			line += ui.dim.Render(" — ") + ui.dim.Render(intent.Description)
		}
		printBullet(line)
	}
	if profile.SuggestedID == "full_stack" {
		fmt.Println()
		printTip("choose API / backend to deploy only the API — no Vercel login needed")
	}
	fmt.Println()
}

func selectProvidersForIntent(intent deployIntent, components []deployComponent) ([]string, string, error) {
	switch intent.ID {
	case "api_backend":
		return selectOneProvider(apiComponents(components), "backend")
	case "web_frontend":
		return selectOneProvider(frontendComponents(components), "frontend")
	case "full_stack":
		return selectFullStackProviders(components)
	default:
		return nil, "", fmt.Errorf("unknown project type %q", intent.ID)
	}
}

func selectOneProvider(components []deployComponent, role string) ([]string, string, error) {
	opts := providerOptionsFromComponents(components)
	if len(opts) == 0 {
		return nil, "", fmt.Errorf("no %s providers detected", role)
	}
	if len(opts) == 1 {
		fmt.Printf("%s provider: %s\n", ui.label.Render(role), ui.value.Render(opts[0].Label))
		return []string{opts[0].Provider}, providerDisplayName(opts[0].Provider), nil
	}

	choice := opts[0].ID
	options := make([]huh.Option[string], 0, len(opts))
	for _, opt := range opts {
		options = append(options, huh.NewOption(opt.Label, opt.ID))
	}

	title := fmt.Sprintf("Choose %s provider", role)
	if err := huh.NewSelect[string]().
		Title(title).
		Description("One platform is enough for this project type.").
		Options(options...).
		Value(&choice).
		Run(); err != nil {
		return nil, "", err
	}

	for _, opt := range opts {
		if opt.ID == choice {
			return []string{opt.Provider}, providerDisplayName(opt.Provider), nil
		}
	}
	return nil, "", fmt.Errorf("unknown provider choice %q", choice)
}

func selectFullStackProviders(components []deployComponent) ([]string, string, error) {
	pairings := buildFullStackPairings(components)
	if len(pairings) == 0 {
		return nil, "", fmt.Errorf("no full-stack provider pairing detected")
	}
	if len(pairings) == 1 {
		p := pairings[0]
		if len(p.Providers) > 1 {
			proceed := false
			if err := huh.NewConfirm().
				Title(fmt.Sprintf("Deploy to %d providers?", len(p.Providers))).
				Description(fmt.Sprintf("%s\n\nChoose API / backend instead if you only need the API.", p.Label)).
				Value(&proceed).
				Run(); err != nil {
				return nil, "", err
			}
			if !proceed {
				return nil, "", fmt.Errorf("pick API / backend from the project type menu to deploy API only")
			}
		}
		fmt.Printf("%s %s\n", ui.label.Render("Providers:"), ui.value.Render(p.Label))
		return p.Providers, p.Label, nil
	}

	choice := pairings[0].ID
	options := make([]huh.Option[string], 0, len(pairings))
	for _, p := range pairings {
		options = append(options, huh.NewOption(p.Label, p.ID))
	}

	if err := huh.NewSelect[string]().
		Title("How should this full-stack app be deployed?").
		Description("Pick one provider for all-in-one, or a split across API and frontend.").
		Options(options...).
		Value(&choice).
		Run(); err != nil {
		return nil, "", err
	}

	for _, p := range pairings {
		if p.ID == choice {
			return p.Providers, p.Label, nil
		}
	}
	return nil, "", fmt.Errorf("unknown pairing %q", choice)
}
