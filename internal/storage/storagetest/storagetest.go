// Package storagetest keeps tests off the player's real save files.
package storagetest

import "testing"

// IsolateConfigDir points the config directory at a temporary one for the rest
// of the test.
//
// os.UserConfigDir consults a different variable on each platform, so setting
// only XDG_CONFIG_HOME isolated Linux and left macOS and Windows reading and
// writing the real config, high scores, and saved game — both clobbering the
// developer's files and letting one test see another's leftovers.
func IsolateConfigDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir) // Unix
	t.Setenv("HOME", dir)            // macOS, under Library/Application Support
	t.Setenv("AppData", dir)         // Windows
	return dir
}
