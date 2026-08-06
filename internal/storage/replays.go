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

const (
	replayVersion = 1
	// MaxReplays is how many finished games stay on disk and in the Watch list.
	MaxReplays = 20
)

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
// After writing, older files beyond MaxReplays are pruned.
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
	if err := writeJSONAtomic(path, payload); err != nil {
		return err
	}
	return PruneReplays(MaxReplays)
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

// ListReplays returns saved games, newest first. limit counts successfully
// parsed recordings; corrupt files are deleted so they cannot steal slots.
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

	out := make([]game.Replay, 0, len(files))
	for _, name := range files {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			_ = os.Remove(path)
			continue
		}
		var rf replayFile
		if err := json.Unmarshal(data, &rf); err != nil {
			_ = os.Remove(path)
			continue
		}
		if rf.Replay.ID == "" {
			rf.Replay.ID = strings.TrimSuffix(name, ".json")
		}
		out = append(out, rf.Replay)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out, nil
}

// DeleteReplay removes one saved game by ID. A missing file is not an error.
func DeleteReplay(id string) error {
	if id == "" || strings.Contains(id, "/") || strings.Contains(id, `\`) || strings.Contains(id, "..") {
		return fmt.Errorf("invalid replay id")
	}
	dir, err := ReplaysDir()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, id+".json")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// PruneReplays keeps the newest keep files by name and deletes the rest.
// keep <= 0 removes every recording.
func PruneReplays(keep int) error {
	dir, err := ReplaysDir()
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			files = append(files, e.Name())
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(files)))

	start := keep
	if keep < 0 {
		start = 0
	}
	if start > len(files) {
		return nil
	}
	for _, name := range files[start:] {
		_ = os.Remove(filepath.Join(dir, name))
	}
	return nil
}
