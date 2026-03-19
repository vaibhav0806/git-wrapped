package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/vaibhav/gh-wrapped/github"
)

// SlideID identifies each slide.
type SlideID int

const (
	SlideTitle       SlideID = iota // 0
	SlideNumbers                    // 1
	SlideHeatmap                    // 2
	SlideChaos                      // 3
	SlideClock                      // 4
	SlideLanguages                  // 5
	SlideVillain                    // 6
	SlideWeekend                    // 7
	SlideNovel                      // 8
	SlidePersonality                // 9
)

// ActiveSlides returns the ordered list of slides that have enough data to show.
func ActiveSlides(s github.Stats) []SlideID {
	slides := []SlideID{SlideTitle}

	if s.HasCalendar {
		slides = append(slides, SlideNumbers)
		slides = append(slides, SlideHeatmap)
		slides = append(slides, SlideChaos)
	}

	if s.TimeBlocks[0]+s.TimeBlocks[1]+s.TimeBlocks[2]+s.TimeBlocks[3] > 0 {
		slides = append(slides, SlideClock)
	}

	if len(s.Languages) > 0 {
		slides = append(slides, SlideLanguages)
	}

	if s.VillainCommits > 0 {
		slides = append(slides, SlideVillain)
	}

	if s.HasCalendar {
		slides = append(slides, SlideWeekend)
	}

	if s.LongestMessage != "" {
		slides = append(slides, SlideNovel)
	}

	slides = append(slides, SlidePersonality)

	return slides
}

// RenderSlide dispatches to the appropriate render function and centers the
// content vertically within the given terminal dimensions.
func RenderSlide(id SlideID, s github.Stats, anim AnimState, width, height int) string {
	var content string
	switch id {
	case SlideTitle:
		content = renderTitle(s, anim, width)
	case SlideNumbers:
		content = renderNumbers(s, anim, width)
	case SlideHeatmap:
		content = renderHeatmap(s, anim, width)
	case SlideChaos:
		content = renderChaos(s, anim, width)
	case SlideClock:
		content = renderClock(s, anim, width)
	case SlideLanguages:
		content = renderLanguages(s, anim, width)
	case SlideVillain:
		content = renderVillain(s, anim, width)
	case SlideWeekend:
		content = renderWeekend(s, anim, width)
	case SlideNovel:
		content = renderNovel(s, anim, width)
	case SlidePersonality:
		content = renderPersonality(s, anim, width)
	default:
		content = ""
	}

	// Center vertically.
	lines := strings.Split(content, "\n")
	contentHeight := len(lines)
	paddingTop := (height - contentHeight) / 2
	if paddingTop < 0 {
		paddingTop = 0
	}
	var sb strings.Builder
	for i := 0; i < paddingTop; i++ {
		sb.WriteByte('\n')
	}
	sb.WriteString(content)
	return sb.String()
}

// ---------------------------------------------------------------------------
// Individual slide renderers
// ---------------------------------------------------------------------------

func renderTitle(s github.Stats, anim AnimState, width int) string {
	progress := anim.Progress()

	heading := GradientText(fmt.Sprintf("DEV WRAPPED %s", s.YearLabel), ColorRed, ColorCyan)
	heading = TitleStyle.Width(width).Render(heading)

	handle := TypewriterAnimation(fmt.Sprintf("@%s", s.Username), progress)
	handle = SubtitleStyle.Width(width).Render(handle)

	footnote := FootnoteStyle.Width(width).Render("your year in code")

	return strings.Join([]string{
		"",
		heading,
		"",
		handle,
		"",
		footnote,
	}, "\n")
}

func renderNumbers(s github.Stats, anim AnimState, width int) string {
	progress := anim.Progress()

	contribs := CounterAnimation(s.TotalContributions, progress)
	repos := CounterAnimation(s.TotalRepos, progress)
	stars := CounterAnimation(s.TotalStars, progress)

	title := GradientText("YEAR IN NUMBERS", ColorPurple, ColorCyan)
	title = TitleStyle.Width(width).Render(title)

	numberStyle := BigNumberStyle.Width(width)
	labelStyle := LabelStyle.Width(width)

	return strings.Join([]string{
		"",
		title,
		"",
		numberStyle.Foreground(ColorCyan).Render(fmt.Sprintf("%d", contribs)),
		labelStyle.Render("contributions"),
		"",
		numberStyle.Foreground(ColorPurple).Render(fmt.Sprintf("%d", repos)),
		labelStyle.Render("repositories"),
		"",
		numberStyle.Foreground(ColorGreen).Render(fmt.Sprintf("%d", stars)),
		labelStyle.Render("stars earned"),
	}, "\n")
}

func renderHeatmap(s github.Stats, anim AnimState, width int) string {
	progress := anim.Progress()

	// Heatmap characters by display intensity (0-4).
	levelChars := []rune{' ', '░', '▒', '▓', '█'}
	levelColors := []lipgloss.Color{ColorDim, ColorDim, ColorGreen, ColorGreen, ColorCyan}

	// Determine how many cells to reveal.
	totalCells := len(s.Calendar)
	revealed := HeatmapAnimation(totalCells, progress)

	title := GradientText("CONTRIBUTION HEATMAP", ColorGreen, ColorCyan)
	title = TitleStyle.Width(width).Render(title)

	// Render calendar grid — 53 weeks wide, up to 7 rows tall.
	cols := 53
	var gridLines []string
	for row := 0; row < 7; row++ {
		var rowSb strings.Builder
		for col := 0; col < cols; col++ {
			idx := col*7 + row
			if idx >= totalCells {
				rowSb.WriteString("  ")
				continue
			}
			if idx >= revealed {
				rowSb.WriteString(lipgloss.NewStyle().Foreground(ColorDim).Render("░ "))
				continue
			}
			day := s.Calendar[idx]
			// Derive display level from Count when Level is 0 but Count > 0.
			displayLevel := day.Level
			if displayLevel == 0 && day.Count > 0 {
				switch {
				case day.Count >= 20:
					displayLevel = 4
				case day.Count >= 10:
					displayLevel = 3
				case day.Count >= 5:
					displayLevel = 2
				default:
					displayLevel = 1
				}
			}
			if displayLevel < 0 {
				displayLevel = 0
			}
			if displayLevel > 4 {
				displayLevel = 4
			}
			ch := string(levelChars[displayLevel])
			color := levelColors[displayLevel]
			rowSb.WriteString(lipgloss.NewStyle().Foreground(color).Render(ch + " "))
		}
		gridLines = append(gridLines, lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(rowSb.String()))
	}

	streakLine := ""
	if s.LongestStreak > 0 {
		streakLine = SubtitleStyle.Width(width).Render(
			fmt.Sprintf("longest streak: %d days  %s → %s",
				s.LongestStreak,
				s.StreakStart.Format("Jan 2"),
				s.StreakEnd.Format("Jan 2"),
			),
		)
	}

	parts := []string{"", title, ""}
	parts = append(parts, gridLines...)
	parts = append(parts, "", streakLine)
	return strings.Join(parts, "\n")
}

func renderChaos(s github.Stats, anim AnimState, width int) string {
	progress := anim.Progress()

	quotes := []string{
		"someone had a deadline.",
		"chaos is a ladder.",
		"we don't talk about this day.",
		"absolutely unhinged behavior.",
		"the git log doesn't lie.",
	}
	// Pick a deterministic quote from the busiest date.
	quoteIdx := s.BusiestDate.Day() % len(quotes)
	quote := quotes[quoteIdx]

	title := GradientText("MOST CHAOTIC DAY", ColorYellow, ColorRed)
	title = TitleStyle.Width(width).Render(title)

	dateStr := TypewriterAnimation(s.BusiestDate.Format("Monday, Jan 2"), progress)
	dateRendered := BigNumberStyle.Width(width).Foreground(ColorYellow).Render(dateStr)

	countStr := fmt.Sprintf("%d contributions", CounterAnimation(s.BusiestCount, progress))
	countRendered := BigNumberStyle.Width(width).Foreground(ColorRed).Render(countStr)

	quoteRendered := FootnoteStyle.Width(width).Render(fmt.Sprintf("« %s »", quote))

	return strings.Join([]string{
		"",
		title,
		"",
		dateRendered,
		"",
		countRendered,
		"",
		quoteRendered,
	}, "\n")
}

func renderClock(s github.Stats, anim AnimState, width int) string {
	progress := anim.Progress()

	title := GradientText("WHEN YOU CODE", ColorPink, ColorPurple)
	title = TitleStyle.Width(width).Render(title)

	labels := [4]string{"Morning (6-12)", "Afternoon (12-18)", "Evening (18-24)", "Night (0-6)"}
	colors := [4]lipgloss.Color{ColorYellow, ColorCyan, ColorPurple, ColorPink}

	// Find max for normalization.
	maxVal := 1
	for _, v := range s.TimeBlocks {
		if v > maxVal {
			maxVal = v
		}
	}

	barAreaWidth := width / 2
	if barAreaWidth < 10 {
		barAreaWidth = 10
	}

	var rows []string
	for i, count := range s.TimeBlocks {
		animated := CounterAnimation(count, progress)
		filledWidth := 0
		if maxVal > 0 {
			filledWidth = int(float64(barAreaWidth) * float64(animated) / float64(maxVal))
		}
		emptyWidth := barAreaWidth - filledWidth

		bar := RenderBar(filledWidth, colors[i]) + RenderBarEmpty(emptyWidth)
		label := LabelStyle.Render(fmt.Sprintf("%-18s %4d", labels[i], animated))
		row := lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(bar + "  " + label)
		rows = append(rows, row)
	}

	verdict := SubtitleStyle.Width(width).Render(s.TimeLabel)

	parts := []string{"", title, ""}
	parts = append(parts, rows...)
	parts = append(parts, "", verdict)
	return strings.Join(parts, "\n")
}

func renderLanguages(s github.Stats, anim AnimState, width int) string {
	progress := anim.Progress()

	title := GradientText("TOP LANGUAGES", ColorRed, ColorYellow)
	title = TitleStyle.Width(width).Render(title)

	barAreaWidth := width / 2
	if barAreaWidth < 10 {
		barAreaWidth = 10
	}

	limit := 5
	if len(s.Languages) < limit {
		limit = len(s.Languages)
	}

	var rows []string
	for i := 0; i < limit; i++ {
		lang := s.Languages[i]
		color := lipgloss.Color(lang.Color)
		if lang.Color == "" {
			color = ColorMuted
		}

		animatedPct := lang.Percent * progress
		filledWidth := int(float64(barAreaWidth) * animatedPct / 100.0)
		emptyWidth := barAreaWidth - filledWidth

		bar := RenderBar(filledWidth, color) + RenderBarEmpty(emptyWidth)
		label := LabelStyle.Render(fmt.Sprintf("%-14s %5.1f%%", lang.Name, lang.Percent))
		row := lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(bar + "  " + label)
		rows = append(rows, row)
	}

	parts := []string{"", title, ""}
	parts = append(parts, rows...)
	return strings.Join(parts, "\n")
}

func renderVillain(s github.Stats, anim AnimState, width int) string {
	progress := anim.Progress()

	title := GradientText("VILLAIN ARC", ColorCrimson, ColorRed)
	title = TitleStyle.Width(width).Render(title)

	commitCount := CounterAnimation(s.VillainCommits, progress)
	countRendered := BigNumberStyle.Width(width).Foreground(ColorCrimson).Render(
		fmt.Sprintf("%d commits", commitCount),
	)

	repoStr := TypewriterAnimation(s.VillainRepo, progress)
	repoRendered := SubtitleStyle.Width(width).Render(fmt.Sprintf("to %s", repoStr))

	tag := FootnoteStyle.Width(width).Render("obsessed much?")

	return strings.Join([]string{
		"",
		title,
		"",
		countRendered,
		repoRendered,
		"",
		tag,
	}, "\n")
}

func renderWeekend(s github.Stats, anim AnimState, width int) string {
	progress := anim.Progress()

	title := GradientText("WEEKEND WARRIOR", ColorCyan, ColorGreen)
	title = TitleStyle.Width(width).Render(title)

	animatedPct := s.WeekendPercent * progress
	weekendWidth := int(float64(width/2) * animatedPct / 100.0)
	weekdayWidth := width/2 - weekendWidth

	weekendBar := RenderBar(weekendWidth, ColorCyan)
	weekdayBar := RenderBarEmpty(weekdayWidth)
	bar := lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(weekendBar + weekdayBar)

	pctLabel := SubtitleStyle.Width(width).Render(
		fmt.Sprintf("%.1f%% of commits on weekends", animatedPct),
	)

	var verdict string
	switch {
	case s.WeekendPercent >= 50:
		verdict = "you live for the weekend."
	case s.WeekendPercent >= 25:
		verdict = "work hard, push harder."
	default:
		verdict = "strictly business. respect."
	}
	verdictRendered := FootnoteStyle.Width(width).Render(verdict)

	return strings.Join([]string{
		"",
		title,
		"",
		bar,
		"",
		pctLabel,
		verdictRendered,
	}, "\n")
}

func renderNovel(s github.Stats, anim AnimState, width int) string {
	progress := anim.Progress()

	title := GradientText("COMMIT AS NOVEL", ColorPurple, ColorPink)
	title = TitleStyle.Width(width).Render(title)

	boxWidth := width - 8
	if boxWidth < 20 {
		boxWidth = 20
	}

	wrapped := wordWrap(s.LongestMessage, boxWidth-4)
	animated := TypewriterAnimation(wrapped, progress)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPurple).
		Padding(1, 2).
		Width(boxWidth).
		Align(lipgloss.Left)

	box := lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(
		boxStyle.Render(animated),
	)

	charCount := SubtitleStyle.Width(width).Render(
		fmt.Sprintf("%d characters · %s", s.LongestMessageLen, s.LongestMessageRepo),
	)

	return strings.Join([]string{
		"",
		title,
		"",
		box,
		"",
		charCount,
	}, "\n")
}

func renderPersonality(s github.Stats, anim AnimState, width int) string {
	progress := anim.Progress()

	title := GradientText("YOUR DEV PERSONALITY", ColorGreen, ColorCyan)
	title = TitleStyle.Width(width).Render(title)

	archetype := TypewriterAnimation(s.Archetype, progress)
	archetypeGradient := GradientText(strings.ToUpper(archetype), ColorCyan, ColorGreen)
	archetypeRendered := BigNumberStyle.Width(width).Render(archetypeGradient)

	pillColors := [3]lipgloss.Color{ColorPurple, ColorPink, ColorYellow}
	var pills []string
	for i, trait := range s.Traits {
		animated := TypewriterAnimation(trait, progress)
		if animated == "" {
			continue
		}
		pill := PillStyle.Foreground(pillColors[i%3]).Render(animated)
		pills = append(pills, pill)
	}
	pillRow := lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(
		strings.Join(pills, "  "),
	)

	gif := FootnoteStyle.Width(width).Render("press g for GIF")

	return strings.Join([]string{
		"",
		title,
		"",
		archetypeRendered,
		"",
		pillRow,
		"",
		gif,
	}, "\n")
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// wordWrap wraps s at maxWidth characters, breaking at word boundaries.
func wordWrap(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return s
	}
	words := strings.Fields(s)
	if len(words) == 0 {
		return s
	}

	var sb strings.Builder
	lineLen := 0
	for i, word := range words {
		wl := len(word)
		if i == 0 {
			sb.WriteString(word)
			lineLen = wl
			continue
		}
		if lineLen+1+wl > maxWidth {
			sb.WriteByte('\n')
			sb.WriteString(word)
			lineLen = wl
		} else {
			sb.WriteByte(' ')
			sb.WriteString(word)
			lineLen += 1 + wl
		}
	}
	return sb.String()
}
