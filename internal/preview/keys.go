package preview

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines key bindings for the preview overlay.
type KeyMap struct {
	Up       key.Binding
	Down     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Close    key.Binding
}

// DefaultKeyMap returns the default preview key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j", "down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "b"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "f"),
			key.WithHelp("pgdn", "page down"),
		),
		Close: key.NewBinding(
			key.WithKeys("esc", "q"),
			key.WithHelp("esc", "close"),
		),
	}
}

// HelpBindings returns preview key bindings for help bar display.
func HelpBindings() []key.Binding {
	km := DefaultKeyMap()
	return []key.Binding{km.Up, km.Down, km.PageUp, km.PageDown, km.Close}
}
