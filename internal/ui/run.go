package ui

import (
	"fmt"
	"os"

	"github.com/TF0119/minesweeper/internal/game"
	"github.com/TF0119/minesweeper/internal/storage"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

// Options configures the UI run.
type Options struct {
	Difficulty game.Difficulty
	Seed       game.Seed
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
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseAllMotion())
	_, err := p.Run()
	return err
}
