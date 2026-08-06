package ui

import (
	"testing"

	"github.com/TF0119/minesweeper/internal/game"
	"github.com/TF0119/minesweeper/internal/storage"
	"github.com/TF0119/minesweeper/internal/storage/storagetest"
	tea "github.com/charmbracelet/bubbletea"
)

// session drives the model the way bubbletea does, so a test can play a whole
// game through Update instead of calling handlers directly.
type session struct {
	t *testing.T
	m Model
}

func newSession(t *testing.T) *session {
	t.Helper()
	storagetest.IsolateConfigDir(t)
	s := &session{t: t, m: NewModel(Options{
		Difficulty: game.PresetDifficulty(game.Beginner),
		Seed:       game.Seed(1234),
		Config:     storage.DefaultConfig(),
		HighScores: storage.DefaultHighScores(),
		Stats:      storage.DefaultStats(),
	})}
	s.send(tea.WindowSizeMsg{Width: 100, Height: 40})
	return s
}

func (s *session) send(msg tea.Msg) {
	s.t.Helper()
	next, _ := s.m.Update(msg)
	s.m = next.(Model)
}

func (s *session) key(k string) {
	s.t.Helper()
	switch k {
	case "esc":
		s.send(tea.KeyMsg{Type: tea.KeyEsc})
	case "enter":
		s.send(tea.KeyMsg{Type: tea.KeyEnter})
	case "space":
		s.send(tea.KeyMsg{Type: tea.KeySpace})
	case "down":
		s.send(tea.KeyMsg{Type: tea.KeyDown})
	default:
		s.send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)})
	}
}

func (s *session) revealAt(c game.Coord) {
	s.m.cursor = c
	s.key("space")
}

func (s *session) selectHub(action hubAction) {
	s.t.Helper()
	s.key("m")
	for i := 0; i < len(hubMenuItems); i++ {
		if hubMenuItems[s.m.menuIndex].action == action {
			break
		}
		s.key("down")
	}
	s.key("enter")
}

// playToWin learns the layout by losing and restarting the same seed, which is
// deterministic because the seed and the opening move never change.
func (s *session) playToWin(first game.Coord) bool {
	s.t.Helper()
	mines := map[game.Coord]bool{}
	w, h := s.m.board.Width(), s.m.board.Height()

	for attempt := 0; attempt < 60; attempt++ {
		s.key("r")
		s.revealAt(first)

		for s.m.board.Status() == game.StatusPlaying {
			picked := false
			for y := 0; y < h && !picked; y++ {
				for x := 0; x < w && !picked; x++ {
					c := game.Coord{X: x, Y: y}
					if mines[c] || s.m.board.CellView(c).State != game.CellHidden {
						continue
					}
					s.revealAt(c)
					picked = true
					if s.m.board.Status() == game.StatusLost {
						mines[c] = true
					}
				}
			}
			if !picked {
				break
			}
		}
		if s.m.board.Status() == game.StatusWon {
			return true
		}
	}
	return false
}

func TestSessionPlayWinWatchAndResume(t *testing.T) {
	s := newSession(t)

	if s.m.screen != ScreenPlaying {
		t.Fatalf("launch screen = %v, want ScreenPlaying", s.m.screen)
	}

	if !s.playToWin(game.Coord{X: 4, Y: 4}) {
		t.Fatalf("never reached a win; status = %v", s.m.board.Status())
	}
	if s.m.screen != ScreenWin {
		t.Fatalf("screen = %v, want ScreenWin", s.m.screen)
	}

	s.selectHub(hubReplays)
	if s.m.screen != ScreenReplays {
		t.Fatalf("Watch opened %v, want ScreenReplays", s.m.screen)
	}
	if len(s.m.replays) == 0 {
		t.Fatal("the finished game was not recorded")
	}

	s.key("enter")
	if s.m.screen != ScreenReplayWatch {
		t.Fatalf("enter opened %v, want ScreenReplayWatch", s.m.screen)
	}
	for i := 0; i <= len(s.m.watchReplay.Moves)+2 && !s.m.replayFinished(); i++ {
		s.send(replayTickMsg{})
	}
	if !s.m.replayFinished() {
		t.Errorf("timelapse stalled at step %d of %d", s.m.watchStep, len(s.m.watchReplay.Moves))
	}

	s.key("esc")
	s.key("esc")
	s.key("esc")
	if s.m.screen != ScreenWin {
		t.Errorf("screen = %v, want ScreenWin where the menu was opened", s.m.screen)
	}
	if len(s.m.screenStack) != 0 {
		t.Errorf("screenStack = %v, want empty", s.m.screenStack)
	}

	s.key("n")
	if s.m.screen != ScreenPlaying || len(s.m.moveLog) != 0 {
		t.Errorf("new game: screen = %v, moves = %d", s.m.screen, len(s.m.moveLog))
	}
}

func TestSessionLossRecordsReplayAndMenuStillOpens(t *testing.T) {
	s := newSession(t)
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
	if s.m.screen != ScreenGameOver {
		t.Fatalf("screen = %v, want ScreenGameOver", s.m.screen)
	}

	s.selectHub(hubReplays)
	if s.m.screen != ScreenReplays || len(s.m.replays) == 0 {
		t.Fatalf("lost game not watchable: screen = %v, replays = %d", s.m.screen, len(s.m.replays))
	}

	s.key("esc")
	s.key("esc")
	if s.m.screen != ScreenGameOver {
		t.Errorf("screen = %v, want ScreenGameOver after closing the menu", s.m.screen)
	}
}

func TestSessionFlagCycleFollowsQuestionMarkSetting(t *testing.T) {
	s := newSession(t)
	c := game.Coord{X: 0, Y: 0}

	s.m.cursor = c
	got := []game.CellState{}
	for i := 0; i < 3; i++ {
		s.key("f")
		got = append(got, s.m.board.CellView(c).State)
	}
	want := []game.CellState{game.CellFlagged, game.CellQuestioned, game.CellHidden}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("questions on, step %d: state = %v, want %v", i, got[i], want[i])
		}
	}

	s.m.config.QuestionMarks = false
	got = got[:0]
	for i := 0; i < 3; i++ {
		s.key("f")
		got = append(got, s.m.board.CellView(c).State)
	}
	want = []game.CellState{game.CellFlagged, game.CellHidden, game.CellFlagged}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("questions off, step %d: state = %v, want %v", i, got[i], want[i])
		}
	}
}
