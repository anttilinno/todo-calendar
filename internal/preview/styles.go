package preview

import (
	"github.com/antti/todo-calendar/internal/theme"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/lipgloss"
)

// Styles holds themed lipgloss styles for the preview overlay.
type Styles struct {
	Title  lipgloss.Style
	Border lipgloss.Style
	Hint   lipgloss.Style
}

// NewStyles builds preview styles from the given theme.
func NewStyles(t theme.Theme) Styles {
	return Styles{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.AccentFg).
			Padding(0, 1),
		Border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.BorderFocused).
			Padding(0, 1),
		Hint: lipgloss.NewStyle().
			Foreground(t.MutedFg),
	}
}

// NewMarkdownRenderer creates a glamour TermRenderer that matches the app theme.
// The themeName determines whether to use a light or dark base style.
// Width controls word wrapping.
func NewMarkdownRenderer(themeName string, width int) (*glamour.TermRenderer, error) {
	var baseStyle ansi.StyleConfig
	switch themeName {
	case "light":
		baseStyle = styles.LightStyleConfig
	default:
		baseStyle = styles.DarkStyleConfig
	}

	// Zero out document margin -- the app handles its own padding via lipgloss.
	zero := uint(0)
	baseStyle.Document.Margin = &zero

	return glamour.NewTermRenderer(
		glamour.WithStyles(baseStyle),
		glamour.WithWordWrap(width),
	)
}
