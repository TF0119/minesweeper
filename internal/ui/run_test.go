package ui

import "testing"

func TestColorDisabledByEnv(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want bool
	}{
		{"unset", nil, false},
		{"colour terminal", map[string]string{"TERM": "xterm-256color"}, false},
		{"NO_COLOR set", map[string]string{"NO_COLOR": "1"}, true},
		{"NO_COLOR any value", map[string]string{"NO_COLOR": "0"}, true},
		{"NO_COLOR empty is not an opt-out", map[string]string{"NO_COLOR": ""}, false},
		{"dumb terminal", map[string]string{"TERM": "dumb"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getenv := func(k string) string { return tt.env[k] }
			if got := colorDisabledByEnv(getenv); got != tt.want {
				t.Errorf("colorDisabledByEnv(%v) = %v, want %v", tt.env, got, tt.want)
			}
		})
	}
}
