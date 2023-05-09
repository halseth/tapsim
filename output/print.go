package output

import (
	"fmt"
	"strings"
)

// ESC is the ASCII code for escape character.
const ESC = 27

// clearCode defines a terminal escape code to clear the currently line and move
// the cursor up.
var clearCode = fmt.Sprintf("%c[%dA%c[2K", ESC, 1, ESC)

// clearLines erases the last count lines in the terminal window.
func clearLines(count int) {
	_, _ = fmt.Print(strings.Repeat(clearCode, count))
}

// DrawTable prints the table line by line, clearing existing lines first.
func DrawTable(table string, clear int) {
	clearLines(clear)
	lines := strings.Split(table, "\n")
	for _, l := range lines {
		fmt.Printf("%s\r\n", l)
	}
}
