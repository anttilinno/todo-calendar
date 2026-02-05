package todolist

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines key bindings for todo list operations.
type KeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Add      key.Binding
	AddDated key.Binding
	Toggle   key.Binding
	Delete   key.Binding
	Edit     key.Binding
	EditDate key.Binding
	Confirm  key.Binding
	Cancel   key.Binding
}

// ShortHelp returns key bindings for the short help view.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Add, k.AddDated, k.Toggle, k.Delete, k.Edit, k.EditDate}
}

// FullHelp returns key bindings for the full help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Add, k.AddDated, k.Toggle, k.Delete, k.Edit, k.EditDate},
	}
}

// DefaultKeyMap returns the default todo list key bindings.
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
		Add: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add todo"),
		),
		AddDated: key.NewBinding(
			key.WithKeys("A"),
			key.WithHelp("A", "add dated"),
		),
		Toggle: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "complete"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
		EditDate: key.NewBinding(
			key.WithKeys("E"),
			key.WithHelp("E", "edit date"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
	}
}
