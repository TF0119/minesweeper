package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/TF0119/minesweeper/internal/game"
	"github.com/TF0119/minesweeper/internal/storage"
	"github.com/TF0119/minesweeper/internal/storage/storagetest"
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

func TestReplayWatchUsesNoGuessGenerator(t *testing.T) {
	d := game.PresetDifficulty(game.Beginner)
	first := game.Coord{X: 4, Y: 4}
	seed := game.Seed(7)

	want := game.NewNoGuessBoard(d, seed)
	want.Reveal(first)
	classic := game.NewBoard(d, seed)
	classic.Reveal(first)

	same := true
	for y := 0; y < d.Height && same; y++ {
		for x := 0; x < d.Width && same; x++ {
			c := game.Coord{X: x, Y: y}
			if want.CellView(c) != classic.CellView(c) {
				same = false
			}
		}
	}
	if same {
		t.Skip("seed 7 yields the same opening for classic and no-guess; pick another later")
	}

	r := game.Replay{
		Seed:       seed,
		Difficulty: d,
		NoGuess:    true,
		Moves:      []game.Move{{Kind: game.MoveReveal, Coord: first}},
	}
	m, _ := testModel().startReplayWatch(r)
	m, _ = m.handleReplayTick()

	for y := 0; y < d.Height; y++ {
		for x := 0; x < d.Width; x++ {
			c := game.Coord{X: x, Y: y}
			if got, wantView := m.watchBoard.CellView(c), want.CellView(c); got != wantView {
				t.Fatalf("no-guess watch cell %+v = %+v, want %+v", c, got, wantView)
			}
		}
	}
}

func TestSaveReplayRecordsNoGuess(t *testing.T) {
	storagetest.IsolateConfigDir(t)
	cfg := storage.DefaultConfig()
	cfg.NoGuess = true
	m := NewModel(Options{
		Difficulty: game.PresetDifficulty(game.Beginner),
		Seed:       game.Seed(3),
		Config:     cfg,
		HighScores: storage.DefaultHighScores(),
		Stats:      storage.DefaultStats(),
	})
	m.boardNoGuess = true
	m.moveLog = []game.Move{{Kind: game.MoveReveal, Coord: game.Coord{X: 4, Y: 4}}}
	m.elapsed = 9
	m.saveReplay(true)

	list, err := storage.ListReplays(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || !list[0].NoGuess {
		t.Errorf("saved replay = %+v, want NoGuess true", list)
	}
}
