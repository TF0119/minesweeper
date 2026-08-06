package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/TF0119/minesweeper/internal/game"
	"github.com/TF0119/minesweeper/internal/storage"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) activeBoard() *game.Board {
	if m.screen == ScreenReplayWatch && m.watchBoard != nil {
		return m.watchBoard
	}
	return m.board
}

func (m Model) recordMove(move game.Move) Model {
	if m.screen != ScreenPlaying {
		return m
	}
	m.moveLog = append(m.moveLog, move)
	return m
}

func (m Model) saveReplay(won bool) {
	if len(m.moveLog) == 0 {
		return
	}
	_ = storage.SaveReplay(game.Replay{
		Seed:       m.board.Seed(),
		Difficulty: m.difficulty,
		NoGuess:    m.boardNoGuess,
		Moves:      append([]game.Move(nil), m.moveLog...),
		Won:        won,
		Seconds:    m.elapsed,
		PlayedAt:   timeNow(),
	})
}

func (m Model) loadReplays() Model {
	list, err := storage.ListReplays(storage.MaxReplays)
	if err != nil {
		m.replays = nil
		return m
	}
	m.replays = list
	return m
}

func (m Model) renderReplays() string {
	if len(m.replays) == 0 {
		return m.renderOverlay("Watch", "No saved games yet.\n\nFinish a game to record a timelapse.\n\nesc back")
	}
	lines := make([]string, 0, len(m.replays)+2)
	for i, r := range m.replays {
		prefix := "  "
		if i == m.menuIndex {
			prefix = "> "
		}
		result := "lost"
		if r.Won {
			result = "won"
		}
		date := r.PlayedAt.UTC().Format("2006-01-02")
		line := fmt.Sprintf("%s%s  %s  %ds  %d moves  %s",
			prefix, m.difficultyLabelFor(r.Difficulty), result, r.Seconds, len(r.Moves), date)
		if r.NoGuess {
			line += "  no-guess"
		}
		lines = append(lines, line)
	}
	lines = append(lines, "", "enter timelapse · x delete · esc back")
	return m.renderOverlay("Watch", strings.Join(lines, "\n"))
}

func (m Model) startReplayWatch(r game.Replay) (Model, tea.Cmd) {
	if m.screen != ScreenReplayWatch {
		m.playVp = m.vp
		m.playCursor = m.cursor
	}
	copy := r
	m.watchReplay = &copy
	m.watchStep = 0
	m.watchBoard = newBoard(r.Difficulty, r.Seed, r.NoGuess)
	m.watchPlaying = true
	m.watchPaused = false
	m.watchInterval = defaultReplayInterval
	m = m.pushScreen(ScreenReplayWatch)
	m.cursor = game.Coord{X: r.Difficulty.Width / 2, Y: r.Difficulty.Height / 2}
	m.vp = fit(viewport{}, m.width, m.height, r.Difficulty.Width, r.Difficulty.Height)
	return m, replayTickCmd(m.watchInterval)
}

func (m Model) resetReplayWatch() Model {
	if m.watchReplay == nil {
		return m
	}
	m.watchStep = 0
	m.watchBoard = newBoard(m.watchReplay.Difficulty, m.watchReplay.Seed, m.watchReplay.NoGuess)
	m.watchPlaying = true
	m.watchPaused = false
	return m
}

func (m Model) advanceReplay() Model {
	if m.watchReplay == nil || m.watchBoard == nil {
		return m
	}
	if m.watchStep >= len(m.watchReplay.Moves) {
		return m
	}
	m.watchStep++
	m.watchReplay.Apply(m.watchBoard, m.watchStep)
	return m
}

func (m Model) replayFinished() bool {
	return m.watchReplay != nil && m.watchStep >= len(m.watchReplay.Moves)
}

func (m Model) handleReplayTick() (Model, tea.Cmd) {
	if m.screen != ScreenReplayWatch || !m.watchPlaying || m.watchPaused || m.replayFinished() {
		return m, nil
	}
	m = m.advanceReplay()
	if m.replayFinished() {
		m.watchPlaying = false
		return m, nil
	}
	return m, replayTickCmd(m.watchInterval)
}

func (m Model) replaySpeedLabel() string {
	if m.watchInterval <= 0 {
		return "?"
	}
	return fmt.Sprintf("%.1f/s", float64(time.Second)/float64(m.watchInterval))
}

func (m Model) renderReplayWatch() string {
	step, total := m.watchStep, 0
	if m.watchReplay != nil {
		total = len(m.watchReplay.Moves)
	}
	state := "playing"
	switch {
	case m.replayFinished():
		state = "finished"
	case m.watchPaused:
		state = "paused"
	}
	body := fmt.Sprintf("seed %s · %d/%d · %s · %s\n\nspace pause/resume · +/- speed · r restart · esc back",
		m.watchReplay.Seed, step, total, m.replaySpeedLabel(), state)
	return m.renderOverlay("Timelapse", body)
}

func (m Model) fasterReplay() (Model, tea.Cmd) {
	if m.watchInterval > minReplayInterval {
		m.watchInterval -= 75 * time.Millisecond
	}
	if m.watchInterval < minReplayInterval {
		m.watchInterval = minReplayInterval
	}
	return m.scheduleReplayTickIfPlaying()
}

func (m Model) slowerReplay() (Model, tea.Cmd) {
	if m.watchInterval < maxReplayInterval {
		m.watchInterval += 75 * time.Millisecond
	}
	if m.watchInterval > maxReplayInterval {
		m.watchInterval = maxReplayInterval
	}
	return m.scheduleReplayTickIfPlaying()
}

func (m Model) scheduleReplayTickIfPlaying() (Model, tea.Cmd) {
	if m.screen == ScreenReplayWatch && m.watchPlaying && !m.watchPaused && !m.replayFinished() {
		return m, replayTickCmd(m.watchInterval)
	}
	return m, nil
}

func (m Model) toggleReplayPause() (Model, tea.Cmd) {
	if m.replayFinished() {
		return m, nil
	}
	m.watchPaused = !m.watchPaused
	if m.watchPaused {
		m.watchPlaying = false
		return m, nil
	}
	m.watchPlaying = true
	return m, replayTickCmd(m.watchInterval)
}

// stopReplayWatch ends playback and hands the viewport and cursor back to the
// live board, which may be a different size than the replay just watched.
func (m Model) stopReplayWatch() Model {
	m.watchBoard = nil
	m.watchReplay = nil
	m.watchPlaying = false
	m.watchPaused = false

	w, h := m.board.Width(), m.board.Height()
	m.cursor = game.Coord{
		X: clamp(m.playCursor.X, 0, w-1),
		Y: clamp(m.playCursor.Y, 0, h-1),
	}
	m.vp = fit(m.playVp, m.width, m.height, w, h)
	m.vp = m.vp.follow(m.cursor, w, h)
	return m
}

func (m Model) handleReplaysKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, m.keys.Quit) {
		return m.popScreen(), nil
	}
	switch msg.String() {
	case "up", "k":
		if len(m.replays) > 0 {
			m.menuIndex = clamp(m.menuIndex-1, 0, len(m.replays)-1)
		}
	case "down", "j":
		if len(m.replays) > 0 {
			m.menuIndex = clamp(m.menuIndex+1, 0, len(m.replays)-1)
		}
	case "esc":
		return m.popScreen(), nil
	case "x":
		if len(m.replays) == 0 {
			return m, nil
		}
		id := m.replays[m.menuIndex].ID
		_ = storage.DeleteReplay(id)
		m = m.loadReplays()
		if len(m.replays) == 0 {
			m.menuIndex = 0
		} else {
			m.menuIndex = clamp(m.menuIndex, 0, len(m.replays)-1)
		}
		m.notice = "deleted"
		return m, nil
	case "enter", " ":
		if len(m.replays) == 0 {
			return m, nil
		}
		return m.startReplayWatch(m.replays[m.menuIndex])
	}
	return m, nil
}

func (m Model) exitReplayWatch() (Model, tea.Cmd) {
	m = m.stopReplayWatch()
	return m.popScreen(), nil
}

func (m Model) handleReplayWatchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, m.keys.Quit) {
		return m.exitReplayWatch()
	}
	switch msg.String() {
	case "esc":
		return m.exitReplayWatch()
	case " ":
		if m.replayFinished() {
			return m, nil
		}
		return m.toggleReplayPause()
	case "+", "=":
		return m.fasterReplay()
	case "-", "_":
		return m.slowerReplay()
	case "r":
		m = m.resetReplayWatch()
		return m, replayTickCmd(m.watchInterval)
	case "enter":
		if m.replayFinished() {
			m = m.resetReplayWatch()
			return m, replayTickCmd(m.watchInterval)
		}
		return m.toggleReplayPause()
	}
	return m, nil
}
