package ui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/TF0119/minesweeper/internal/game"
	"github.com/TF0119/minesweeper/internal/storage"
	"github.com/charmbracelet/lipgloss"
)

const (
	cellWidth = 3
	// maxDisplaySeconds matches the classic three-digit timer.
	maxDisplaySeconds = 999
)

// menuPresets is the difficulty menu, shared by rendering and key handling.
var menuPresets = []game.Preset{game.Beginner, game.Intermediate, game.Expert}

// Screen identifies the active UI screen.
type Screen int

const (
	ScreenPlaying Screen = iota
	ScreenDifficultyMenu
	ScreenStats
	ScreenHelp
	ScreenGameOver
	ScreenWin
)

// View renders the TUI.
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	parts := []string{m.renderHUD(), m.renderBoard()}

	if s := m.renderScrollIndicator(); s != "" {
		parts = append(parts, s)
	}
	if m.errMsg != "" {
		parts = append(parts, m.styles.Warning.Render(m.errMsg))
	}

	switch m.screen {
	case ScreenDifficultyMenu:
		parts = append(parts, m.renderDifficultyMenu())
	case ScreenStats:
		parts = append(parts, m.renderStats())
	case ScreenHelp:
		parts = append(parts, m.renderHelp())
	case ScreenGameOver:
		parts = append(parts, m.renderGameOver())
	case ScreenWin:
		parts = append(parts, m.renderWin())
	}

	parts = append(parts, m.renderStatusBar())
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m Model) renderCell(c game.Coord) string {
	view := m.board.CellView(c)
	isCursor := c == m.cursor && m.screen == ScreenPlaying

	var content string
	switch view.State {
	case game.CellHidden:
		content = m.glyphs.hidden
	case game.CellFlagged:
		content = m.glyphs.flag
	case game.CellQuestioned:
		content = m.glyphs.question
	case game.CellRevealed:
		switch {
		case view.ShowMine:
			content = m.glyphs.mine
		case view.Adjacent == 0:
			content = " "
		default:
			content = fmt.Sprintf("%d", view.Adjacent)
		}
	}

	text := centerCell(content, cellWidth)
	switch {
	case isCursor:
		return m.styles.Cursor.Render(text)
	case view.State == game.CellFlagged:
		return m.styles.Flagged.Render(text)
	case view.State == game.CellQuestioned:
		return m.styles.Questioned.Render(text)
	case view.State == game.CellRevealed && view.ShowMine:
		return m.styles.Mine.Render(text)
	case view.State == game.CellRevealed && view.Adjacent > 0:
		return m.styles.Digits[view.Adjacent].Render(text)
	case view.State == game.CellRevealed:
		return m.styles.Revealed.Render(text)
	default:
		return m.styles.Hidden.Render(text)
	}
}

// centerCell pads s to width columns. Glyphs are single-width, so rune count
// is a valid column count here.
func centerCell(s string, width int) string {
	n := utf8.RuneCountInString(s)
	if n >= width {
		return s
	}
	pad := width - n
	left := pad / 2
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", pad-left)
}

func (m Model) renderBoard() string {
	rows := make([]string, 0, m.vp.rows)
	for row := 0; row < m.vp.rows; row++ {
		cells := make([]string, 0, m.vp.cols)
		for col := 0; col < m.vp.cols; col++ {
			cells = append(cells, m.renderCell(game.Coord{
				X: m.vp.offsetX + col,
				Y: m.vp.offsetY + row,
			}))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cells...))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m Model) renderHUD() string {
	best := "---"
	if b := m.highscores.Best(m.difficulty.Key()); b >= 0 {
		best = fmt.Sprintf("%03d", b)
	}
	line := fmt.Sprintf(
		" Mines:%03d  Time:%03d  Best:%s  [%s]  %s%s ",
		m.board.RemainingMines(), m.elapsed, best,
		m.difficultyLabel(), m.seedLabel(), m.noGuessLabel(),
	)
	return m.styles.HUD.Render(line)
}

func (m Model) difficultyLabel() string {
	if m.difficulty.Preset == game.Custom {
		return fmt.Sprintf("custom %dx%d", m.difficulty.Width, m.difficulty.Height)
	}
	return m.difficulty.Preset.String()
}

func (m Model) seedLabel() string {
	if m.isDailyBoard() {
		return "daily " + game.DailyDate(timeNow())
	}
	return "seed " + m.board.Seed().String()
}

// noGuessLabel reports on the no-guess promise. The search can come up empty,
// and saying so is better than letting the player trust a guarantee that does
// not hold for this board.
func (m Model) noGuessLabel() string {
	if !m.config.NoGuess {
		return ""
	}
	if m.board.ElapsedReady() && !m.board.NoGuess() {
		return "  guess needed"
	}
	return "  no-guess"
}

// renderScrollIndicator shows which slice of a clipped board is on screen.
func (m Model) renderScrollIndicator() string {
	w, h := m.board.Width(), m.board.Height()
	if !m.vp.scrolls(w, h) {
		return ""
	}
	return m.styles.StatusBar.Render(fmt.Sprintf(
		" view %d-%d/%d cols · %d-%d/%d rows ",
		m.vp.offsetX+1, m.vp.offsetX+m.vp.cols, w,
		m.vp.offsetY+1, m.vp.offsetY+m.vp.rows, h,
	))
}

// statusHints are the status bar texts from most to least detailed. Listing
// every key stopped fitting an 80-column terminal, and a wrapped status bar
// pushes the board off screen.
var statusHints = []string{
	" arrows/hjkl move · space reveal · f mark · c chord · n new · r restart · d difficulty · s stats · ? help · q quit ",
	" move · space reveal · f mark · c chord · n new · s stats · ? help · q quit ",
	" ? help · q quit ",
}

// renderStatusBar shows the most detailed hint line that fits.
func (m Model) renderStatusBar() string {
	hint := statusHints[len(statusHints)-1]
	for _, candidate := range statusHints {
		// A zero width means the terminal size is not known yet; assume room.
		if m.width == 0 || lipgloss.Width(candidate) <= m.width {
			hint = candidate
			break
		}
	}
	return m.styles.StatusBar.Render(hint)
}

func (m Model) renderOverlay(title, body string) string {
	return m.styles.Overlay.Render(m.styles.Title.Render(title) + "\n\n" + body)
}

func (m Model) renderDifficultyMenu() string {
	lines := make([]string, 0, len(menuPresets)+2)
	for i, p := range menuPresets {
		d := game.PresetDifficulty(p)
		prefix := "  "
		if i == m.menuIndex {
			prefix = "> "
		}
		lines = append(lines, fmt.Sprintf("%s%-13s %dx%d, %d mines",
			prefix, p.String(), d.Width, d.Height, d.Mines))
	}
	lines = append(lines, "", "enter select · esc back")
	return m.renderOverlay("Difficulty", strings.Join(lines, "\n"))
}

func (m Model) renderStats() string {
	rows := []string{fmt.Sprintf("%-13s %6s %6s %8s %7s %7s",
		"", "played", "won", "win rate", "avg", "streak")}

	for _, p := range menuPresets {
		t := m.stats.For(game.PresetDifficulty(p).Key())
		rows = append(rows, fmt.Sprintf("%-13s %6d %6d %7.0f%% %7s %7s",
			p.String(), t.Played, t.Won, t.WinRate()*100,
			secondsLabel(t.AverageWinSeconds()), streakLabel(t)))
	}

	rows = append(rows, "", "Only finished games count. Starting a new board mid-game does not.")
	return m.renderOverlay("Statistics", strings.Join(rows, "\n"))
}

// secondsLabel renders a duration that may not exist yet.
func secondsLabel(sec int) string {
	if sec < 0 {
		return "—"
	}
	return fmt.Sprintf("%ds", sec)
}

// streakLabel shows the run in progress against the best one so far.
func streakLabel(t storage.Tally) string {
	return fmt.Sprintf("%d/%d", t.CurrentStreak, t.BestStreak)
}

func (m Model) renderHelp() string {
	body := strings.Join(m.keys.helpLines(), "\n") +
		"\n\n" + m.markCycleHelp() +
		"\nMouse: left reveals, shift+left or right marks." +
		"\nSome terminals paste on right-click; shift+left always works."
	return m.renderOverlay("Help", body)
}

// markCycleHelp describes what f does, which depends on whether question marks
// are enabled in the config.
func (m Model) markCycleHelp() string {
	if m.config.QuestionMarks {
		return "Marks cycle flag → ? → clear. A ? is only a note: it does not\n" +
			"stop a reveal and does not count towards a chord."
	}
	return "Marks toggle between flag and clear."
}

func (m Model) renderGameOver() string {
	return m.renderOverlay("Game Over", fmt.Sprintf(
		"You hit a mine after %d seconds.\n\n%s", m.elapsed, m.retryHint()))
}

func (m Model) renderWin() string {
	body := fmt.Sprintf("Cleared in %d seconds.", m.elapsed)
	if b := m.highscores.Best(m.difficulty.Key()); b >= 0 {
		body += fmt.Sprintf("\nBest: %d seconds.", b)
	}
	return m.renderOverlay("Victory", body+"\n\n"+m.retryHint())
}

func (m Model) retryHint() string {
	return fmt.Sprintf("n: new board · r: replay seed %s · q: quit", m.board.Seed())
}
