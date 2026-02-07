package todolist

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines key bindings for todo list operations.
type KeyMap struct {
	Up             key.Binding
	Down           key.Binding
	MoveUp         key.Binding
	MoveDown       key.Binding
	Add            key.Binding
	AddDated       key.Binding
	Toggle         key.Binding
	Delete         key.Binding
	Edit           key.Binding
	EditDate       key.Binding
	Filter         key.Binding
	Preview        key.Binding
	OpenEditor     key.Binding
	TemplateUse    key.Binding
	TemplateCreate key.Binding
	Confirm        key.Binding
	Cancel         key.Binding
	SwitchField    key.Binding
}

// ShortHelp returns key bindings for the short help view.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.MoveUp, k.MoveDown, k.Add, k.AddDated, k.Toggle, k.Delete, k.Edit, k.EditDate, k.Filter, k.Preview, k.OpenEditor, k.TemplateUse, k.TemplateCreate, k.SwitchField}
}

// FullHelp returns key bindings for the full help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.MoveUp, k.MoveDown, k.Add, k.AddDated, k.Toggle, k.Delete, k.Edit, k.EditDate, k.Filter, k.Preview, k.OpenEditor, k.TemplateUse, k.TemplateCreate, k.SwitchField},
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
		MoveUp: key.NewBinding(
			key.WithKeys("K"),
			key.WithHelp("K", "move up"),
		),
		MoveDown: key.NewBinding(
			key.WithKeys("J"),
			key.WithHelp("J", "move down"),
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
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		Preview: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "preview"),
		),
		OpenEditor: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "open editor"),
		),
		TemplateUse: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "from template"),
		),
		TemplateCreate: key.NewBinding(
			key.WithKeys("T"),
			key.WithHelp("T", "new template"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		SwitchField: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch field"),
		),
	}
}
