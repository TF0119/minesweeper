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

	opts.NoColor = opts.NoColor || colorDisabledByEnv(os.Getenv)

	m := NewModel(opts)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseAllMotion())
	_, err := p.Run()
	return err
}

// colorDisabledByEnv reports whether the environment opts out of colour.
//
// lipgloss already drops colour on its own in these cases, but it keeps
// applying bold and reverse. Detecting the opt-out here lets us switch to the
// monochrome styles, which carry state through those attributes instead — most
// importantly the cursor, which would otherwise lose its highlight entirely.
func colorDisabledByEnv(getenv func(string) string) bool {
	return getenv("NO_COLOR") != "" || getenv("TERM") == "dumb"
}
