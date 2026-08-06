package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/TF0119/minesweeper/internal/game"
	"github.com/TF0119/minesweeper/internal/storage"
	"github.com/charmbracelet/lipgloss"
)

func TestSeedLabelDistinguishesDailyBoards(t *testing.T) {
	fixed := time.Date(2026, 8, 5, 9, 0, 0, 0, time.UTC)
	restore := timeNow
	timeNow = func() time.Time { return fixed }
	defer func() { timeNow = restore }()

	daily := NewModel(Options{
		Difficulty: game.PresetDifficulty(game.Beginner),
		Seed:       game.DailySeed(fixed),
		Config:     storage.DefaultConfig(),
		HighScores: storage.DefaultHighScores(),
	})
	if got := daily.seedLabel(); got != "daily 2026-08-05" {
		t.Errorf("daily label = %q", got)
	}

	ordinary := NewModel(Options{
		Difficulty: game.PresetDifficulty(game.Beginner),
		Seed:       game.Seed(42),
		Config:     storage.DefaultConfig(),
		HighScores: storage.DefaultHighScores(),
	})
	if got := ordinary.seedLabel(); got != "seed 42" {
		t.Errorf("seed label = %q", got)
	}
}

// The difficulty menu lists presets only, so a custom board has to say so or
// the highlight on the first row looks like the current difficulty.
func TestDifficultyMenuNamesTheCustomBoardInPlay(t *testing.T) {
	m := testModel()
	m.difficulty = game.Difficulty{Preset: game.Custom, Width: 20, Height: 10, Mines: 30}
	m = m.withPresetIndex()

	menu := m.renderDifficultyMenu()
	if !strings.Contains(menu, "custom 20x10, 30 mines") {
		t.Errorf("custom board not named in the menu:\n%s", menu)
	}

	m.difficulty = game.PresetDifficulty(game.Expert)
	if got := m.renderDifficultyMenu(); strings.Contains(got, "playing custom") {
		t.Errorf("preset board should not report a custom line:\n%s", got)
	}
}

func TestScrollIndicatorAppearsOnlyWhenClipped(t *testing.T) {
	m := NewModel(Options{
		Difficulty: game.PresetDifficulty(game.Expert),
		Seed:       game.Seed(1),
		Config:     storage.DefaultConfig(),
		HighScores: storage.DefaultHighScores(),
	})

	m.vp = fit(m.vp, 200, 40, m.board.Width(), m.board.Height())
	if got := m.renderScrollIndicator(); got != "" {
		t.Errorf("no indicator expected for a board that fits, got %q", got)
	}

	m.vp = fit(m.vp, 40, 12, m.board.Width(), m.board.Height())
	if got := m.renderScrollIndicator(); !strings.Contains(got, "/30 cols") {
		t.Errorf("indicator = %q, want the visible column range", got)
	}
}

func TestBoardRendersOnlyTheVisibleWindow(t *testing.T) {
	m := NewModel(Options{
		Difficulty: game.PresetDifficulty(game.Expert),
		Seed:       game.Seed(1),
		Config:     storage.DefaultConfig(),
		HighScores: storage.DefaultHighScores(),
	})
	m.vp = fit(m.vp, 40, 12, m.board.Width(), m.board.Height())

	lines := strings.Split(m.renderBoard(), "\n")
	if len(lines) != m.vp.rows {
		t.Errorf("rendered %d rows, want %d", len(lines), m.vp.rows)
	}
}

func TestCenterCell(t *testing.T) {
	tests := []struct{ in, want string }{
		{"1", " 1 "},
		{" ", "   "},
		{"⚑", " ⚑ "},
	}
	for _, tt := range tests {
		if got := centerCell(tt.in, cellWidth); got != tt.want {
			t.Errorf("centerCell(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// Colour is decoration, never the only carrier of state: a terminal that drops
// styling must still show which cells are unopened.
func TestBoardStaysReadableWithoutColor(t *testing.T) {
	m := NewModel(Options{
		Difficulty: game.PresetDifficulty(game.Beginner),
		Seed:       game.Seed(2),
		Config:     storage.DefaultConfig(),
		HighScores: storage.DefaultHighScores(),
		NoColor:    true,
	})
	m.screen = ScreenHelp // suppress the cursor so plain cells are compared

	opened := m.board.Reveal(game.Coord{X: 4, Y: 4})
	var empty game.Coord
	for _, c := range opened.Changed {
		if m.board.CellView(c).Adjacent == 0 {
			empty = c
			break
		}
	}

	var hidden game.Coord
	for y := 0; y < m.board.Height() && hidden == (game.Coord{}); y++ {
		for x := 0; x < m.board.Width(); x++ {
			c := game.Coord{X: x, Y: y}
			if m.board.CellView(c).State == game.CellHidden {
				hidden = c
				break
			}
		}
	}

	if got, want := m.renderCell(hidden), m.renderCell(empty); got == want {
		t.Errorf("hidden and revealed-empty cells both render as %q", got)
	}
}

func TestHelpListsEveryBinding(t *testing.T) {
	m := testModel()
	body := m.renderHelp()
	for _, b := range m.keys.bindings() {
		if !strings.Contains(body, b.Help().Desc) {
			t.Errorf("help is missing %q", b.Help().Desc)
		}
	}
}

func TestStatsScreenShowsEveryDifficulty(t *testing.T) {
	stats := storage.DefaultStats()
	stats.RecordWin(game.PresetDifficulty(game.Beginner).Key(), 42)
	stats.RecordLoss(game.PresetDifficulty(game.Expert).Key())

	scores := storage.DefaultHighScores()
	scores.TryUpdate(game.PresetDifficulty(game.Beginner).Key(), 42, game.PresetDifficulty(game.Beginner))

	m := NewModel(Options{
		Difficulty: game.PresetDifficulty(game.Beginner),
		Seed:       game.Seed(1),
		Config:     storage.DefaultConfig(),
		HighScores: scores,
		Stats:      stats,
	})

	body := m.renderStats()
	for _, p := range menuPresets {
		if !strings.Contains(body, p.String()) {
			t.Errorf("statistics screen is missing %q", p)
		}
	}
	if !strings.Contains(body, "best") {
		t.Error("statistics screen should have a best column")
	}
	if !strings.Contains(body, "42s") {
		t.Error("statistics screen should show the average/best time of the won game")
	}
	// Intermediate has never been played, so it has no average to show.
	if !strings.Contains(body, "—") {
		t.Error("an unplayed difficulty should show a placeholder, not a fake 0s")
	}
}

func TestStatusBarShrinksToFitTheTerminal(t *testing.T) {
	m := testModel()
	hints := m.statusHintsFor()

	widest := lipgloss.Width(hints[0])
	tests := []struct {
		name  string
		width int
	}{
		{"unknown size assumes room", 0},
		{"wide terminal", widest + 10},
		{"eighty columns", 80},
		{"very narrow", 20},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.width = tt.width
			want := hints[len(hints)-1]
			for _, candidate := range hints {
				if tt.width == 0 || lipgloss.Width(candidate) <= tt.width {
					want = candidate
					break
				}
			}
			if got := m.renderStatusBar(); !strings.Contains(got, strings.TrimSpace(want)) {
				t.Errorf("width %d rendered %q, want the hint %q", tt.width, got, want)
			}
		})
	}
}

// Every hint has to actually fit the width it is chosen for, or shrinking
// achieves nothing.
func TestNarrowestStatusHintFitsASmallTerminal(t *testing.T) {
	const narrowest = 40
	m := testModel()
	for _, screen := range []Screen{
		ScreenPlaying, ScreenMenu, ScreenDifficultyMenu, ScreenSettings,
		ScreenStats, ScreenHelp, ScreenReplays, ScreenReplayWatch,
		ScreenGameOver, ScreenWin,
	} {
		m.screen = screen
		hints := m.statusHintsFor()
		if got := lipgloss.Width(hints[len(hints)-1]); got > narrowest {
			t.Errorf("screen %v: shortest hint is %d columns, too wide for a %d-column terminal",
				screen, got, narrowest)
		}
	}
}

func TestStatusBarFollowsTheScreen(t *testing.T) {
	m := testModel()
	m.width = 120
	m.screen = ScreenMenu
	if got := m.renderStatusBar(); !strings.Contains(got, "enter select") {
		t.Errorf("menu status = %q, want enter select", got)
	}
	m.screen = ScreenReplayWatch
	if got := m.renderStatusBar(); !strings.Contains(got, "space pause") {
		t.Errorf("timelapse status = %q, want space pause", got)
	}
	m.screen = ScreenGameOver
	if got := m.renderStatusBar(); !strings.Contains(got, "r restart") {
		t.Errorf("game over status = %q, want r restart", got)
	}
}

func TestRetryHintUsesRestartWording(t *testing.T) {
	m := testModel()
	got := m.retryHint()
	if strings.Contains(got, "replay") {
		t.Errorf("retryHint = %q, must not say replay for the same-seed restart", got)
	}
	if !strings.Contains(got, "restart seed") || !strings.Contains(got, "m: menu") {
		t.Errorf("retryHint = %q, want restart seed and m: menu", got)
	}
}

func TestNoGuessLabelFollowsTheBoardNotTheConfig(t *testing.T) {
	m := testModel()
	m.boardNoGuess = true
	m.config.NoGuess = false
	if got := m.noGuessLabel(); got != "  no-guess" {
		t.Errorf("label = %q, want no-guess for a board generated that way", got)
	}

	m.boardNoGuess = false
	m.config.NoGuess = true
	if got := m.noGuessLabel(); got != "" {
		t.Errorf("label = %q, want empty when only the next-board setting is on", got)
	}
}

func TestNoGuessLabelDuringWatchUsesTheReplay(t *testing.T) {
	r := game.Replay{
		Seed:       game.Seed(1),
		Difficulty: game.PresetDifficulty(game.Beginner),
		NoGuess:    true,
		Moves:      []game.Move{{Kind: game.MoveReveal, Coord: game.Coord{X: 4, Y: 4}}},
	}
	m, _ := testModel().startReplayWatch(r)
	m.boardNoGuess = false
	m.config.NoGuess = false
	if got := m.noGuessLabel(); got != "  no-guess" {
		t.Errorf("watch label = %q, want no-guess from the replay", got)
	}
}

func TestHubResumeLabelDependsOnReturnTarget(t *testing.T) {
	m := testModel()
	m = m.pushScreen(ScreenMenu)
	if got := m.hubItemLabel(0); got != "Resume" {
		t.Errorf("from play: %q, want Resume", got)
	}

	m = testModel()
	m.screen = ScreenGameOver
	m = m.pushScreen(ScreenMenu)
	if got := m.hubItemLabel(0); got != "Back" {
		t.Errorf("from game over: %q, want Back", got)
	}
}

func TestWatchListShowsDifficultyAndDate(t *testing.T) {
	// The list formats in local time, so pin the zone: otherwise this test
	// reads a different date depending on where it runs.
	restore := time.Local
	time.Local = time.UTC
	defer func() { time.Local = restore }()

	m := testModel()
	m.replays = []game.Replay{{
		Seed:       game.Seed(7),
		Difficulty: game.PresetDifficulty(game.Beginner),
		NoGuess:    true,
		Won:        true,
		Seconds:    42,
		Moves:      []game.Move{{Kind: game.MoveReveal}},
		PlayedAt:   time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC),
	}}
	got := m.renderReplays()
	for _, want := range []string{"beginner", "won", "42s", "2026-08-06", "no-guess", "x delete"} {
		if !strings.Contains(got, want) {
			t.Errorf("watch list = %q, missing %q", got, want)
		}
	}
	if strings.Contains(got, "seed 7") || strings.Contains(got, " 7 ") {
		t.Errorf("watch list should not lead with the seed: %q", got)
	}
}

func TestWatchStatusBarMentionsDelete(t *testing.T) {
	m := testModel()
	m.screen = ScreenReplays
	m.replays = []game.Replay{{
		Seed:       game.Seed(1),
		Difficulty: game.PresetDifficulty(game.Beginner),
	}}
	m.width = 120
	if got := m.renderStatusBar(); !strings.Contains(got, "x delete") {
		t.Errorf("watch status = %q, want x delete", got)
	}
}

// With nothing recorded, enter and x are dead keys. The screen body already
// says so, and the status bar has to agree with it.
func TestEmptyWatchStatusBarOffersOnlyBack(t *testing.T) {
	m := testModel()
	m.screen = ScreenReplays
	m.width = 120
	got := m.renderStatusBar()
	for _, unwanted := range []string{"x delete", "timelapse"} {
		if strings.Contains(got, unwanted) {
			t.Errorf("empty watch status = %q, should not offer %q", got, unwanted)
		}
	}
	if !strings.Contains(got, "esc back") {
		t.Errorf("empty watch status = %q, want esc back", got)
	}
}
