package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Color constants
const (
	ColorRed     = lipgloss.Color("#ff6b6b")
	ColorYellow  = lipgloss.Color("#feca57")
	ColorGreen   = lipgloss.Color("#0be881")
	ColorCyan    = lipgloss.Color("#48dbfb")
	ColorPurple  = lipgloss.Color("#a29bfe")
	ColorPink    = lipgloss.Color("#ff9ff3")
	ColorCrimson = lipgloss.Color("#ff4757")
	ColorWhite   = lipgloss.Color("#e8e8e8")
	ColorDim     = lipgloss.Color("#555555")
	ColorMuted   = lipgloss.Color("#888888")
	ColorBg      = lipgloss.Color("#0d1117")
	ColorPanel   = lipgloss.Color("#161b22")
)

// SlideAccents maps slide index to accent color.
var SlideAccents = [10]lipgloss.Color{
	ColorCyan,
	ColorPurple,
	ColorGreen,
	ColorYellow,
	ColorPink,
	ColorRed,
	ColorCrimson,
	ColorCyan,
	ColorPurple,
	ColorGreen,
}

// Reusable styles
var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Align(lipgloss.Center)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Align(lipgloss.Center)

	BigNumberStyle = lipgloss.NewStyle().
			Bold(true).
			Align(lipgloss.Center)

	LabelStyle = lipgloss.NewStyle().
			Foreground(ColorDim).
			Align(lipgloss.Center)

	FootnoteStyle = lipgloss.NewStyle().
			Foreground(ColorDim).
			Italic(true).
			Align(lipgloss.Center)

	PillStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Bold(true)

	BarStyle = lipgloss.NewStyle()
)

// RenderBar renders a filled bar of the given width in the given color.
func RenderBar(width int, color lipgloss.Color) string {
	if width <= 0 {
		return ""
	}
	return BarStyle.Foreground(color).Render(strings.Repeat("█", width))
}

// RenderBarEmpty renders an unfilled (dim) bar of the given width.
func RenderBarEmpty(width int) string {
	if width <= 0 {
		return ""
	}
	return BarStyle.Foreground(ColorDim).Render(strings.Repeat("░", width))
}

