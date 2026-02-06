package settings

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/antti/todo-calendar/internal/config"
	"github.com/antti/todo-calendar/internal/holidays"
	"github.com/antti/todo-calendar/internal/theme"
)

// option represents a single configurable setting with cycling values.
type option struct {
	label   string   // displayed label: "Theme", "Country", "First Day of Week"
	values  []string // config values: ["dark", "light", "nord", "solarized"]
	display []string // display values: ["Dark", "Light", "Nord", "Solarized"]
	index   int      // currently selected index
}

// ThemeChangedMsg is emitted when the user cycles the theme option.
// The parent model uses this to trigger live preview.
type ThemeChangedMsg struct {
	Theme theme.Theme
}

// SaveMsg is emitted when the user presses Enter to save settings.
type SaveMsg struct {
	Cfg config.Config
}

// CancelMsg is emitted when the user presses Escape to cancel settings.
type CancelMsg struct{}

// Model represents the settings overlay.
type Model struct {
	options []option
	cursor  int // which option row is selected (0, 1, or 2)
	width   int
	height  int
	keys    KeyMap
	styles  Styles
}

// New creates a new settings model from the current configuration.
func New(cfg config.Config, t theme.Theme) Model {
	themeNames := theme.Names()
	themeDisplay := make([]string, len(themeNames))
	for i, name := range themeNames {
		themeDisplay[i] = strings.ToUpper(name[:1]) + name[1:]
	}

	countries := holidays.SupportedCountries()
	countryDisplay := countryLabels(countries)

	dayValues := []string{"sunday", "monday"}
	dayDisplay := []string{"Sunday", "Monday"}

	return Model{
		options: []option{
			{label: "Theme", values: themeNames, display: themeDisplay, index: indexOf(themeNames, cfg.Theme)},
			{label: "Country", values: countries, display: countryDisplay, index: indexOf(countries, cfg.Country)},
			{label: "First Day of Week", values: dayValues, display: dayDisplay, index: indexOf(dayValues, cfg.FirstDayOfWeek)},
		},
		keys:   DefaultKeyMap(),
		styles: NewStyles(t),
	}
}

// Config returns a config.Config reflecting the current option selections.
func (m Model) Config() config.Config {
	return config.Config{
		Theme:          m.options[0].values[m.options[0].index],
		Country:        m.options[1].values[m.options[1].index],
		FirstDayOfWeek: m.options[2].values[m.options[2].index],
	}
}

// SetTheme replaces the styles with ones built from the given theme.
func (m *Model) SetTheme(t theme.Theme) {
	m.styles = NewStyles(t)
}

// SetSize stores width and height for centered rendering.
func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// Update handles messages for the settings overlay.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}

		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}

		case key.Matches(msg, m.keys.Left):
			opt := &m.options[m.cursor]
			opt.index--
			if opt.index < 0 {
				opt.index = len(opt.values) - 1
			}
			if m.cursor == 0 {
				newTheme := theme.ForName(opt.values[opt.index])
				return m, func() tea.Msg {
					return ThemeChangedMsg{Theme: newTheme}
				}
			}

		case key.Matches(msg, m.keys.Right):
			opt := &m.options[m.cursor]
			opt.index++
			if opt.index >= len(opt.values) {
				opt.index = 0
			}
			if m.cursor == 0 {
				newTheme := theme.ForName(opt.values[opt.index])
				return m, func() tea.Msg {
					return ThemeChangedMsg{Theme: newTheme}
				}
			}

		case key.Matches(msg, m.keys.Save):
			cfg := m.Config()
			return m, func() tea.Msg {
				return SaveMsg{Cfg: cfg}
			}

		case key.Matches(msg, m.keys.Cancel):
			return m, func() tea.Msg {
				return CancelMsg{}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// View renders the settings overlay.
func (m Model) View() string {
	var b strings.Builder

	title := m.styles.Title.Render("Settings")
	b.WriteString(title)
	b.WriteString("\n\n")

	for i, opt := range m.options {
		isSelected := i == m.cursor

		value := fmt.Sprintf("<  %s  >", opt.display[opt.index])

		if isSelected {
			label := m.styles.SelectedLabel.Render(fmt.Sprintf("> %-20s", opt.label))
			value = m.styles.SelectedValue.Render(value)
			b.WriteString(label + value + "\n")
		} else {
			label := m.styles.Label.Render(fmt.Sprintf("  %-20s", opt.label))
			value = m.styles.Value.Render(value)
			b.WriteString(label + value + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(m.styles.Hint.Render("  enter save  |  esc cancel  |  <-/-> change value"))

	content := b.String()

	// Center vertically if we have height information.
	if m.height > 0 {
		lines := strings.Count(content, "\n") + 1
		topPad := (m.height - lines) / 2
		if topPad > 0 {
			content = strings.Repeat("\n", topPad) + content
		}
	}

	return content
}

// HelpBindings returns settings-specific key bindings for help bar display.
func (m Model) HelpBindings() []key.Binding {
	return []key.Binding{m.keys.Left, m.keys.Right, m.keys.Up, m.keys.Down, m.keys.Save, m.keys.Cancel}
}

// indexOf returns the index of val in slice, or 0 if not found.
func indexOf(slice []string, val string) int {
	for i, v := range slice {
		if v == val {
			return i
		}
	}
	return 0
}

// countryNames maps country codes to display names.
var countryNames = map[string]string{
	"de": "Germany",
	"dk": "Denmark",
	"ee": "Estonia",
	"es": "Spain",
	"fi": "Finland",
	"fr": "France",
	"gb": "United Kingdom",
	"it": "Italy",
	"no": "Norway",
	"se": "Sweden",
	"us": "United States",
}

// countryLabels maps country codes to "XX - Country Name" display strings.
func countryLabels(codes []string) []string {
	labels := make([]string, len(codes))
	for i, code := range codes {
		name := countryNames[code]
		if name == "" {
			name = strings.ToUpper(code)
		}
		labels[i] = fmt.Sprintf("%s - %s", strings.ToUpper(code), name)
	}
	return labels
}
