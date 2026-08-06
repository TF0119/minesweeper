package ui

import (
	"testing"

	"github.com/TF0119/minesweeper/internal/game"
	tea "github.com/charmbracelet/bubbletea"
)

func expertReplay() game.Replay {
	return game.Replay{
		Seed:       game.Seed(777),
		Difficulty: game.PresetDifficulty(game.Expert),
		Moves: []game.Move{
			{Kind: game.MoveReveal, Coord: game.Coord{X: 15, Y: 8}},
			{Kind: game.MoveReveal, Coord: game.Coord{X: 20, Y: 10}},
		},
		Seconds: 33,
	}
}

// Watching a replay of another difficulty borrows the viewport and cursor for a
// board of a different size. Both have to come back, or the live board renders
// at the replay's size with the cursor stranded outside it.
func TestWatchRestoresPlayViewportAndCursor(t *testing.T) {
	m := testModel()
	next, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	m = next.(Model)

	wantVp, wantCursor := m.vp, m.cursor

	m = m.pushScreen(ScreenReplays)
	m.replays = []game.Replay{expertReplay()}
	m, _ = m.startReplayWatch(m.replays[0])
	if m.vp.cols == wantVp.cols && m.vp.rows == wantVp.rows {
		t.Fatal("timelapse should size the viewport to the replay board")
	}

	watchNext, _ := m.handleReplayWatchKey(tea.KeyMsg{Type: tea.KeyEsc})
	m = watchNext.(Model)
	m = m.popScreen()

	if m.vp != wantVp {
		t.Errorf("viewport = %+v, want %+v restored after watching", m.vp, wantVp)
	}
	if m.cursor != wantCursor {
		t.Errorf("cursor = %+v, want %+v restored after watching", m.cursor, wantCursor)
	}
	if !m.cursor.InBounds(m.board.Width(), m.board.Height()) {
		t.Errorf("cursor %+v outside the %dx%d live board",
			m.cursor, m.board.Width(), m.board.Height())
	}

	// The restored cursor must still be usable.
	revealed, _ := m.applyReveal(m.cursor)
	if revealed.board.CellView(revealed.cursor).State != game.CellRevealed {
		t.Error("reveal at the restored cursor did nothing")
	}
}

// A resize while a timelapse is on screen must clip the replay board, not the
// live one, or a large replay is drawn at the small board's size.
func TestResizeDuringWatchFollowsReplayBoard(t *testing.T) {
	m := testModel()
	next, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	m = next.(Model)

	m = m.pushScreen(ScreenReplays)
	m.replays = []game.Replay{expertReplay()}
	m, _ = m.startReplayWatch(m.replays[0])

	next, _ = m.Update(tea.WindowSizeMsg{Width: 40, Height: 15})
	m = next.(Model)

	if m.vp.cols > m.watchBoard.Width() || m.vp.rows > m.watchBoard.Height() {
		t.Errorf("viewport %+v exceeds the replay board", m.vp)
	}
	if m.vp.cols <= m.board.Width() && m.vp.rows <= m.board.Height() {
		t.Errorf("viewport %+v was fitted to the live %dx%d board instead of the replay board",
			m.vp, m.board.Width(), m.board.Height())
	}
	if m.renderScrollIndicator() == "" {
		t.Error("a clipped replay board should report which slice is on screen")
	}
}

// A recorded game should contain only moves that changed the board; replaying
// no-ops spends timelapse frames showing nothing happen.
func TestNoOpActionsAreNotRecorded(t *testing.T) {
	m := testModel()
	m, _ = m.applyReveal(game.Coord{X: 4, Y: 4})
	after := len(m.moveLog)

	m, _ = m.applyReveal(game.Coord{X: 4, Y: 4}) // already revealed
	m, _ = m.applyChord(game.Coord{X: 0, Y: 0})  // hidden cell, nothing to chord
	m, _ = m.applyFlag(game.Coord{X: 4, Y: 4})   // revealed cells cannot be marked

	if len(m.moveLog) != after {
		t.Errorf("move log grew from %d to %d on no-op actions", after, len(m.moveLog))
	}
}
