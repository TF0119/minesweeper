package ui

import (
	"fmt"
	"strings"
)

type settingKind int

const (
	settingTheme settingKind = iota
	settingNoGuess
	settingQuestionMarks
	settingEmoji
)

type settingDef struct {
	label string
	kind  settingKind
}

var settingDefs = []settingDef{
	{label: "Theme", kind: settingTheme},
	{label: "No-guess boards", kind: settingNoGuess},
	{label: "Question marks", kind: settingQuestionMarks},
	{label: "Emoji glyphs", kind: settingEmoji},
}

func (m Model) renderSettings() string {
	return m.renderMenuList("Settings", len(settingDefs), func(i int) string {
		d := settingDefs[i]
		return fmt.Sprintf("%-16s %s", d.label+":", m.settingValue(d.kind))
	})
}

func (m Model) settingValue(kind settingKind) string {
	switch kind {
	case settingTheme:
		return string(themeFromConfig(m.config))
	case settingNoGuess:
		return onOff(m.config.NoGuess)
	case settingQuestionMarks:
		return onOff(m.config.QuestionMarks)
	case settingEmoji:
		return onOff(m.config.UseEmoji)
	default:
		return ""
	}
}

func onOff(v bool) string {
	if v {
		return "on"
	}
	return "off"
}

func themeIndex(t Theme) int {
	for i, th := range Themes {
		if th == t {
			return i
		}
	}
	return 0
}

// cycleSetting advances the selected setting and rebuilds whatever part of the
// UI that setting affects. Visual prefs refresh immediately; no-guess applies
// to the next board only, which matches how -no-guess on the CLI works.
func (m Model) cycleSetting(kind settingKind) Model {
	switch kind {
	case settingTheme:
		next := Themes[(themeIndex(themeFromConfig(m.config))+1)%len(Themes)]
		m.config.Theme = string(next)
		m.styles = NewStyles(next, m.useColor)
	case settingNoGuess:
		m.config.NoGuess = !m.config.NoGuess
	case settingQuestionMarks:
		m.config.QuestionMarks = !m.config.QuestionMarks
	case settingEmoji:
		m.config.UseEmoji = !m.config.UseEmoji
		m.glyphs = newGlyphs(m.config.UseEmoji)
	}
	return m
}

func (m Model) renderMenuList(title string, count int, line func(int) string) string {
	lines := make([]string, 0, count+2)
	for i := 0; i < count; i++ {
		prefix := "  "
		if i == m.menuIndex {
			prefix = "> "
		}
		lines = append(lines, prefix+line(i))
	}
	lines = append(lines, "", "enter select · esc back")
	return m.renderOverlay(title, strings.Join(lines, "\n"))
}
