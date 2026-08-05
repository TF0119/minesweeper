package ui

import (
	"math/rand"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/takeru0119/minesweeper/internal/game"
	"github.com/takeru0119/minesweeper/internal/storage"
)

// Model is the bubbletea model.
type Model struct {
	board      *game.Board
	cursor     game.Coord
	screen     Screen
	difficulty game.Difficulty
	config     storage.Config
	highscores storage.HighScores
	elapsed    int
	timerActive bool
	styles     Styles
	keys       KeyMap
	width      int
	height     int
	errMsg     string
	menuIndex  int
	quitting   bool
	noColor    bool
	rng        *rand.Rand
}

type tickMsg struct{}

// NewModel creates a new UI model.
func NewModel(opts Options) Model {
	useColor := !opts.NoColor
	cfg := opts.Config
	d := opts.Difficulty
	b := game.NewBoard(d, rand.New(rand.NewSource(time.Now().UnixNano())))

	return Model{
		board:      b,
		cursor:     game.Coord{X: d.Width / 2, Y: d.Height / 2},
		screen:     ScreenPlaying,
		difficulty: d,
		config:     cfg,
		highscores: opts.HighScores,
		styles:     NewStyles(useColor),
		keys:       DefaultKeyMap(),
		noColor:    opts.NoColor,
		rng:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}
