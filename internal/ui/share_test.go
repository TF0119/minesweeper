package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/TF0119/minesweeper/internal/game"
	"github.com/TF0119/minesweeper/internal/storage"
)

func TestShareCardDailyWin(t *testing.T) {
	fixed := time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)
	prev := timeNow
	timeNow = func() time.Time { return fixed }
	defer func() { timeNow = prev }()

	m := NewModel(Options{
		Difficulty: game.PresetDifficulty(game.Expert),
		Seed:       game.DailySeed(fixed),
		Config:     storage.Config{NoGuess: true},
		HighScores: storage.DefaultHighScores(),
	})
	m.elapsed = 142
	m.boardNoGuess = true
	m.shareCard = m.buildShareCard()

	got := m.shareCard
	wantLines := []string{
		"minesweeper · daily 2026-08-06",
		"expert · no-guess · 142s",
		repoURL,
	}
	for _, want := range wantLines {
		if !strings.Contains(got, want) {
			t.Errorf("share card %q missing %q", got, want)
		}
	}
	if lines := strings.Split(strings.TrimSpace(got), "\n"); len(lines) != 3 {
		t.Errorf("share card has %d lines, want 3:\n%s", len(lines), got)
	}
}

func TestShareCardSeedWin(t *testing.T) {
	m := NewModel(Options{
		Difficulty: game.PresetDifficulty(game.Beginner),
		Seed:       game.Seed(1487233901),
		Config:     storage.DefaultConfig(),
		HighScores: storage.DefaultHighScores(),
	})
	m.elapsed = 30
	m.boardNoGuess = false
	got := m.buildShareCard()
	if !strings.Contains(got, "minesweeper · seed 1487233901") {
		t.Errorf("share card = %q, want seed line", got)
	}
	if strings.Contains(got, "no-guess") {
		t.Errorf("share card = %q, must not claim no-guess", got)
	}
	if !strings.Contains(got, "beginner · 30s") {
		t.Errorf("share card = %q, want difficulty and time", got)
	}
}

func TestShareCardCustomBoard(t *testing.T) {
	m := NewModel(Options{
		Difficulty: game.Difficulty{Preset: game.Custom, Width: 20, Height: 10, Mines: 30},
		Seed:       game.Seed(7),
		Config:     storage.DefaultConfig(),
		HighScores: storage.DefaultHighScores(),
	})
	m.elapsed = 55
	got := m.buildShareCard()
	if !strings.Contains(got, "custom 20x10/30") {
		t.Errorf("share card = %q, want custom size label", got)
	}
}

func TestShareCardOnlyPrintedAfterQuitFromWin(t *testing.T) {
	m := testModel()
	m.shareCard = "card"
	if got := m.ShareCard(); got != "" {
		t.Errorf("ShareCard before quit = %q, want empty", got)
	}
	m.quitting = true
	if got := m.ShareCard(); got != "card" {
		t.Errorf("ShareCard after quit = %q, want card", got)
	}
}

func TestNewGameClearsShareCard(t *testing.T) {
	m := testModel()
	m.shareCard = "stale"
	m, _ = m.startNewGame(m.difficulty, game.Seed(2))
	if m.shareCard != "" {
		t.Errorf("shareCard = %q after new game, want empty", m.shareCard)
	}
}

func TestRenderWinIncludesShareCard(t *testing.T) {
	m := testModel()
	m.screen = ScreenWin
	m.elapsed = 12
	m.shareCard = m.buildShareCard()
	got := m.renderWin()
	if !strings.Contains(got, repoURL) {
		t.Errorf("victory overlay missing repo URL:\n%s", got)
	}
	if !strings.Contains(got, "Cleared in 12 seconds.") {
		t.Errorf("victory overlay missing clear time:\n%s", got)
	}
}
