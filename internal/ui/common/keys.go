package common

import "charm.land/bubbles/v2/key"

// KeyMap defines all application keybindings.
type KeyMap struct {
	Quit      key.Binding
	View1     key.Binding
	View2     key.Binding
	View3     key.Binding
	View4     key.Binding
	Tab       key.Binding
	Add       key.Binding
	Delete    key.Binding
	Edit      key.Binding
	Search    key.Binding
	Help      key.Binding
	Up        key.Binding
	Down      key.Binding
	Enter     key.Binding
	Escape    key.Binding
	Refresh   key.Binding
	Sort      key.Binding
	Filter    key.Binding
	ToggleDone key.Binding
}

// DefaultKeyMap returns the default keybinding set.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		View1: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "portfolio"),
		),
		View2: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "watchlist"),
		),
		View3: key.NewBinding(
			key.WithKeys("3"),
			key.WithHelp("3", "charts"),
		),
		View4: key.NewBinding(
			key.WithKeys("4"),
			key.WithHelp("4", "news"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "sidebar"),
		),
		Add: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Sort: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "sort"),
		),
		Filter: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "filter"),
		),
		ToggleDone: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "toggle done"),
		),
	}
}
