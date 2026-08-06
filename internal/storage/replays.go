package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/TF0119/minesweeper/internal/game"
)

const replayVersion = 1

type replayFile struct {
	Version int         `json:"version"`
	Replay  game.Replay `json:"replay"`
}

// ReplaysDir returns the directory where finished games are stored.
func ReplaysDir() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "replays"), nil
}

// SaveReplay writes a finished game to disk. The ID is derived from the time
// and seed so filenames stay unique and sortable without a central index.
func SaveReplay(r game.Replay) error {
	dir, err := ReplaysDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if r.ID == "" {
		r.ID = newReplayID(r)
	}
	r.PlayedAt = r.PlayedAt.UTC()
	payload := replayFile{Version: replayVersion, Replay: r}
	path := filepath.Join(dir, r.ID+".json")
	return writeJSONAtomic(path, payload)
}

func newReplayID(r game.Replay) string {
	when := r.PlayedAt
	if when.IsZero() {
		when = time.Now().UTC()
	}
	result := "lost"
	if r.Won {
		result = "won"
	}
	return fmt.Sprintf("%s-%s-seed%s",
		when.Format("20060102-150405"), result, r.Seed.String())
}

// ListReplays returns saved games, newest first.
func ListReplays(limit int) ([]game.Replay, error) {
	dir, err := ReplaysDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			files = append(files, e.Name())
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(files)))

	if limit > 0 && len(files) > limit {
		files = files[:limit]
	}

	out := make([]game.Replay, 0, len(files))
	for _, name := range files {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			continue
		}
		var rf replayFile
		if err := json.Unmarshal(data, &rf); err != nil {
			continue
		}
		if rf.Replay.ID == "" {
			rf.Replay.ID = strings.TrimSuffix(name, ".json")
		}
		out = append(out, rf.Replay)
	}
	return out, nil
}
