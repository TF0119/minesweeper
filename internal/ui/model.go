package ui

import (
	"time"

	"github.com/TF0119/minesweeper/internal/game"
	"github.com/TF0119/minesweeper/internal/storage"
	tea "github.com/charmbracelet/bubbletea"
)

// Model is the bubbletea model.
type Model struct {
	board       *game.Board
	cursor      game.Coord
	screen      Screen
	screenStack []Screen // return targets for esc / back; see pushScreen
	difficulty  game.Difficulty
	config      storage.Config
	highscores  storage.HighScores
	stats       storage.Stats
	useColor    bool

	dailySeed game.Seed // today's challenge, for labelling the current board

	// boardNoGuess records what the current board was generated with. The
	// config setting only applies to the next board, so it cannot stand in
	// when this one has to be rebuilt.
	boardNoGuess bool

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

	moveLog     []game.Move
	replays     []game.Replay
	watchReplay *game.Replay
	watchStep   int
	watchBoard  *game.Board

	// Timelapse playback: auto-advances recorded moves at watchInterval.
	watchPlaying  bool
	watchPaused   bool
	watchInterval time.Duration

	// A timelapse borrows the viewport and cursor for a board that may be a
	// different size, so the live game's pair is parked here until it ends.
	playVp     viewport
	playCursor game.Coord
}

type tickMsg struct{}

type replayTickMsg struct{}

// timeNow is a variable so tests can pin the clock used for daily labelling.
var timeNow = time.Now

// NewModel creates a new UI model.
func NewModel(opts Options) Model {
	d := opts.Difficulty
	m := Model{
		board:        newBoard(d, opts.Seed, opts.Config.NoGuess),
		cursor:       game.Coord{X: d.Width / 2, Y: d.Height / 2},
		screen:       ScreenPlaying,
		difficulty:   d,
		boardNoGuess: opts.Config.NoGuess,
		config:       opts.Config,
		highscores:   opts.HighScores,
		stats:        opts.Stats,
		useColor:     !opts.NoColor,
		dailySeed:    game.DailySeed(timeNow()),
		vp:           fit(viewport{}, 0, 0, d.Width, d.Height),
		styles:       NewStyles(themeFromConfig(opts.Config), !opts.NoColor),
		glyphs:       newGlyphs(opts.Config.UseEmoji),
		keys:         DefaultKeyMap(),
	}
	if opts.Session != nil {
		m = m.restoreSession(*opts.Session)
	}
	m.playVp = m.vp
	m.playCursor = m.cursor
	return m
}

// restoreSession rebuilds the board the player left behind by replaying their
// moves onto a fresh one. A game that turns out to be over is dropped rather
// than shown: the file is a courtesy, not a source of truth worth trusting
// over the rules.
func (m Model) restoreSession(s storage.Session) Model {
	b := newBoard(s.Difficulty, s.Seed, s.NoGuess)
	game.Replay{Seed: s.Seed, Difficulty: s.Difficulty, Moves: s.Moves}.Apply(b, len(s.Moves))
	if b.Status() != game.StatusPlaying {
		return m
	}

	w, h := b.Width(), b.Height()
	m.board = b
	m.difficulty = s.Difficulty
	m.boardNoGuess = s.NoGuess
	m.moveLog = append([]game.Move(nil), s.Moves...)
	m.cursor = game.Coord{
		X: clamp(s.Cursor.X, 0, w-1),
		Y: clamp(s.Cursor.Y, 0, h-1),
	}
	m.vp = fit(viewport{}, 0, 0, w, h)

	// The clock counts time spent playing, so it picks up where it stopped
	// rather than from zero or from when the game was closed.
	m.elapsed = clamp(s.Seconds, 0, maxDisplaySeconds)
	if b.ElapsedReady() {
		m.timerActive = true
		m.timerStart = timeNow().Add(-time.Duration(m.elapsed) * time.Second)
	}
	return m
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

const (
	defaultReplayInterval = 300 * time.Millisecond
	minReplayInterval     = 75 * time.Millisecond
	maxReplayInterval     = 1500 * time.Millisecond
)

func replayTickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return replayTickMsg{}
	})
}
