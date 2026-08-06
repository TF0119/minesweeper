package ui

import (
	"testing"

	"github.com/TF0119/minesweeper/internal/game"
	"github.com/TF0119/minesweeper/internal/storage"
	tea "github.com/charmbracelet/bubbletea"
)

func testModel() Model {
	return NewModel(Options{
		Difficulty: game.PresetDifficulty(game.Beginner),
		Seed:       game.Seed(1),
		Config:     storage.DefaultConfig(),
		HighScores: storage.DefaultHighScores(),
	})
}

func TestCursorStartsAtCenter(t *testing.T) {
	m := testModel()
	if m.cursor != (game.Coord{X: 4, Y: 4}) {
		t.Errorf("initial cursor = %+v, want (4,4)", m.cursor)
	}
}

func TestWithCursorMovesAndClamps(t *testing.T) {
	tests := []struct {
		name   string
		steps  [][2]int
		wantXY game.Coord
	}{
		{"right", [][2]int{{1, 0}}, game.Coord{X: 5, Y: 4}},
		{"down", [][2]int{{0, 1}}, game.Coord{X: 4, Y: 5}},
		{"clamped at top", [][2]int{{0, -100}}, game.Coord{X: 4, Y: 0}},
		{"clamped at right edge", [][2]int{{100, 0}}, game.Coord{X: 8, Y: 4}},
		{"diagonal via two steps", [][2]int{{-1, 0}, {0, -1}}, game.Coord{X: 3, Y: 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := testModel()
			for _, s := range tt.steps {
				m = m.withCursor(s[0], s[1])
			}
			if m.cursor != tt.wantXY {
				t.Errorf("cursor = %+v, want %+v", m.cursor, tt.wantXY)
			}
		})
	}
}

func TestNewGameUsesFreshSeedAndRestartReplaysIt(t *testing.T) {
	m := testModel()
	original := m.board.Seed()

	restarted, _ := m.startNewGame(m.difficulty, original)
	if restarted.board.Seed() != original {
		t.Errorf("restart seed = %v, want %v", restarted.board.Seed(), original)
	}

	fresh, _ := m.startNewGame(m.difficulty, original+1)
	if fresh.board.Seed() == original {
		t.Error("new game kept the old seed")
	}
}

func TestNewGameResetsTimer(t *testing.T) {
	m := testModel()
	m.timerActive = true
	m.elapsed = 42

	m, _ = m.startNewGame(m.difficulty, game.Seed(7))
	if m.timerActive || m.elapsed != 0 || !m.timerStart.IsZero() {
		t.Errorf("timer not reset: active=%v elapsed=%d start=%v",
			m.timerActive, m.elapsed, m.timerStart)
	}
}

func TestPresetIndexFindsCurrentDifficulty(t *testing.T) {
	m := testModel()
	m.difficulty = game.PresetDifficulty(game.Expert)
	if got := m.presetIndex(); menuPresets[got] != game.Expert {
		t.Errorf("presetIndex = %d, resolves to %v", got, menuPresets[got])
	}

	m.difficulty = game.Difficulty{Preset: game.Custom, Width: 5, Height: 5, Mines: 3}
	if got := m.presetIndex(); got != 0 {
		t.Errorf("custom difficulty should default to index 0, got %d", got)
	}
}

// SGR terminals deliver press and release for the same click. Acting on both
// would cycle the mark twice and flash a flag that immediately disappears.
func TestRightClickPressAndReleaseFlagsOnce(t *testing.T) {
	m := testModel()
	m.width, m.height = 120, 40
	m.vp = fit(m.vp, m.width, m.height, m.board.Width(), m.board.Height())
	c := game.Coord{X: 4, Y: 4}
	m.cursor = c

	press := tea.MouseMsg{X: 4*cellWidth + 1, Y: 1 + c.Y, Button: tea.MouseButtonRight, Action: tea.MouseActionPress}
	release := tea.MouseMsg{X: press.X, Y: press.Y, Button: tea.MouseButtonRight, Action: tea.MouseActionRelease}

	next, _ := m.handleMouse(press)
	m = next.(Model)
	if got := m.board.CellView(c).State; got != game.CellHidden {
		t.Fatalf("after press: state = %v, want hidden (wait for release)", got)
	}

	next, _ = m.handleMouse(release)
	m = next.(Model)
	if got := m.board.CellView(c).State; got != game.CellFlagged {
		t.Errorf("after release: state = %v, want flagged", got)
	}
}

func TestStartupGoesStraightToPlaying(t *testing.T) {
	m := testModel()
	if m.screen != ScreenPlaying {
		t.Errorf("screen = %v, want ScreenPlaying on launch", m.screen)
	}
}

func TestMenuOpensFromPlayAndResumes(t *testing.T) {
	m := testModel()
	m = m.openMenu()
	if m.screen != ScreenMenu {
		t.Fatalf("screen = %v, want ScreenMenu", m.screen)
	}
	m = m.popScreen()
	if m.screen != ScreenPlaying {
		t.Errorf("screen = %v, want ScreenPlaying after resume", m.screen)
	}
}

func TestSettingsCycleUpdatesTheme(t *testing.T) {
	m := testModel()
	m = m.pushScreen(ScreenSettings)
	before := m.config.Theme
	m = m.cycleSetting(settingTheme)
	if m.config.Theme == before {
		t.Error("theme should have cycled")
	}
}

func TestHubDailyStartsTodaysSeed(t *testing.T) {
	m := testModel()
	m, _ = m.activateHubItem(hubDaily)
	if !m.isDailyBoard() {
		t.Errorf("seed = %v, want today's daily %v", m.board.Seed(), m.dailySeed)
	}
}

func TestMenuOpensWithMKey(t *testing.T) {
	m := testModel()
	next, _ := m.handlePlayingKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	m = next.(Model)
	if m.screen != ScreenMenu {
		t.Errorf("screen = %v, want ScreenMenu", m.screen)
	}
}

func TestMenuOpensFromGameOver(t *testing.T) {
	m := testModel()
	m.screen = ScreenGameOver
	next, _ := m.handleOverlayKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	m = next.(Model)
	if m.screen != ScreenMenu {
		t.Fatalf("screen = %v, want ScreenMenu", m.screen)
	}
	if len(m.screenStack) != 1 || m.screenStack[0] != ScreenGameOver {
		t.Errorf("screenStack = %v, want [ScreenGameOver]", m.screenStack)
	}
	m = m.popScreen()
	if m.screen != ScreenGameOver {
		t.Errorf("screen = %v, want ScreenGameOver after resume", m.screen)
	}
}
