# VHS script for docs/demo.gif — https://github.com/charmbracelet/vhs
#
# Regenerate with:
#   go build -o bin/minesweeper ./cmd/minesweeper && vhs docs/demo.tape
#
# Seed 2 on Beginner is used so the recording is reproducible: the opening move
# floods a large area and leaves a chord that opens three more cells.

Output docs/demo.gif

Set Shell bash
Set FontSize 18
Set Width 1240
Set Height 336
Set Padding 20
Set Framerate 24
Set TypingSpeed 60ms

Hide
Type "export PATH=$PWD/bin:$PATH PS1='$ '" Enter
Type "clear" Enter
Show

Sleep 500ms
Type "minesweeper -difficulty beginner -seed 2"
Sleep 400ms
Enter
Sleep 1.2s

# Opening move is always safe, and zero cells cascade.
Space
Sleep 1.6s

# Walk to the cell next to the deduced mine.
Left
Sleep 200ms
Left
Sleep 200ms
Left
Sleep 200ms
Up
Sleep 700ms

# Flag it.
Type "f"
Sleep 1s

# Chord on the adjacent 1 to open the rest.
Right
Sleep 400ms
Type "c"
Sleep 2.5s

Type "q"
Sleep 800ms
