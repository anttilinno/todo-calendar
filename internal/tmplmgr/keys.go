package tmplmgr

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines key bindings for the template management overlay.
type KeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Create   key.Binding
	Delete   key.Binding
	Rename   key.Binding
	Edit     key.Binding
	Confirm  key.Binding
	Cancel   key.Binding
	Schedule key.Binding
	Left     key.Binding
	Right    key.Binding
	Toggle   key.Binding
}

// ShortHelp returns key bindings for the short help view.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Create, k.Delete, k.Rename, k.Edit, k.Schedule, k.Cancel}
}

// FullHelp returns key bindings for the full help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}

// DefaultKeyMap returns the default template management key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/up", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/dn", "down"),
		),
		Create: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "new"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
		Rename: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "rename"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "close"),
		),
		Schedule: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "schedule"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("left/h", "prev type"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("right/l", "next type"),
		),
		Toggle: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "toggle"),
		),
	}
}
