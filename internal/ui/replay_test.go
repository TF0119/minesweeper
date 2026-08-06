package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/TF0119/minesweeper/internal/game"
)

func TestReplayTimelapseAdvancesOnTick(t *testing.T) {
	r := game.Replay{
		Seed:       game.Seed(1),
		Difficulty: game.PresetDifficulty(game.Beginner),
		Moves: []game.Move{
			{Kind: game.MoveReveal, Coord: game.Coord{X: 4, Y: 4}},
			{Kind: game.MoveMark, Coord: game.Coord{X: 1, Y: 1}},
		},
	}
	m, _ := testModel().startReplayWatch(r)
	if m.watchStep != 0 {
		t.Fatalf("watchStep = %d, want 0 before first tick", m.watchStep)
	}

	next, cmd := m.handleReplayTick()
	m = next
	if m.watchStep != 1 {
		t.Errorf("watchStep = %d, want 1 after tick", m.watchStep)
	}
	if cmd == nil {
		t.Error("expected another tick while moves remain")
	}
}

func TestReplayPauseStopsTicks(t *testing.T) {
	r := game.Replay{
		Seed:       game.Seed(1),
		Difficulty: game.PresetDifficulty(game.Beginner),
		Moves:      []game.Move{{Kind: game.MoveReveal, Coord: game.Coord{X: 4, Y: 4}}},
	}
	m, _ := testModel().startReplayWatch(r)
	m.watchPaused = true
	m.watchPlaying = false

	next, cmd := m.handleReplayTick()
	m = next
	if m.watchStep != 0 {
		t.Errorf("paused timelapse advanced to step %d", m.watchStep)
	}
	if cmd != nil {
		t.Error("paused timelapse should not schedule ticks")
	}
}

func TestReplaySpeedAdjustsInterval(t *testing.T) {
	m := testModel()
	m.watchInterval = defaultReplayInterval
	before := m.watchInterval
	m, _ = m.fasterReplay()
	if m.watchInterval >= before {
		t.Errorf("faster: interval %v did not decrease from %v", m.watchInterval, before)
	}
	m, _ = m.slowerReplay()
	if m.watchInterval <= before-75*time.Millisecond {
		t.Error("slower should increase interval again")
	}
}

func TestReplaySpeedReschedulesTickWhilePlaying(t *testing.T) {
	r := game.Replay{
		Seed:       game.Seed(1),
		Difficulty: game.PresetDifficulty(game.Beginner),
		Moves: []game.Move{
			{Kind: game.MoveReveal, Coord: game.Coord{X: 4, Y: 4}},
			{Kind: game.MoveReveal, Coord: game.Coord{X: 3, Y: 4}},
		},
	}
	m, _ := testModel().startReplayWatch(r)
	_, cmd := m.fasterReplay()
	if cmd == nil {
		t.Error("expected tick reschedule after speed change during playback")
	}
}

func TestReplayFinishedStopsPlayback(t *testing.T) {
	r := game.Replay{
		Seed:       game.Seed(1),
		Difficulty: game.PresetDifficulty(game.Beginner),
		Moves:      []game.Move{{Kind: game.MoveReveal, Coord: game.Coord{X: 4, Y: 4}}},
	}
	m, _ := testModel().startReplayWatch(r)
	m, _ = m.handleReplayTick() // apply sole move
	if !m.replayFinished() {
		t.Fatal("expected replay to be finished")
	}
	next, cmd := m.handleReplayTick()
	m = next
	if cmd != nil || m.watchPlaying {
		t.Error("finished timelapse should stop scheduling ticks")
	}
}

func TestReplayWatchHUDUsesReplayBoard(t *testing.T) {
	r := game.Replay{
		Seed:       game.Seed(99),
		Difficulty: game.PresetDifficulty(game.Beginner),
		Moves:      []game.Move{{Kind: game.MoveReveal, Coord: game.Coord{X: 4, Y: 4}}},
		Seconds:    17,
	}
	m, _ := testModel().startReplayWatch(r)
	flag := game.MarkFlag
	m.watchReplay.Moves = append(m.watchReplay.Moves, game.Move{
		Kind: game.MoveMark, Coord: game.Coord{X: 1, Y: 1}, TargetMark: &flag,
	})
	m.watchReplay.Apply(m.watchBoard, 2)

	hud := m.renderHUD()
	if !strings.Contains(hud, "seed 99") {
		t.Errorf("HUD should show replay seed, got %q", hud)
	}
	if !strings.Contains(hud, "017") {
		t.Errorf("HUD should show replay time, got %q", hud)
	}
	liveMines := m.board.RemainingMines()
	watchMines := m.watchBoard.RemainingMines()
	if liveMines != watchMines && strings.Contains(hud, fmt.Sprintf("%03d", liveMines)) {
		t.Errorf("HUD mine count should follow watch board (%d), not live board (%d)", watchMines, liveMines)
	}
}
