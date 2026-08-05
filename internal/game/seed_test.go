package game

import (
	"testing"
	"time"
)

func TestSameSeedProducesSameBoard(t *testing.T) {
	d := PresetDifficulty(Intermediate)
	first := Coord{X: 3, Y: 3}

	layout := func(seed Seed) []bool {
		b := NewBoard(d, seed)
		b.Reveal(first)
		mines := make([]bool, d.Width*d.Height)
		for y := 0; y < d.Height; y++ {
			for x := 0; x < d.Width; x++ {
				c := Coord{X: x, Y: y}
				mines[c.index(d.Width)] = b.hasMine(c)
			}
		}
		return mines
	}

	a := layout(Seed(12345))
	b := layout(Seed(12345))
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("same seed produced different layout at index %d", i)
		}
	}

	c := layout(Seed(12346))
	same := true
	for i := range a {
		if a[i] != c[i] {
			same = false
			break
		}
	}
	if same {
		t.Error("different seeds produced identical layouts")
	}
}

func TestDailySeedIsStablePerDate(t *testing.T) {
	morning := time.Date(2026, 8, 5, 1, 0, 0, 0, time.UTC)
	evening := time.Date(2026, 8, 5, 23, 59, 0, 0, time.UTC)
	nextDay := time.Date(2026, 8, 6, 0, 0, 0, 0, time.UTC)

	if DailySeed(morning) != DailySeed(evening) {
		t.Error("daily seed changed within the same UTC date")
	}
	if DailySeed(morning) == DailySeed(nextDay) {
		t.Error("daily seed did not change across dates")
	}
}

func TestParseSeed(t *testing.T) {
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		in      string
		want    Seed
		wantErr bool
	}{
		{"0", 0, false},
		{"4294967295", Seed(^uint32(0)), false},
		{DailyKeyword, DailySeed(now), false},
		{"4294967296", 0, true},
		{"-1", 0, true},
		{"abc", 0, true},
		{"", 0, true},
	}
	for _, tt := range tests {
		got, err := ParseSeed(tt.in, now)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseSeed(%q) err = %v, wantErr %v", tt.in, err, tt.wantErr)
			continue
		}
		if err == nil && got != tt.want {
			t.Errorf("ParseSeed(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestSeedString(t *testing.T) {
	if Seed(42).String() != "42" {
		t.Errorf("Seed(42).String() = %q", Seed(42).String())
	}
}
