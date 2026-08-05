package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/key"
	"github.com/takeru0119/minesweeper/internal/game"
	"github.com/takeru0119/minesweeper/internal/storage"
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
		return m, nil
	case tickMsg:
		if m.timerActive && m.board.Status() == game.StatusPlaying {
			m.elapsed++
		}
		return m, tickCmd()
	case tea.QuitMsg:
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.screen {
	case ScreenHelp, ScreenGameOver, ScreenWin:
		if key.Matches(msg, m.keys.New) {
			return m.startNewGame(m.difficulty)
		}
		if key.Matches(msg, m.keys.Quit) {
			m.quitting = true
			return m, tea.Quit
		}
		return m, nil
	case ScreenDifficultyMenu:
		return m.handleDifficultyMenuKey(msg)
	}

	if key.Matches(msg, m.keys.Quit) {
		_ = storage.SaveConfig(m.config)
		m.quitting = true
		return m, tea.Quit
	}
	if key.Matches(msg, m.keys.Help) {
		m.screen = ScreenHelp
		return m, nil
	}
	if key.Matches(msg, m.keys.Difficulty) {
		m.screen = ScreenDifficultyMenu
		m.menuIndex = 0
		return m, nil
	}
	if key.Matches(msg, m.keys.New) {
		return m.startNewGame(m.difficulty)
	}
	if key.Matches(msg, m.keys.Up) {
		m.moveCursor(0, -1)
	} else if key.Matches(msg, m.keys.Down) {
		m.moveCursor(0, 1)
	} else if key.Matches(msg, m.keys.Left) {
		m.moveCursor(-1, 0)
	} else if key.Matches(msg, m.keys.Right) {
		m.moveCursor(1, 0)
	} else if key.Matches(msg, m.keys.Reveal) {
		return m.applyReveal(m.cursor)
	} else if key.Matches(msg, m.keys.Flag) {
		return m.applyFlag(m.cursor)
	} else if key.Matches(msg, m.keys.Chord) {
		return m.applyChord(m.cursor)
	}
	return m, nil
}

func (m Model) handleDifficultyMenuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, m.keys.Quit) {
		m.screen = ScreenPlaying
		return m, nil
	}
	switch msg.String() {
	case "up", "k":
		if m.menuIndex > 0 {
			m.menuIndex--
		}
	case "down", "j":
		if m.menuIndex < 2 {
			m.menuIndex++
		}
	case "enter", " ":
		presets := []game.Preset{game.Beginner, game.Intermediate, game.Expert}
		d := game.PresetDifficulty(presets[m.menuIndex])
		m.config.LastPreset = presets[m.menuIndex].String()
		_ = storage.SaveConfig(m.config)
		m.screen = ScreenPlaying
		return m.startNewGame(d)
	case "esc":
		m.screen = ScreenPlaying
	}
	return m, nil
}

func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if m.screen != ScreenPlaying {
		return m, nil
	}
	if msg.Button != tea.MouseButtonLeft && msg.Button != tea.MouseButtonRight {
		return m, nil
	}
	x := msg.X / cellWidth
	y := msg.Y - 2 // offset for HUD row
	if x < 0 || y < 0 {
		return m, nil
	}
	c := game.Coord{X: x, Y: y}
	if !c.InBounds(m.board.Width(), m.board.Height()) {
		return m, nil
	}
	m.cursor = c
	if msg.Button == tea.MouseButtonLeft {
		return m.applyReveal(c)
	}
	return m.applyFlag(c)
}

func (m Model) moveCursor(dx, dy int) {
	nx := m.cursor.X + dx
	ny := m.cursor.Y + dy
	if nx < 0 {
		nx = 0
	}
	if ny < 0 {
		ny = 0
	}
	if nx >= m.board.Width() {
		nx = m.board.Width() - 1
	}
	if ny >= m.board.Height() {
		ny = m.board.Height() - 1
	}
	m.cursor = game.Coord{X: nx, Y: ny}
}

func (m Model) startNewGame(d game.Difficulty) (Model, tea.Cmd) {
	m.board = game.NewBoard(d, m.rng)
	m.difficulty = d
	m.cursor = game.Coord{X: d.Width / 2, Y: d.Height / 2}
	m.elapsed = 0
	m.timerActive = false
	m.screen = ScreenPlaying
	m.errMsg = ""
	return m, nil
}

func (m Model) applyReveal(c game.Coord) (Model, tea.Cmd) {
	if m.board.Status() != game.StatusPlaying {
		return m, nil
	}
	wasReady := m.board.ElapsedReady()
	res := m.board.Reveal(c)
	return m.afterAction(res, wasReady)
}

func (m Model) applyFlag(c game.Coord) (Model, tea.Cmd) {
	if m.board.Status() != game.StatusPlaying {
		return m, nil
	}
	res := m.board.ToggleFlag(c)
	if res.Ok {
		return m, nil
	}
	return m, nil
}

func (m Model) applyChord(c game.Coord) (Model, tea.Cmd) {
	if m.board.Status() != game.StatusPlaying {
		return m, nil
	}
	wasReady := m.board.ElapsedReady()
	res := m.board.Chord(c)
	return m.afterAction(res, wasReady)
}

func (m Model) afterAction(res game.ActionResult, wasReady bool) (Model, tea.Cmd) {
	if !res.Ok {
		return m, nil
	}
	var cmd tea.Cmd
	if !wasReady && m.board.ElapsedReady() && !m.timerActive {
		m.timerActive = true
		cmd = tickCmd()
	}
	switch res.Status {
	case game.StatusLost:
		m.timerActive = false
		m.screen = ScreenGameOver
	case game.StatusWon:
		m.timerActive = false
		if m.highscores.TryUpdate(m.difficulty.Key(), m.elapsed, m.difficulty) {
			_ = storage.SaveHighScores(m.highscores)
		}
		m.screen = ScreenWin
	}
	return m, cmd
}
