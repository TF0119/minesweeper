package game

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"strconv"
	"time"
)

// DailyKeyword selects the daily challenge seed on the command line.
const DailyKeyword = "daily"

// Seed identifies a reproducible board layout. It is deliberately 32-bit so
// that seeds stay short enough to share verbally or in a chat message.
type Seed uint32

// RandomSeed returns an unpredictable seed.
func RandomSeed() Seed {
	return Seed(rand.New(rand.NewSource(time.Now().UnixNano())).Uint32())
}

// DailySeed derives the seed for the daily challenge on t's UTC date. Every
// player who starts the challenge on the same date gets the same board.
func DailySeed(t time.Time) Seed {
	h := fnv.New32a()
	_, _ = h.Write([]byte(DailyDate(t)))
	return Seed(h.Sum32())
}

// DailyDate returns the UTC date string identifying a daily challenge.
func DailyDate(t time.Time) string {
	return t.UTC().Format("2006-01-02")
}

// ParseSeed accepts a decimal seed or the "daily" keyword. now supplies the
// clock for the daily challenge so callers can test it.
func ParseSeed(s string, now time.Time) (Seed, error) {
	if s == DailyKeyword {
		return DailySeed(now), nil
	}
	n, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("seed must be %q or a number in 0-%d", DailyKeyword, ^uint32(0))
	}
	return Seed(n), nil
}

// String formats the seed for display and sharing.
func (s Seed) String() string {
	return strconv.FormatUint(uint64(s), 10)
}

func (s Seed) rand() *rand.Rand {
	return rand.New(rand.NewSource(int64(s)))
}
