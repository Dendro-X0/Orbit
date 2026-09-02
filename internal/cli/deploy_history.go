package cli

import (
	"fmt"
	"time"

	"github.com/Dendro-X0/Orbit/internal/run"
	"github.com/charmbracelet/huh"
)

func findPreviousDeploy(root string, selected []string) *run.DeployRecord {
	label := providerListLabel(selected)
	record, ok := run.LastSuccessfulDeploy(root, label)
	if !ok {
		return nil
	}
	return record
}

func printPreviousDeploy(record *run.DeployRecord) {
	if record == nil {
		return
	}
	fmt.Println()
	printSection("Already deployed")
	printKV("Provider", record.Provider)
	if !record.DeployedAt.IsZero() {
		printKV("When", formatTimeAgo(record.DeployedAt))
	}
	if record.Duration != "" {
		printKV("Duration", record.Duration)
	}
	if record.APIURL != "" {
		printKV("API URL", record.APIURL)
	}
	if record.DocsURL != "" {
		printKV("Docs URL", record.DocsURL)
	}
	if record.RunDir != "" {
		printKV("Run", record.RunDir)
	}
	printTip("This scope is already live — re-deploy only if you changed code or config.")
}

func formatTimeAgo(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}
	elapsed := time.Since(t)
	switch {
	case elapsed < time.Minute:
		return "just now"
	case elapsed < time.Hour:
		mins := int(elapsed.Round(time.Minute) / time.Minute)
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	case elapsed < 24*time.Hour:
		hours := int(elapsed.Round(time.Hour) / time.Hour)
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case elapsed < 48*time.Hour:
		return "yesterday"
	default:
		return t.Local().Format("2006-01-02 15:04")
	}
}

func confirmRedeploy(record *run.DeployRecord) (bool, error) {
	if record == nil {
		return true, nil
	}
	live := record.APIURL
	if live == "" {
		live = record.DocsURL
	}
	if live == "" {
		live = record.URL
	}

	desc := "Re-deploying overwrites the current release. This is usually unnecessary unless you changed code, secrets, or infrastructure."
	if live != "" {
		desc = fmt.Sprintf("Already live at %s\n\n%s", live, desc)
	}

	proceed := false
	if err := huh.NewConfirm().
		Title("Re-deploy to production?").
		Description(desc).
		Value(&proceed).
		Run(); err != nil {
		return false, err
	}
	return proceed, nil
}

func openDeployRecordURL(record *run.DeployRecord, target string) error {
	if record == nil {
		return fmt.Errorf("no previous deploy found for this scope")
	}
	url := ""
	switch target {
	case "docs":
		url = record.DocsURL
	case "api":
		url = record.APIURL
	default:
		if record.APIURL != "" {
			url = record.APIURL
		} else {
			url = record.DocsURL
		}
	}
	if url == "" {
		return fmt.Errorf("no %s URL in the last successful deploy for this scope", target)
	}
	fmt.Printf("Opening %s\n", styledURL(url))
	return openURL(url)
}
