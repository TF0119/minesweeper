package ui

import (
	"testing"

	"github.com/TF0119/minesweeper/internal/game"
	"github.com/TF0119/minesweeper/internal/storage"
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
