package ui

import (
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/key"
)

// KeyMap defines key bindings.
type KeyMap struct {
	Up, Down, Left, Right    key.Binding
	Reveal, Flag, Chord      key.Binding
	New, Restart, Difficulty key.Binding
	Stats, Help              key.Binding
	Quit                     key.Binding
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
			key.WithHelp("space/enter", "reveal cell"),
		),
		Flag: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "mark cell (or shift+click)"),
		),
		Chord: key.NewBinding(
			key.WithKeys("c", "shift+enter"),
			key.WithHelp("c", "chord (auto-reveal)"),
		),
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new game (new seed)"),
		),
		Restart: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "restart same seed"),
		),
		Difficulty: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "difficulty"),
		),
		Stats: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "statistics"),
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

// bindings returns every binding in display order. The help overlay and the
// status bar both derive from this, so key changes stay in one place.
func (k KeyMap) bindings() []key.Binding {
	return []key.Binding{
		k.Up, k.Down, k.Left, k.Right,
		k.Reveal, k.Flag, k.Chord,
		k.New, k.Restart, k.Difficulty, k.Stats, k.Help, k.Quit,
	}
}

// helpLines renders one "keys  description" line per binding.
func (k KeyMap) helpLines() []string {
	bs := k.bindings()
	width := 0
	for _, b := range bs {
		if n := utf8.RuneCountInString(b.Help().Key); n > width {
			width = n
		}
	}
	lines := make([]string, 0, len(bs))
	for _, b := range bs {
		pad := strings.Repeat(" ", width-utf8.RuneCountInString(b.Help().Key))
		lines = append(lines, b.Help().Key+pad+"   "+b.Help().Desc)
	}
	return lines
}
