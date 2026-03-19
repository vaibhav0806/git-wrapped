package ui

import (
	"fmt"
	"strconv"
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

// GradientText applies a per-character linear color gradient from `from` to `to`.
func GradientText(text string, from, to lipgloss.Color) string {
	runes := []rune(text)
	n := len(runes)
	if n == 0 {
		return ""
	}

	r1, g1, b1 := hexToRGB(string(from))
	r2, g2, b2 := hexToRGB(string(to))

	var sb strings.Builder
	for i, ch := range runes {
		var t float64
		if n > 1 {
			t = float64(i) / float64(n-1)
		}
		r := lerp(r1, r2, t)
		g := lerp(g1, g2, t)
		b := lerp(b1, b2, t)
		color := lipgloss.Color(rgbToHex(r, g, b))
		sb.WriteString(lipgloss.NewStyle().Foreground(color).Render(string(ch)))
	}
	return sb.String()
}

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

// hexToRGB converts a hex color string (#rrggbb or rrggbb) to r, g, b in [0,255].
func hexToRGB(hex string) (r, g, b uint8) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 0, 0, 0
	}
	val, err := strconv.ParseUint(hex, 16, 32)
	if err != nil {
		return 0, 0, 0
	}
	r = uint8(val >> 16)
	g = uint8((val >> 8) & 0xff)
	b = uint8(val & 0xff)
	return
}

// rgbToHex converts r, g, b in [0,255] to a #rrggbb hex string.
func rgbToHex(r, g, b uint8) string {
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// lerp linearly interpolates between a and b by t in [0,1].
func lerp(a, b uint8, t float64) uint8 {
	// Use float64 arithmetic to avoid uint8 underflow when b < a.
	return uint8(float64(a)*(1-t) + float64(b)*t)
}
