package ui

import tea "github.com/charmbracelet/bubbletea"

// mouseClick describes an actionable mouse click, if any.
type mouseClick struct {
	flag   bool
	reveal bool
	chord  bool
}

// classifyMouseEvent normalizes terminal-specific mouse encodings.
// Shift+left-click flags cells when the terminal steals right-click (paste).
func classifyMouseEvent(msg tea.MouseMsg, lastBtn tea.MouseButton) (click mouseClick, storeBtn tea.MouseButton) {
	storeBtn = lastBtn

	if msg.Action != tea.MouseActionPress && msg.Action != tea.MouseActionRelease {
		return mouseClick{}, storeBtn
	}

	btn := msg.Button
	if btn == tea.MouseButtonNone {
		btn = buttonFromLegacyType(msg.Type)
	}
	if btn == tea.MouseButtonNone && msg.Action == tea.MouseActionRelease {
		btn = lastBtn
	}
	if btn != tea.MouseButtonNone && msg.Action == tea.MouseActionPress {
		storeBtn = btn
	}

	// Shift+left: flag (WSL / Windows Terminal often paste on right-click).
	// Act on press only; release would toggle twice.
	if msg.Shift && btn == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress {
		return mouseClick{flag: true}, storeBtn
	}

	switch btn {
	case tea.MouseButtonRight:
		switch msg.Action {
		case tea.MouseActionRelease:
			return mouseClick{flag: true}, storeBtn
		case tea.MouseActionPress:
			// X10 terminals sometimes emit only a press with the legacy Type set.
			if msg.Button == tea.MouseButtonNone && buttonFromLegacyType(msg.Type) != tea.MouseButtonNone {
				return mouseClick{flag: true}, storeBtn
			}
		}
	case tea.MouseButtonMiddle:
		switch msg.Action {
		case tea.MouseActionRelease:
			return mouseClick{chord: true}, storeBtn
		case tea.MouseActionPress:
			if msg.Button == tea.MouseButtonNone && buttonFromLegacyType(msg.Type) != tea.MouseButtonNone {
				return mouseClick{chord: true}, storeBtn
			}
		}
	case tea.MouseButtonLeft:
		if msg.Action == tea.MouseActionPress {
			return mouseClick{reveal: true}, storeBtn
		}
	}
	return mouseClick{}, storeBtn
}

func buttonFromLegacyType(t tea.MouseEventType) tea.MouseButton {
	switch t {
	case tea.MouseLeft:
		return tea.MouseButtonLeft
	case tea.MouseRight:
		return tea.MouseButtonRight
	case tea.MouseMiddle:
		return tea.MouseButtonMiddle
	default:
		return tea.MouseButtonNone
	}
}

func isMouseClick(msg tea.MouseMsg) bool {
	return msg.Action == tea.MouseActionPress || msg.Action == tea.MouseActionRelease
}
