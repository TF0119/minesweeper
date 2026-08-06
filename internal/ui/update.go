package ui

import (
	"time"

	"github.com/TF0119/minesweeper/internal/game"
	"github.com/TF0119/minesweeper/internal/storage"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.MouseMsg:
		return m.handleMouse(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.vp = fit(m.vp, m.width, m.height, m.board.Width(), m.board.Height())
		m.vp = m.vp.follow(m.cursor, m.board.Width(), m.board.Height())
		return m, nil
	case tickMsg:
		if m.timerActive && m.board.Status() == game.StatusPlaying {
			m.elapsed = m.computeElapsed()
			return m, tickCmd()
		}
		return m, nil
	case replayTickMsg:
		return m.handleReplayTick()
	case tea.QuitMsg:
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.screen {
	case ScreenMenu:
		return m.handleHubMenuKey(msg)
	case ScreenSettings:
		return m.handleSettingsKey(msg)
	case ScreenDifficultyMenu:
		return m.handleDifficultyMenuKey(msg)
	case ScreenReplays:
		return m.handleReplaysKey(msg)
	case ScreenReplayWatch:
		return m.handleReplayWatchKey(msg)
	case ScreenStats, ScreenHelp, ScreenGameOver, ScreenWin:
		return m.handleOverlayKey(msg)
	}
	return m.handlePlayingKey(msg)
}

// pushScreen opens a sub-screen and remembers where to return on esc.
func (m Model) pushScreen(screen Screen) Model {
	m.returnScreen = m.screen
	m.screen = screen
	m.menuIndex = 0
	if screen == ScreenReplays {
		m = m.loadReplays()
	}
	return m
}

// popScreen returns to the screen that opened the current one.
func (m Model) popScreen() Model {
	m.screen = m.returnScreen
	return m
}

func (m Model) openMenu() Model {
	m.returnScreen = ScreenPlaying
	m.screen = ScreenMenu
	m.menuIndex = 0
	return m
}

func (m Model) handleOverlayKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.New):
		return m.startNewGame(m.difficulty, game.RandomSeed())
	case key.Matches(msg, m.keys.Restart):
		return m.startNewGame(m.difficulty, m.board.Seed())
	case key.Matches(msg, m.keys.Quit):
		m.quitting = true
		return m, tea.Quit
	}
	if m.screen == ScreenGameOver || m.screen == ScreenWin {
		return m, nil
	}
	if msg.String() == "esc" {
		return m.popScreen(), nil
	}
	return m, nil
}

func (m Model) handlePlayingKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		_ = storage.SaveConfig(m.config)
		m.quitting = true
		return m, tea.Quit
	case key.Matches(msg, m.keys.Menu), msg.String() == "esc":
		return m.openMenu(), nil
	case key.Matches(msg, m.keys.Help):
		return m.pushScreen(ScreenHelp), nil
	case key.Matches(msg, m.keys.Stats):
		return m.pushScreen(ScreenStats), nil
	case key.Matches(msg, m.keys.Difficulty):
		return m.pushScreen(ScreenDifficultyMenu).withPresetIndex(), nil
	case key.Matches(msg, m.keys.New):
		return m.startNewGame(m.difficulty, game.RandomSeed())
	case key.Matches(msg, m.keys.Restart):
		return m.startNewGame(m.difficulty, m.board.Seed())
	case key.Matches(msg, m.keys.Up):
		return m.withCursor(0, -1), nil
	case key.Matches(msg, m.keys.Down):
		return m.withCursor(0, 1), nil
	case key.Matches(msg, m.keys.Left):
		return m.withCursor(-1, 0), nil
	case key.Matches(msg, m.keys.Right):
		return m.withCursor(1, 0), nil
	case key.Matches(msg, m.keys.Reveal):
		return m.applyReveal(m.cursor)
	case key.Matches(msg, m.keys.Flag):
		return m.applyFlag(m.cursor)
	case key.Matches(msg, m.keys.Chord):
		return m.applyChord(m.cursor)
	}
	return m, nil
}

func (m Model) withPresetIndex() Model {
	m.menuIndex = m.presetIndex()
	return m
}

func (m Model) handleHubMenuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, m.keys.Quit) {
		m.quitting = true
		return m, tea.Quit
	}
	switch msg.String() {
	case "up", "k":
		m.menuIndex = clamp(m.menuIndex-1, 0, len(hubMenuItems)-1)
	case "down", "j":
		m.menuIndex = clamp(m.menuIndex+1, 0, len(hubMenuItems)-1)
	case "esc":
		return m.popScreen(), nil
	case "enter", " ":
		return m.activateHubItem(hubMenuItems[m.menuIndex].action)
	}
	return m, nil
}

func (m Model) activateHubItem(action hubAction) (Model, tea.Cmd) {
	switch action {
	case hubResume:
		return m.popScreen(), nil
	case hubNewGame:
		return m.startNewGame(m.difficulty, game.RandomSeed())
	case hubDaily:
		return m.startDailyGame()
	case hubDifficulty:
		return m.pushScreen(ScreenDifficultyMenu).withPresetIndex(), nil
	case hubStatistics:
		return m.pushScreen(ScreenStats), nil
	case hubSettings:
		return m.pushScreen(ScreenSettings), nil
	case hubHelp:
		return m.pushScreen(ScreenHelp), nil
	case hubReplays:
		return m.pushScreen(ScreenReplays), nil
	case hubQuit:
		_ = storage.SaveConfig(m.config)
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) startDailyGame() (Model, tea.Cmd) {
	seed := game.DailySeed(timeNow())
	return m.startNewGame(m.difficulty, seed)
}

func (m Model) handleSettingsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, m.keys.Quit) {
		return m.popScreen(), nil
	}
	switch msg.String() {
	case "up", "k":
		m.menuIndex = clamp(m.menuIndex-1, 0, len(settingDefs)-1)
	case "down", "j":
		m.menuIndex = clamp(m.menuIndex+1, 0, len(settingDefs)-1)
	case "esc":
		_ = storage.SaveConfig(m.config)
		return m.popScreen(), nil
	case "enter", " ":
		m = m.cycleSetting(settingDefs[m.menuIndex].kind)
		_ = storage.SaveConfig(m.config)
	}
	return m, nil
}

func (m Model) handleDifficultyMenuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, m.keys.Quit) {
		return m.popScreen(), nil
	}
	switch msg.String() {
	case "up", "k":
		m.menuIndex = clamp(m.menuIndex-1, 0, len(menuPresets)-1)
	case "down", "j":
		m.menuIndex = clamp(m.menuIndex+1, 0, len(menuPresets)-1)
	case "enter", " ":
		preset := menuPresets[m.menuIndex]
		m.config.LastPreset = preset.String()
		_ = storage.SaveConfig(m.config)
		return m.startNewGame(game.PresetDifficulty(preset), game.RandomSeed())
	case "esc":
		return m.popScreen(), nil
	}
	return m, nil
}

// presetIndex locates the current difficulty in the menu, defaulting to the
// first entry for custom boards.
func (m Model) presetIndex() int {
	for i, p := range menuPresets {
		if p == m.difficulty.Preset {
			return i
		}
	}
	return 0
}

func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if m.screen != ScreenPlaying || !isMouseClick(msg) {
		return m, nil
	}

	click, lastBtn := classifyMouseEvent(msg, m.lastMouseBtn)
	m.lastMouseBtn = lastBtn
	if !click.flag && !click.reveal {
		return m, nil
	}

	c, ok := m.vp.toBoard(msg.X, msg.Y, m.board.Width(), m.board.Height())
	if !ok {
		return m, nil
	}
	m.cursor = c
	if click.flag {
		return m.applyFlag(c)
	}
	return m.applyReveal(c)
}

func (m Model) withCursor(dx, dy int) Model {
	w, h := m.board.Width(), m.board.Height()
	m.cursor = game.Coord{
		X: clamp(m.cursor.X+dx, 0, w-1),
		Y: clamp(m.cursor.Y+dy, 0, h-1),
	}
	m.vp = m.vp.follow(m.cursor, w, h)
	return m
}

func (m Model) startNewGame(d game.Difficulty, seed game.Seed) (Model, tea.Cmd) {
	m.board = newBoard(d, seed, m.config.NoGuess)
	m.difficulty = d
	m.cursor = game.Coord{X: d.Width / 2, Y: d.Height / 2}
	m.elapsed = 0
	m.timerActive = false
	m.timerStart = time.Time{}
	m.screen = ScreenPlaying
	m.returnScreen = ScreenPlaying
	m.errMsg = ""
	m.moveLog = nil
	m.watchBoard = nil
	m.watchReplay = nil
	m.watchPlaying = false
	m.watchPaused = false
	m.vp = fit(viewport{}, m.width, m.height, d.Width, d.Height)
	m.vp = m.vp.follow(m.cursor, d.Width, d.Height)
	return m, nil
}

func (m Model) computeElapsed() int {
	if m.timerStart.IsZero() {
		return 0
	}
	sec := int(time.Since(m.timerStart).Seconds())
	return clamp(sec, 0, maxDisplaySeconds)
}

func (m Model) stopTimer() Model {
	if m.timerActive {
		m.elapsed = m.computeElapsed()
	}
	m.timerActive = false
	return m
}

func (m Model) applyReveal(c game.Coord) (Model, tea.Cmd) {
	m = m.recordMove(game.MoveReveal, c)
	wasReady := m.board.ElapsedReady()
	return m.afterAction(m.board.Reveal(c), wasReady)
}

func (m Model) applyFlag(c game.Coord) (Model, tea.Cmd) {
	m = m.recordMove(game.MoveMark, c)
	m.board.CycleMark(c, m.config.QuestionMarks)
	return m, nil
}

func (m Model) applyChord(c game.Coord) (Model, tea.Cmd) {
	m = m.recordMove(game.MoveChord, c)
	wasReady := m.board.ElapsedReady()
	return m.afterAction(m.board.Chord(c), wasReady)
}

func (m Model) afterAction(res game.ActionResult, wasReady bool) (Model, tea.Cmd) {
	if !res.Ok {
		return m, nil
	}

	var cmd tea.Cmd
	if !wasReady && m.board.ElapsedReady() {
		m.timerStart = time.Now()
		m.timerActive = true
		m.elapsed = 0
		cmd = tickCmd()
	}

	switch res.Status {
	case game.StatusLost:
		m = m.stopTimer()
		m.saveReplay(false)
		m.stats.RecordLoss(m.difficulty.Key())
		_ = storage.SaveStats(m.stats)
		m.screen = ScreenGameOver
	case game.StatusWon:
		m = m.stopTimer()
		m.saveReplay(true)
		if m.highscores.TryUpdate(m.difficulty.Key(), m.elapsed, m.difficulty) {
			_ = storage.SaveHighScores(m.highscores)
		}
		m.stats.RecordWin(m.difficulty.Key(), m.elapsed)
		_ = storage.SaveStats(m.stats)
		m.screen = ScreenWin
	}
	return m, cmd
}
