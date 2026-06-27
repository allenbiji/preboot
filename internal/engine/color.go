package engine

import (
	"os"

	"github.com/mattn/go-isatty"
)

// colorEnabled is true when stdout is connected to a real terminal.
// Declared as a var so tests can override it.
var colorEnabled = isatty.IsTerminal(os.Stdout.Fd()) &&
	os.Getenv("NO_COLOR") == "" &&
	os.Getenv("TERM") != "dumb"

// colorize wraps text with the given ANSI escape code only when color is active.
func colorize(code, text string) string {
	if !colorEnabled {
		return text
	}
	return code + text + Reset
}
