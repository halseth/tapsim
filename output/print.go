package output

import (
	"fmt"
	"strings"
)

// ESC is the ASCII code for escape character.
const ESC = 27

// clearCode defines a terminal escape code to clear the current line and move
// the cursor up.
const cursorToBeginning = "\u001B[0G"
const clearLine = "\u001B[2K"
const cursorUp = "\u001B[1A"

var clearCode = fmt.Sprintf("%s%s%s%s", cursorToBeginning, clearLine, cursorUp, clearLine)

// ClearLines erases the last count lines in the terminal window.
func ClearLines(count int) {
	_, _ = fmt.Print(strings.Repeat(clearCode, count))
}

// DrawTable prints the table line by line, clearing existing lines first.
func DrawTable(table string, clear int) {
	ClearLines(clear)
	lines := strings.Split(table, "\n")
	for _, l := range lines {
		fmt.Printf("%s\r\n", l)
	}
}
