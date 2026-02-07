package preview

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"github.com/antti/todo-calendar/internal/theme"
)

// CloseMsg is emitted when the user closes the preview overlay.
type CloseMsg struct{}

// Model represents the markdown preview overlay.
type Model struct {
	title     string
	viewport  viewport.Model
	renderer  *glamour.TermRenderer
	rawBody   string
	width     int
	height    int
	keys      KeyMap
	styles    Styles
	themeName string
}

// New creates a new preview model for the given todo title and body.
func New(title, body, themeName string, t theme.Theme, width, height int) Model {
	s := NewStyles(t)
	keys := DefaultKeyMap()

	contentWidth := width - 4
	if contentWidth < 10 {
		contentWidth = 10
	}
	contentHeight := height - 4
	if contentHeight < 1 {
		contentHeight = 1
	}

	renderer, err := NewMarkdownRenderer(themeName, contentWidth)

	var rendered string
	if body == "" {
		rendered = lipgloss.NewStyle().Foreground(t.MutedFg).Render("(no body)")
	} else if err != nil {
		rendered = body
	} else {
		rendered, err = renderer.Render(body)
		if err != nil {
			rendered = body
		}
		rendered = strings.TrimRight(rendered, "\n")
	}

	vp := viewport.New(contentWidth, contentHeight)
	vp.SetContent(rendered)

	return Model{
		title:     title,
		viewport:  vp,
		renderer:  renderer,
		rawBody:   body,
		width:     width,
		height:    height,
		keys:      keys,
		styles:    s,
		themeName: themeName,
	}
}

// Update handles messages for the preview model.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, m.keys.Close) {
			return m, func() tea.Msg { return CloseMsg{} }
		}
		// Forward to viewport for scrolling
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.rebuildContent()
		return m, nil
	}

	return m, nil
}

// View renders the preview overlay.
func (m Model) View() string {
	var b strings.Builder

	titleBar := m.styles.Title.Render(m.title)
	b.WriteString(titleBar)
	b.WriteString("\n")

	b.WriteString(m.viewport.View())

	return m.styles.Border.
		Width(m.width - 2).
		Height(m.height - 2).
		Render(b.String())
}

// SetTheme updates the styles and rebuilds the renderer for the new theme.
func (m *Model) SetTheme(t theme.Theme) {
	m.styles = NewStyles(t)
	m.rebuildContent()
}

// HelpBindings returns preview key bindings for the help bar.
func (m Model) HelpBindings() []key.Binding {
	return HelpBindings()
}

// rebuildContent recreates the renderer and viewport content for the current dimensions.
func (m *Model) rebuildContent() {
	contentWidth := m.width - 4
	if contentWidth < 10 {
		contentWidth = 10
	}
	contentHeight := m.height - 4
	if contentHeight < 1 {
		contentHeight = 1
	}

	renderer, err := NewMarkdownRenderer(m.themeName, contentWidth)
	if err == nil {
		m.renderer = renderer
	}

	var rendered string
	if m.rawBody == "" {
		rendered = "(no body)"
	} else if m.renderer != nil {
		rendered, err = m.renderer.Render(m.rawBody)
		if err != nil {
			rendered = m.rawBody
		}
		rendered = strings.TrimRight(rendered, "\n")
	} else {
		rendered = m.rawBody
	}

	m.viewport = viewport.New(contentWidth, contentHeight)
	m.viewport.SetContent(rendered)
}
