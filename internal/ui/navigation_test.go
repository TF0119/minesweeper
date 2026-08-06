package ui

import (
	"testing"

	"github.com/TF0119/minesweeper/internal/game"
	tea "github.com/charmbracelet/bubbletea"
)

func sampleReplay() game.Replay {
	return game.Replay{
		Seed:       game.Seed(42),
		Difficulty: game.PresetDifficulty(game.Beginner),
		Moves:      []game.Move{{Kind: game.MoveReveal, Coord: game.Coord{X: 4, Y: 4}}},
	}
}

func openMenuWatchReplay(m Model) Model {
	m = m.openMenu()
	m, _ = m.activateHubItem(hubReplays)
	m.replays = []game.Replay{sampleReplay()}
	next, _ := m.startReplayWatch(m.replays[0])
	return next
}

func TestMenuWatchRoundTripReturnsToPlay(t *testing.T) {
	m := openMenuWatchReplay(testModel())

	next, _ := m.handleReplayWatchKey(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(Model)
	if m.screen != ScreenReplays {
		t.Fatalf("after watch esc: screen = %v, want ScreenReplays", m.screen)
	}

	next, _ = m.handleReplaysKey(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(Model)
	if m.screen != ScreenMenu {
		t.Fatalf("after list esc: screen = %v, want ScreenMenu", m.screen)
	}

	next, _ = m.handleHubMenuKey(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(Model)
	if m.screen != ScreenPlaying {
		t.Fatalf("after menu esc: screen = %v, want ScreenPlaying (stack=%v)", m.screen, m.screenStack)
	}
}

func TestMenuNotStuckAfterWatchResume(t *testing.T) {
	m := openMenuWatchReplay(testModel())

	next, _ := m.handleReplayWatchKey(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(Model)
	next, _ = m.handleReplaysKey(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(Model)

	next, _ = m.handleHubMenuKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = next.(Model)
	if m.screen != ScreenPlaying {
		t.Fatalf("resume after watch: screen = %v, want ScreenPlaying (stack=%v)", m.screen, m.screenStack)
	}
}

func TestReplayWatchQuitKeyExits(t *testing.T) {
	m := openMenuWatchReplay(testModel())
	next, _ := m.handleReplayWatchKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = next.(Model)
	if m.screen != ScreenReplays {
		t.Fatalf("q from watch: screen = %v, want ScreenReplays", m.screen)
	}
	if m.watchReplay != nil || m.watchBoard != nil {
		t.Error("watch state should be cleared after exit")
	}
}

func TestReplayWatchEscAfterFinished(t *testing.T) {
	m := openMenuWatchReplay(testModel())
	m.watchStep = len(m.watchReplay.Moves)
	m.watchPlaying = false

	next, _ := m.handleReplayWatchKey(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(Model)
	if m.screen != ScreenReplays {
		t.Fatalf("esc after finished watch: screen = %v, want ScreenReplays", m.screen)
	}
}

func TestMenuStatsWatchReturnsToMenu(t *testing.T) {
	m := testModel()
	m = m.openMenu()
	m, _ = m.activateHubItem(hubStatistics)
	m = m.popScreen() // esc from stats back to menu
	m, _ = m.activateHubItem(hubReplays)
	m.replays = []game.Replay{sampleReplay()}
	m, _ = m.startReplayWatch(m.replays[0])

	next, _ := m.handleReplayWatchKey(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(Model)
	next, _ = m.handleReplaysKey(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(Model)
	if m.screen != ScreenMenu {
		t.Fatalf("after watch from menu: screen = %v, want ScreenMenu", m.screen)
	}

	next, _ = m.handleHubMenuKey(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(Model)
	if m.screen != ScreenPlaying {
		t.Fatalf("after closing menu: screen = %v, want ScreenPlaying", m.screen)
	}
}

func TestMenuFromGameOverWatchAndBack(t *testing.T) {
	m := testModel()
	m.screen = ScreenGameOver
	m = m.openMenu()
	m, _ = m.activateHubItem(hubReplays)
	m.replays = []game.Replay{sampleReplay()}
	m, _ = m.startReplayWatch(m.replays[0])

	next, _ := m.handleReplayWatchKey(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(Model)
	next, _ = m.handleReplaysKey(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(Model)
	next, _ = m.handleHubMenuKey(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(Model)
	if m.screen != ScreenGameOver {
		t.Fatalf("screen = %v, want ScreenGameOver", m.screen)
	}
}

func TestHubSettingsAndBack(t *testing.T) {
	m := testModel()
	m = m.openMenu()
	m, _ = m.activateHubItem(hubSettings)
	if m.screen != ScreenSettings {
		t.Fatalf("screen = %v, want ScreenSettings", m.screen)
	}
	next, _ := m.handleSettingsKey(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(Model)
	if m.screen != ScreenMenu {
		t.Fatalf("screen = %v, want ScreenMenu", m.screen)
	}
	next, _ = m.handleHubMenuKey(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(Model)
	if m.screen != ScreenPlaying {
		t.Fatalf("screen = %v, want ScreenPlaying", m.screen)
	}
}

func TestNewGameClearsScreenStack(t *testing.T) {
	m := openMenuWatchReplay(testModel())
	m, _ = m.startNewGame(m.difficulty, game.Seed(99))
	if m.screen != ScreenPlaying {
		t.Fatalf("screen = %v, want ScreenPlaying", m.screen)
	}
	if len(m.screenStack) != 0 {
		t.Fatalf("screenStack = %v, want empty after new game", m.screenStack)
	}
}

func TestStatsOverlayQReturnsToPlay(t *testing.T) {
	m := testModel()
	m = m.pushScreen(ScreenStats)
	next, _ := m.handleOverlayKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = next.(Model)
	if m.quitting {
		t.Fatal("q on stats should return to play, not quit")
	}
	if m.screen != ScreenPlaying {
		t.Fatalf("screen = %v, want ScreenPlaying", m.screen)
	}
}

func TestStatsMenuResumeReturnsToPlay(t *testing.T) {
	m := testModel()
	m = m.pushScreen(ScreenStats)
	next, _ := m.handleOverlayKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	m = next.(Model)
	if m.screen != ScreenMenu {
		t.Fatalf("screen = %v, want ScreenMenu", m.screen)
	}
	next, _ = m.handleHubMenuKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = next.(Model)
	if m.screen != ScreenPlaying {
		t.Fatalf("resume after stats→menu: screen = %v, want ScreenPlaying", m.screen)
	}
}

func TestGameOverQStillQuits(t *testing.T) {
	m := testModel()
	m.screen = ScreenGameOver
	next, cmd := m.handleOverlayKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = next.(Model)
	if !m.quitting || cmd == nil {
		t.Fatal("q on game over should quit")
	}
}
