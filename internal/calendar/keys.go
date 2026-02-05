package calendar

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines key bindings for calendar navigation.
type KeyMap struct {
	PrevMonth key.Binding
	NextMonth key.Binding
}

// ShortHelp returns key bindings for the short help view.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.PrevMonth, k.NextMonth}
}

// FullHelp returns key bindings for the full help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.PrevMonth, k.NextMonth},
	}
}

// DefaultKeyMap returns the default calendar key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		PrevMonth: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("<-/h", "prev month"),
		),
		NextMonth: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("->/l", "next month"),
		),
	}
}
