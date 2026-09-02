package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// ui holds semantic terminal styles for Orbit output.
var ui uiStyles

type uiStyles struct {
	title    lipgloss.Style
	section  lipgloss.Style
	label    lipgloss.Style
	value    lipgloss.Style
	dim      lipgloss.Style
	success  lipgloss.Style
	error    lipgloss.Style
	warn     lipgloss.Style
	info     lipgloss.Style
	url      lipgloss.Style
	cmd      lipgloss.Style
	provider lipgloss.Style
	path     lipgloss.Style
	tip      lipgloss.Style
}

func init() {
	enabled := colorsEnabled()
	if !enabled {
		lipgloss.SetColorProfile(termenv.Ascii)
	}

	ui = uiStyles{
		title: lipgloss.NewStyle().Bold(true),
		section: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")),
		label: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Bold(true),
		value: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		dim: lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")),
		success: lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true),
		error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true),
		warn: lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true),
		info: lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")),
		url: lipgloss.NewStyle().
			Foreground(lipgloss.Color("81")).
			Underline(true),
		cmd: lipgloss.NewStyle().
			Foreground(lipgloss.Color("141")).
			Bold(true),
		provider: lipgloss.NewStyle().
			Foreground(lipgloss.Color("43")).
			Bold(true),
		path: lipgloss.NewStyle().
			Foreground(lipgloss.Color("117")),
		tip: lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Italic(true),
	}
}

func colorsEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("FORCE_COLOR") != "" || os.Getenv("CLICOLOR_FORCE") != "" {
		return true
	}
	return isInteractive() || termenv.DefaultOutput().ColorProfile() != termenv.Ascii
}

func okMark() string  { return ui.success.Render("✓") }
func failMark() string { return ui.error.Render("✗") }
func warnMark() string { return ui.warn.Render("⚠") }

func styledURL(s string) string {
	if s == "" {
		return s
	}
	return ui.url.Render(s)
}

func styledCmd(s string) string {
	if s == "" {
		return s
	}
	return ui.cmd.Render(s)
}

func styledProvider(s string) string {
	if s == "" {
		return s
	}
	return ui.provider.Render(s)
}

func styledPath(s string) string {
	if s == "" {
		return s
	}
	return ui.path.Render(s)
}

func styledValue(s string) string {
	if s == "" {
		return s
	}
	lower := strings.ToLower(s)
	switch {
	case strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://"):
		return styledURL(s)
	case strings.HasPrefix(s, "orbit "):
		return highlightCmdLine(s)
	case lower == "production" || lower == "preview":
		return ui.info.Render(s)
	case strings.Contains(s, ".orbit/"):
		return styledPath(s)
	default:
		return ui.value.Render(s)
	}
}

func highlightCmdLine(s string) string {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return s
	}
	var parts []string
	for i, f := range fields {
		if f == "orbit" || (i > 0 && fields[i-1] == "orbit") || (i > 1 && fields[i-2] == "orbit") {
			parts = append(parts, ui.cmd.Render(f))
			continue
		}
		parts = append(parts, f)
	}
	return strings.Join(parts, " ")
}

func printTitle(title string) {
	fmt.Println(ui.title.Render(title))
}

func printSection(title string) {
	fmt.Println(ui.section.Render(title))
}

func printKV(key, value string) {
	fmt.Printf("  %s %s\n", ui.label.Render(key+":"), styledValue(value))
}

func printKVPlain(key, value string) {
	fmt.Printf("  %s %s\n", ui.label.Render(key+":"), ui.value.Render(value))
}

func printLabeled(width int, key, value string) {
	padded := ui.label.Render(key + ":")
	if width > 0 {
		fmt.Printf("%-*s %s\n", width, padded, styledValue(value))
		return
	}
	fmt.Printf("%s %s\n", padded, styledValue(value))
}

func printBullet(line string) {
	fmt.Printf("  %s %s\n", ui.dim.Render("•"), line)
}

func printBulletStyled(prefix, emphasis, suffix string) {
	fmt.Printf("  %s %s%s\n", ui.dim.Render("•"), emphasis, suffix)
}

func printTip(msg string) {
	fmt.Printf("  %s %s\n", ui.tip.Render("Tip:"), ui.dim.Render(msg))
}

func printSuccess(msg string) {
	fmt.Printf("%s %s\n", okMark(), ui.success.Render(msg))
}

func printWarning(msg string) {
	fmt.Printf("%s %s\n", warnMark(), ui.warn.Render(msg))
}

func printError(msg string) {
	fmt.Printf("%s %s\n", failMark(), ui.error.Render(msg))
}

func printIndentedSuccess(msg string) {
	fmt.Printf("  %s %s\n", okMark(), ui.success.Render(msg))
}

func printIndentedError(msg string) {
	fmt.Printf("  %s %s\n", failMark(), ui.error.Render(msg))
}

func printNextStep(cmd string) {
	fmt.Printf("  %s %s\n", ui.info.Render("→"), highlightCmdLine(cmd))
}
