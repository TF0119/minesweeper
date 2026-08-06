package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/TF0119/minesweeper/internal/game"
	"github.com/TF0119/minesweeper/internal/storage"
	tea "github.com/charmbracelet/bubbletea"
)

// playedSession quits part way through a game and returns what was saved.
func playedSession(t *testing.T) (*session, storage.Session) {
	t.Helper()
	s := newSession(t)
	s.revealAt(game.Coord{X: 4, Y: 4})
	s.m.cursor = game.Coord{X: 7, Y: 2}
	s.key("f")
	s.m.elapsed = 12
	s.key("q")

	if !s.m.quitting {
		t.Fatal("q did not quit")
	}
	saved, ok, err := storage.LoadSession()
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("quitting mid-game saved no session")
	}
	return s, saved
}

func resume(t *testing.T, saved storage.Session) Model {
	t.Helper()
	return NewModel(Options{
		Difficulty: game.PresetDifficulty(game.Beginner),
		Seed:       game.RandomSeed(),
		Config:     storage.DefaultConfig(),
		HighScores: storage.DefaultHighScores(),
		Stats:      storage.DefaultStats(),
		Session:    &saved,
	})
}

func TestQuittingMidGameSavesAResumableSession(t *testing.T) {
	before, saved := playedSession(t)

	if saved.Seed != before.m.board.Seed() {
		t.Errorf("seed = %v, want %v", saved.Seed, before.m.board.Seed())
	}
	if len(saved.Moves) != len(before.m.moveLog) {
		t.Errorf("moves = %d, want %d", len(saved.Moves), len(before.m.moveLog))
	}
	if saved.Cursor != before.m.cursor {
		t.Errorf("cursor = %+v, want %+v", saved.Cursor, before.m.cursor)
	}
}

func TestResumeRebuildsTheBoardAndClock(t *testing.T) {
	before, saved := playedSession(t)
	m := resume(t, saved)

	if m.board.Seed() != before.m.board.Seed() {
		t.Errorf("resumed seed = %v, want %v", m.board.Seed(), before.m.board.Seed())
	}
	if m.screen != ScreenPlaying {
		t.Errorf("screen = %v, want ScreenPlaying", m.screen)
	}
	if m.cursor != before.m.cursor {
		t.Errorf("cursor = %+v, want %+v", m.cursor, before.m.cursor)
	}
	if m.elapsed != saved.Seconds {
		t.Errorf("elapsed = %d, want %d", m.elapsed, saved.Seconds)
	}
	if !m.timerActive {
		t.Error("a resumed game in progress should keep timing")
	}
	if m.Init() == nil {
		t.Error("a running clock needs the tick started on launch")
	}

	w, h := before.m.board.Width(), before.m.board.Height()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := game.Coord{X: x, Y: y}
			if got, want := m.board.CellView(c), before.m.board.CellView(c); got != want {
				t.Fatalf("cell %+v = %+v, want %+v", c, got, want)
			}
		}
	}
}

// Resuming has to keep playing, not just look right.
func TestResumedGameAcceptsMoves(t *testing.T) {
	_, saved := playedSession(t)
	m := resume(t, saved)

	moves := len(m.moveLog)
	var target game.Coord
	found := false
	for y := 0; y < m.board.Height() && !found; y++ {
		for x := 0; x < m.board.Width() && !found; x++ {
			c := game.Coord{X: x, Y: y}
			if m.board.CellView(c).State == game.CellHidden && m.board.CanReveal(c) {
				target, found = c, true
			}
		}
	}
	if !found {
		t.Skip("no hidden cell left to reveal")
	}

	m.cursor = target
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = next.(Model)
	if len(m.moveLog) != moves+1 {
		t.Errorf("move log = %d, want %d: the resumed game did not record the move",
			len(m.moveLog), moves+1)
	}
}

func TestFinishingAGameClearsTheSession(t *testing.T) {
	s := newSession(t)
	s.revealAt(game.Coord{X: 4, Y: 4})
	s.key("q")
	if _, ok, _ := storage.LoadSession(); !ok {
		t.Fatal("expected a session to clear")
	}

	// Lose on purpose: every hidden cell gets revealed until one is a mine.
	s2 := newSession(t)
	s2.revealAt(game.Coord{X: 4, Y: 4})
	w, h := s2.m.board.Width(), s2.m.board.Height()
	for y := 0; y < h && s2.m.board.Status() == game.StatusPlaying; y++ {
		for x := 0; x < w && s2.m.board.Status() == game.StatusPlaying; x++ {
			c := game.Coord{X: x, Y: y}
			if s2.m.board.CellView(c).State == game.CellHidden {
				s2.revealAt(c)
			}
		}
	}
	if s2.m.screen != ScreenGameOver {
		t.Fatalf("screen = %v, want ScreenGameOver", s2.m.screen)
	}
	if _, ok, _ := storage.LoadSession(); ok {
		t.Error("a finished game is still offered for resume")
	}
}

// Quitting before the first move must not resurrect the previous game.
func TestQuittingAFreshBoardClearsAnOlderSession(t *testing.T) {
	s := newSession(t)
	s.revealAt(game.Coord{X: 4, Y: 4})
	s.key("q")
	if _, ok, _ := storage.LoadSession(); !ok {
		t.Fatal("expected a saved session")
	}

	s.m.quitting = false
	s.key("n")
	s.key("q")
	if _, ok, _ := storage.LoadSession(); ok {
		t.Error("a board with no moves left the older session in place")
	}
}

func TestResumeIgnoresAFinishedSession(t *testing.T) {
	s := newSession(t)
	seed := s.m.board.Seed()
	s.revealAt(game.Coord{X: 4, Y: 4})
	w, h := s.m.board.Width(), s.m.board.Height()
	for y := 0; y < h && s.m.board.Status() == game.StatusPlaying; y++ {
		for x := 0; x < w && s.m.board.Status() == game.StatusPlaying; x++ {
			c := game.Coord{X: x, Y: y}
			if s.m.board.CellView(c).State == game.CellHidden {
				s.revealAt(c)
			}
		}
	}

	finished := storage.Session{
		Seed:       seed,
		Difficulty: s.m.difficulty,
		Moves:      append([]game.Move(nil), s.m.moveLog...),
		Seconds:    5,
		SavedAt:    time.Now(),
	}
	if err := storage.SaveSession(finished); err != nil {
		t.Fatal(err)
	}
	m := resume(t, finished)
	if m.board.Seed() == seed {
		t.Error("a finished session should be dropped, not restored")
	}
	if m.timerActive {
		t.Error("dropping a session should not start the clock")
	}
	if _, ok, _ := storage.LoadSession(); ok {
		t.Error("a finished session file should be cleared after it is ignored")
	}
}

func TestNewGameClearsASavedSession(t *testing.T) {
	s := newSession(t)
	s.revealAt(game.Coord{X: 4, Y: 4})
	s.key("q")
	if _, ok, _ := storage.LoadSession(); !ok {
		t.Fatal("expected a saved session")
	}

	s.m.quitting = false
	s.key("n")
	if _, ok, _ := storage.LoadSession(); ok {
		t.Error("starting a new board left the previous session on disk")
	}
}

func TestResumeShowsNoticeUntilAKeyIsPressed(t *testing.T) {
	_, saved := playedSession(t)
	m := resume(t, saved)
	if m.notice != "resumed" {
		t.Errorf("notice = %q, want resumed", m.notice)
	}
	if !strings.Contains(m.View(), "resumed") {
		t.Error("View should show the resumed notice")
	}

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = next.(Model)
	if m.notice != "" {
		t.Errorf("notice = %q after a key, want empty", m.notice)
	}
}

func TestCtrlCQuitsFromAnySubScreen(t *testing.T) {
	for _, action := range []hubAction{hubSettings, hubDifficulty, hubReplays} {
		s := newSession(t)
		s.selectHub(action)
		opened := s.m.screen

		next, cmd := s.m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
		m := next.(Model)
		if !m.quitting || cmd == nil {
			t.Errorf("ctrl+c on %v did not quit (screen is now %v)", opened, m.screen)
		}
	}
}
