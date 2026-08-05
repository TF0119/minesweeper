package ui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines key bindings.
type KeyMap struct {
	Up, Down, Left, Right key.Binding
	Reveal, Flag, Chord  key.Binding
	New, Difficulty, Help key.Binding
	Quit                  key.Binding
}

// DefaultKeyMap returns default bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "right"),
		),
		Reveal: key.NewBinding(
			key.WithKeys(" ", "enter"),
			key.WithHelp("space", "reveal"),
		),
		Flag: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "flag"),
		),
		Chord: key.NewBinding(
			key.WithKeys("c", "shift+enter"),
			key.WithHelp("c", "chord"),
		),
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new game"),
		),
		Difficulty: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "difficulty"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Reveal, k.Flag, k.Chord, k.Help, k.Quit}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Reveal, k.Flag, k.Chord},
		{k.New, k.Difficulty, k.Help, k.Quit},
	}
}
