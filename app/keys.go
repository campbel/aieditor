package app

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Top    key.Binding
	Bottom key.Binding
	Enter  key.Binding
	Focus  key.Binding
	Blur   key.Binding
	Save   key.Binding
	Load   key.Binding
	Undo   key.Binding
	Quit   key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Save, k.Quit}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter, k.Save, k.Quit},
	}
}

var Keys = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "down"),
	),
	Top: key.NewBinding(
		key.WithKeys("shift+up"),
		key.WithHelp("shift+↑", "top"),
	),
	Bottom: key.NewBinding(
		key.WithKeys("shift+down"),
		key.WithHelp("shift+↓", "bottom"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "submit"),
	),
	Focus: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "begin input"),
	),
	Blur: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel input"),
	),
	Save: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "save"),
	),
	Load: key.NewBinding(
		key.WithKeys("ctrl+l"),
		key.WithHelp("ctrl+l", "load"),
	),
	Undo: key.NewBinding(
		key.WithKeys("ctrl+z"),
		key.WithHelp("ctrl+z", "undo"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "exit"),
	),
}

type CompareKeysMap struct {
	Up     key.Binding
	Down   key.Binding
	Top    key.Binding
	Bottom key.Binding
	Accept key.Binding
	Reject key.Binding
	Next   key.Binding
	Prev   key.Binding
	Exit   key.Binding
	Retry  key.Binding
}

var CompareKeys = CompareKeysMap{
	Up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "down"),
	),
	Top: key.NewBinding(
		key.WithKeys("shift+up"),
		key.WithHelp("shift+↑", "top"),
	),
	Bottom: key.NewBinding(
		key.WithKeys("shift+down"),
		key.WithHelp("shift+↓", "bottom"),
	),
	Accept: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "accept"),
	),
	Reject: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "reject"),
	),
	Next: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next"),
	),
	Prev: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev"),
	),
	Exit: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "exit"),
	),
	Retry: key.NewBinding(
		key.WithKeys("ctrl+r"),
		key.WithHelp("ctrl+r", "retry"),
	),
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k CompareKeysMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Accept, k.Reject, k.Next, k.Prev, k.Retry}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k CompareKeysMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		k.ShortHelp(),
	}
}
