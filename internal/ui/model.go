package ui

import (
	"time"

	"github.com/TF0119/minesweeper/internal/game"
	"github.com/TF0119/minesweeper/internal/storage"
	tea "github.com/charmbracelet/bubbletea"
)

// Model is the bubbletea model.
type Model struct {
	board      *game.Board
	cursor     game.Coord
	screen     Screen
	difficulty game.Difficulty
	config     storage.Config
	highscores storage.HighScores

	dailySeed game.Seed // today's challenge, for labelling the current board

	elapsed     int
	timerActive bool
	timerStart  time.Time

	vp     viewport
	width  int
	height int

	styles Styles
	glyphs glyphs
	keys   KeyMap

	errMsg       string
	menuIndex    int
	quitting     bool
	lastMouseBtn tea.MouseButton
}

type tickMsg struct{}

// timeNow is a variable so tests can pin the clock used for daily labelling.
var timeNow = time.Now

// NewModel creates a new UI model.
func NewModel(opts Options) Model {
	d := opts.Difficulty
	return Model{
		board:      game.NewBoard(d, opts.Seed),
		cursor:     game.Coord{X: d.Width / 2, Y: d.Height / 2},
		screen:     ScreenPlaying,
		difficulty: d,
		config:     opts.Config,
		highscores: opts.HighScores,
		dailySeed:  game.DailySeed(timeNow()),
		vp:         fit(viewport{}, 0, 0, d.Width, d.Height),
		styles:     NewStyles(!opts.NoColor),
		glyphs:     newGlyphs(opts.Config.UseEmoji),
		keys:       DefaultKeyMap(),
	}
}

// isDailyBoard reports whether the current board is today's daily challenge.
func (m Model) isDailyBoard() bool {
	return m.board.Seed() == m.dailySeed
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}
