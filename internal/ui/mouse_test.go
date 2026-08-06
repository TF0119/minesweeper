package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestClassifyMouseEvent(t *testing.T) {
	tests := []struct {
		name   string
		msg    tea.MouseMsg
		last   tea.MouseButton
		want   mouseClick
		wantSt tea.MouseButton
	}{
		{
			name:   "right press waits for release",
			msg:    tea.MouseMsg{Button: tea.MouseButtonRight, Action: tea.MouseActionPress},
			want:   mouseClick{},
			wantSt: tea.MouseButtonRight,
		},
		{
			name: "right release sgr",
			msg:  tea.MouseMsg{Button: tea.MouseButtonRight, Action: tea.MouseActionRelease},
			want: mouseClick{flag: true},
		},
		{
			name: "middle release chords",
			msg:  tea.MouseMsg{Button: tea.MouseButtonMiddle, Action: tea.MouseActionRelease},
			want: mouseClick{chord: true},
		},
		{
			name:   "middle press waits for release",
			msg:    tea.MouseMsg{Button: tea.MouseButtonMiddle, Action: tea.MouseActionPress},
			want:   mouseClick{},
			wantSt: tea.MouseButtonMiddle,
		},
		{
			name: "legacy middle type",
			msg:  tea.MouseMsg{Type: tea.MouseMiddle, Action: tea.MouseActionPress},
			want: mouseClick{chord: true},
		},
		{
			name: "x10 release uses last button",
			msg:  tea.MouseMsg{Button: tea.MouseButtonNone, Action: tea.MouseActionRelease, Type: tea.MouseRelease},
			last: tea.MouseButtonRight,
			want: mouseClick{flag: true},
		},
		{
			name: "x10 middle release uses last button",
			msg:  tea.MouseMsg{Button: tea.MouseButtonNone, Action: tea.MouseActionRelease, Type: tea.MouseRelease},
			last: tea.MouseButtonMiddle,
			want: mouseClick{chord: true},
		},
		{
			name:   "shift left flags",
			msg:    tea.MouseMsg{Button: tea.MouseButtonLeft, Action: tea.MouseActionPress, Shift: true},
			want:   mouseClick{flag: true},
			wantSt: tea.MouseButtonLeft,
		},
		{
			name:   "left reveals",
			msg:    tea.MouseMsg{Button: tea.MouseButtonLeft, Action: tea.MouseActionPress},
			want:   mouseClick{reveal: true},
			wantSt: tea.MouseButtonLeft,
		},
		{
			name: "legacy right type",
			msg:  tea.MouseMsg{Type: tea.MouseRight, Action: tea.MouseActionPress},
			want: mouseClick{flag: true},
		},
		{
			name: "motion ignored",
			msg:  tea.MouseMsg{Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion},
			want: mouseClick{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, st := classifyMouseEvent(tt.msg, tt.last)
			if got != tt.want {
				t.Errorf("click = %+v, want %+v", got, tt.want)
			}
			if tt.wantSt != 0 && st != tt.wantSt {
				t.Errorf("storeBtn = %v, want %v", st, tt.wantSt)
			}
		})
	}
}
