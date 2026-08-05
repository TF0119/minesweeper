package ui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/takeru0119/minesweeper/internal/game"
	"github.com/takeru0119/minesweeper/internal/storage"
	"golang.org/x/term"
)

// Options configures the UI run.
type Options struct {
	Difficulty game.Difficulty
	Config     storage.Config
	HighScores storage.HighScores
	NoColor    bool
}

// Run starts the TUI program.
func Run(opts Options) error {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return fmt.Errorf("minesweeper requires an interactive terminal")
	}

	m := NewModel(opts)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}
