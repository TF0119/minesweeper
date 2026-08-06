package ui

// hubAction is what a hub menu entry does. Keeping actions as data lets the
// renderer and key handler share one ordered list without a switch per label.
type hubAction int

const (
	hubResume hubAction = iota
	hubNewGame
	hubDaily
	hubDifficulty
	hubStatistics
	hubSettings
	hubHelp
	hubReplays
	hubQuit
)

type hubItem struct {
	label  string
	action hubAction
}

// hubMenuItems is the in-game menu. It is never shown at startup; the player
// opens it when they want settings, stats, or a new board without quitting.
var hubMenuItems = []hubItem{
	{label: "Resume", action: hubResume},
	{label: "New game", action: hubNewGame},
	{label: "Daily challenge", action: hubDaily},
	{label: "Difficulty", action: hubDifficulty},
	{label: "Statistics", action: hubStatistics},
	{label: "Settings", action: hubSettings},
	{label: "Watch", action: hubReplays},
	{label: "Help", action: hubHelp},
	{label: "Quit", action: hubQuit},
}

func (m Model) renderHubMenu() string {
	return m.renderMenuList("Menu", len(hubMenuItems), func(i int) string {
		return hubMenuItems[i].label
	})
}
