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
	stats      storage.Stats

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
		board:      newBoard(d, opts.Seed, opts.Config.NoGuess),
		cursor:     game.Coord{X: d.Width / 2, Y: d.Height / 2},
		screen:     ScreenPlaying,
		difficulty: d,
		config:     opts.Config,
		highscores: opts.HighScores,
		stats:      opts.Stats,
		dailySeed:  game.DailySeed(timeNow()),
		vp:         fit(viewport{}, 0, 0, d.Width, d.Height),
		styles:     NewStyles(themeFromConfig(opts.Config), !opts.NoColor),
		glyphs:     newGlyphs(opts.Config.UseEmoji),
		keys:       DefaultKeyMap(),
	}
}

// themeFromConfig resolves the saved theme name. A name this build does not
// know is treated as classic rather than an error: a config file written by a
// newer version should not stop the game from starting.
func themeFromConfig(c storage.Config) Theme {
	theme, _ := ParseTheme(c.Theme)
	return theme
}

// newBoard builds the kind of board the player asked for. No-guess boards cost
// a search on the opening move, so they are only built when requested.
func newBoard(d game.Difficulty, seed game.Seed, noGuess bool) *game.Board {
	if noGuess {
		return game.NewNoGuessBoard(d, seed)
	}
	return game.NewBoard(d, seed)
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
