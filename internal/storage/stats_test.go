package storage

import "testing"

func TestStatsTracksStreaks(t *testing.T) {
	s := DefaultStats()
	const key = "beginner"

	s.RecordWin(key, 30)
	s.RecordWin(key, 50)
	s.RecordLoss(key)
	s.RecordWin(key, 40)

	got := s.For(key)
	if got.Played != 4 || got.Won != 3 {
		t.Errorf("played/won = %d/%d, want 4/3", got.Played, got.Won)
	}
	if got.CurrentStreak != 1 {
		t.Errorf("current streak = %d, want 1: the loss reset it", got.CurrentStreak)
	}
	if got.BestStreak != 2 {
		t.Errorf("best streak = %d, want 2", got.BestStreak)
	}
}

func TestTallyDerivedValues(t *testing.T) {
	t.Run("unplayed difficulty", func(t *testing.T) {
		var empty Tally
		if got := empty.WinRate(); got != 0 {
			t.Errorf("WinRate() = %v, want 0", got)
		}
		if got := empty.AverageWinSeconds(); got != -1 {
			t.Errorf("AverageWinSeconds() = %d, want -1 for no wins", got)
		}
	})

	t.Run("averages cover wins only", func(t *testing.T) {
		s := DefaultStats()
		s.RecordWin("expert", 100)
		s.RecordWin("expert", 200)
		s.RecordLoss("expert")

		got := s.For("expert")
		if avg := got.AverageWinSeconds(); avg != 150 {
			t.Errorf("AverageWinSeconds() = %d, want 150: the loss must not drag it down", avg)
		}
		if rate := got.WinRate(); rate < 0.66 || rate > 0.67 {
			t.Errorf("WinRate() = %v, want about 2/3", rate)
		}
	})
}

// Difficulties are tracked separately, so a run of easy wins must not flatter
// the expert record.
func TestStatsKeepDifficultiesApart(t *testing.T) {
	s := DefaultStats()
	s.RecordWin("beginner", 10)
	s.RecordLoss("expert")

	if got := s.For("beginner").Won; got != 1 {
		t.Errorf("beginner wins = %d, want 1", got)
	}
	if got := s.For("expert").Won; got != 0 {
		t.Errorf("expert wins = %d, want 0", got)
	}
	if got := s.For("intermediate").Played; got != 0 {
		t.Errorf("untouched difficulty played = %d, want 0", got)
	}
}
