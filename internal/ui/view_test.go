package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/TF0119/minesweeper/internal/game"
	"github.com/TF0119/minesweeper/internal/storage"
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

func TestHelpListsEveryBinding(t *testing.T) {
	m := testModel()
	body := m.renderHelp()
	for _, b := range m.keys.bindings() {
		if !strings.Contains(body, b.Help().Desc) {
			t.Errorf("help is missing %q", b.Help().Desc)
		}
	}
}
