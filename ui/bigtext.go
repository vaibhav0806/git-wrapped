package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Pre-built digit patterns using block characters (3 lines tall).
var digitPatterns = map[rune][3]string{
	'0': {"█▀█", "█ █", "▀▀▀"},
	'1': {" █ ", " █ ", " ▀ "},
	'2': {"▀▀█", "█▀▀", "▀▀▀"},
	'3': {"▀▀█", " ▀█", "▀▀▀"},
	'4': {"█ █", "▀▀█", "  ▀"},
	'5': {"█▀▀", "▀▀█", "▀▀▀"},
	'6': {"█▀▀", "█▀█", "▀▀▀"},
	'7': {"▀▀█", "  █", "  ▀"},
	'8': {"█▀█", "█▀█", "▀▀▀"},
	'9': {"█▀█", "▀▀█", "▀▀▀"},
	',': {"   ", "   ", " ▄ "},
	'.': {"   ", "   ", " ▀ "},
	' ': {"   ", "   ", "   "},
}

// RenderBigNumber renders a number string in large 3-line block text.
func RenderBigNumber(s string, color lipgloss.Color) string {
	lines := [3]strings.Builder{}

	for _, ch := range s {
		pattern, ok := digitPatterns[ch]
		if !ok {
			pattern = [3]string{"   ", "   ", "   "}
		}
		for row := 0; row < 3; row++ {
			if lines[row].Len() > 0 {
				lines[row].WriteString(" ")
			}
			lines[row].WriteString(pattern[row])
		}
	}

	style := lipgloss.NewStyle().Foreground(color).Bold(true)
	var result []string
	for row := 0; row < 3; row++ {
		result = append(result, style.Render(lines[row].String()))
	}
	return strings.Join(result, "\n")
}
