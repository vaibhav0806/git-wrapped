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
	SlideTitle       SlideID = iota
	SlideNumbers
	SlideHeatmap
	SlideChaos
	SlideClock
	SlideLanguages
	SlideVillain
	SlideWeekend
	SlideNovel
	SlidePersonality
)

// ActiveSlides returns the ordered list of slides that have enough data to show.
func ActiveSlides(s github.Stats) []SlideID {
	slides := []SlideID{SlideTitle}
	if s.HasCalendar {
		slides = append(slides, SlideNumbers, SlideHeatmap, SlideChaos)
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

// RenderSlide renders a slide centered in a panel.
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
	}

	// Center vertically
	lines := strings.Split(content, "\n")
	padTop := (height - len(lines)) / 2
	if padTop < 0 {
		padTop = 0
	}
	return strings.Repeat("\n", padTop) + content
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// panel wraps content in a centered bordered box with accent color.
func panel(content string, accent lipgloss.Color, width int) string {
	boxWidth := width - 4
	if boxWidth > 80 {
		boxWidth = 80
	}
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accent).
		Padding(1, 3).
		Width(boxWidth).
		Align(lipgloss.Center).
		Render(content)
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(box)
}

// center centers text within width.
func center(s string, width int) string {
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(s)
}

// bigText renders text large and bold with color.
func bigText(s string, color lipgloss.Color) string {
	return lipgloss.NewStyle().
		Foreground(color).
		Bold(true).
		Render(s)
}

// dimText renders dim muted text.
func dimText(s string) string {
	return lipgloss.NewStyle().Foreground(ColorDim).Render(s)
}

// mutedText renders muted text.
func mutedText(s string) string {
	return lipgloss.NewStyle().Foreground(ColorMuted).Render(s)
}

// accentLine renders a decorative line.
func accentLine(width int, color lipgloss.Color) string {
	line := strings.Repeat("─", width)
	return lipgloss.NewStyle().Foreground(color).Render(line)
}

// statBlock renders a number + label vertically.
func statBlock(value string, label string, color lipgloss.Color) string {
	num := lipgloss.NewStyle().Foreground(color).Bold(true).Render(value)
	lbl := lipgloss.NewStyle().Foreground(ColorMuted).Render(label)
	return num + "\n" + lbl
}

// pill renders a colored pill/tag.
func pill(text string, fg lipgloss.Color, bg lipgloss.Color) string {
	return lipgloss.NewStyle().
		Foreground(fg).
		Background(bg).
		Bold(true).
		Padding(0, 2).
		Render(text)
}

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

// ---------------------------------------------------------------------------
// Slide renderers
// ---------------------------------------------------------------------------

func renderTitle(s github.Stats, anim AnimState, width int) string {
	p := anim.Progress()

	// Big sparkle decorations
	sparkle := dimText("✦  ✦  ✦")

	// Main title with gradient
	title := GradientText("  D E V   W R A P P E D  ", ColorRed, ColorCyan)
	titleStyled := lipgloss.NewStyle().Bold(true).Render(title)

	// Year with accent
	year := bigText(s.YearLabel, ColorCyan)

	// Username with typewriter
	handle := TypewriterAnimation("@"+s.Username, p)
	handleStyled := lipgloss.NewStyle().Foreground(ColorPurple).Bold(true).Render(handle)

	// Subtitle
	sub := mutedText("your year in code")

	inner := strings.Join([]string{
		"",
		sparkle,
		"",
		"",
		titleStyled,
		year,
		"",
		"",
		handleStyled,
		"",
		sub,
		"",
		sparkle,
		"",
	}, "\n")

	return panel(inner, ColorPurple, width)
}

func renderNumbers(s github.Stats, anim AnimState, width int) string {
	p := anim.Progress()

	contribs := CounterAnimation(s.TotalContributions, p)
	repos := CounterAnimation(s.TotalRepos, p)
	stars := CounterAnimation(s.TotalStars, p)

	heading := GradientText("YOUR YEAR IN NUMBERS", ColorYellow, ColorCyan)
	headingStyled := lipgloss.NewStyle().Bold(true).Render(heading)

	divider := accentLine(40, ColorDim)

	// Stat blocks side by side
	col1 := statBlock(fmt.Sprintf("%d", contribs), "contributions", ColorCyan)
	col2 := statBlock(fmt.Sprintf("%d", repos), "repositories", ColorPurple)
	col3 := statBlock(fmt.Sprintf("%d", stars), "stars earned", ColorYellow)

	statsRow := lipgloss.JoinHorizontal(lipgloss.Center,
		lipgloss.NewStyle().Width(22).Align(lipgloss.Center).Render(col1),
		lipgloss.NewStyle().Width(22).Align(lipgloss.Center).Render(col2),
		lipgloss.NewStyle().Width(22).Align(lipgloss.Center).Render(col3),
	)

	inner := strings.Join([]string{
		"",
		headingStyled,
		"",
		divider,
		"",
		"",
		statsRow,
		"",
		"",
		divider,
		"",
	}, "\n")

	return panel(inner, ColorYellow, width)
}

func renderHeatmap(s github.Stats, anim AnimState, width int) string {
	p := anim.Progress()

	heading := GradientText("CONTRIBUTION HEATMAP", ColorGreen, ColorCyan)
	headingStyled := lipgloss.NewStyle().Bold(true).Render(heading)

	totalCells := len(s.Calendar)
	revealed := HeatmapAnimation(totalCells, p)

	// Grid: 7 rows x up to 53 columns
	cols := 53
	boxWidth := width - 8
	if boxWidth > 80 {
		boxWidth = 80
	}
	// Scale columns to fit box
	maxCols := (boxWidth - 4) / 2
	if cols > maxCols {
		cols = maxCols
	}

	var gridLines []string
	for row := 0; row < 7; row++ {
		var rowSb strings.Builder
		for col := 0; col < cols; col++ {
			idx := row*53 + col // always use 53 as the original stride
			if idx >= totalCells {
				rowSb.WriteString("  ")
				continue
			}
			if idx >= revealed {
				rowSb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#1a1a2e")).Render("▪ "))
				continue
			}
			day := s.Calendar[idx]
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
			switch displayLevel {
			case 0:
				rowSb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#1a1a2e")).Render("▪ "))
			case 1:
				rowSb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#0e4429")).Render("▪ "))
			case 2:
				rowSb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#006d32")).Render("▪ "))
			case 3:
				rowSb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#26a641")).Render("▪ "))
			default:
				rowSb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#39d353")).Render("▪ "))
			}
		}
		gridLines = append(gridLines, rowSb.String())
	}

	grid := strings.Join(gridLines, "\n")

	// Streak callout
	streak := ""
	if s.LongestStreak > 0 {
		streakNum := bigText(fmt.Sprintf("%d", s.LongestStreak), ColorGreen)
		streak = streakNum + mutedText(" day streak") +
			dimText(fmt.Sprintf("  %s → %s", s.StreakStart.Format("Jan 2"), s.StreakEnd.Format("Jan 2")))
	}

	inner := strings.Join([]string{
		"",
		headingStyled,
		"",
		grid,
		"",
		streak,
		"",
	}, "\n")

	return panel(inner, ColorGreen, width)
}

func renderChaos(s github.Stats, anim AnimState, width int) string {
	p := anim.Progress()

	quotes := []string{
		"were you okay?",
		"someone had a deadline.",
		"chaos is a ladder.",
		"absolutely unhinged.",
		"the git log doesn't lie.",
	}
	quote := quotes[s.BusiestDate.Day()%len(quotes)]

	heading := lipgloss.NewStyle().Foreground(ColorRed).Bold(true).Render("YOUR MOST CHAOTIC DAY")

	date := TypewriterAnimation(s.BusiestDate.Format("January 2, 2006"), p)
	dateStyled := lipgloss.NewStyle().Foreground(ColorWhite).Bold(true).Render(date)

	count := CounterAnimation(s.BusiestCount, p)
	countStyled := lipgloss.NewStyle().Foreground(ColorRed).Bold(true).Render(fmt.Sprintf("%d", count))
	countLine := countStyled + mutedText(" contributions")

	quoteStyled := lipgloss.NewStyle().Foreground(ColorDim).Italic(true).Render(fmt.Sprintf("« %s »", quote))

	inner := strings.Join([]string{
		"",
		heading,
		"",
		"",
		dateStyled,
		"",
		countLine,
		"",
		"",
		quoteStyled,
		"",
	}, "\n")

	return panel(inner, ColorRed, width)
}

func renderClock(s github.Stats, anim AnimState, width int) string {
	p := anim.Progress()

	heading := GradientText("WHEN YOU CODE", ColorPurple, ColorPink)
	headingStyled := lipgloss.NewStyle().Bold(true).Render(heading)

	labels := [4]string{"Morning   ", "Afternoon ", "Evening   ", "Night     "}
	icons := [4]string{"☀️ ", "🌤 ", "🌙", "🌑"}
	colors := [4]lipgloss.Color{ColorYellow, ColorCyan, ColorPurple, ColorPink}

	maxVal := 1
	total := 0
	for _, v := range s.TimeBlocks {
		total += v
		if v > maxVal {
			maxVal = v
		}
	}

	barWidth := 30

	var rows []string
	for i, count := range s.TimeBlocks {
		animated := CounterAnimation(count, p)
		filled := 0
		if maxVal > 0 {
			filled = int(float64(barWidth) * float64(animated) / float64(maxVal))
		}
		empty := barWidth - filled

		pct := 0.0
		if total > 0 {
			pct = float64(count) / float64(total) * 100
		}

		icon := icons[i]
		label := lipgloss.NewStyle().Foreground(colors[i]).Bold(true).Render(labels[i])
		bar := RenderBar(filled, colors[i]) + RenderBarEmpty(empty)
		pctStr := dimText(fmt.Sprintf(" %4.0f%%", pct))

		rows = append(rows, icon+" "+label+" "+bar+pctStr)
	}

	divider := accentLine(50, ColorDim)

	// Verdict
	verdictText := s.TimeLabel
	verdictStyled := lipgloss.NewStyle().Foreground(ColorPurple).Bold(true).Render(verdictText)

	footnote := dimText("based on your last 30 days")

	inner := strings.Join([]string{
		"",
		headingStyled,
		"",
		divider,
		"",
	}, "\n")
	for _, r := range rows {
		inner += r + "\n"
	}
	inner += strings.Join([]string{
		"",
		divider,
		"",
		verdictStyled,
		footnote,
		"",
	}, "\n")

	return panel(inner, ColorPurple, width)
}

func renderLanguages(s github.Stats, anim AnimState, width int) string {
	p := anim.Progress()

	heading := GradientText("YOUR TOP LANGUAGES", ColorYellow, ColorRed)
	headingStyled := lipgloss.NewStyle().Bold(true).Render(heading)

	barWidth := 30
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

		animPct := lang.Percent * p
		filled := int(float64(barWidth) * animPct / 100.0)
		empty := barWidth - filled

		nameStyled := lipgloss.NewStyle().Foreground(color).Bold(true).Width(12).Render(lang.Name)
		bar := RenderBar(filled, color) + RenderBarEmpty(empty)
		pctStr := dimText(fmt.Sprintf(" %5.1f%%", lang.Percent))

		rows = append(rows, nameStyled+" "+bar+pctStr)
	}

	// Fun subtitle
	subtitle := ""
	if len(s.Languages) >= 4 && s.Languages[0].Percent < 40 {
		subtitle = mutedText("polyglot energy ✨")
	} else if len(s.Languages) > 0 && s.Languages[0].Percent > 70 {
		subtitle = mutedText(s.Languages[0].Name + " loyalist 💪")
	}

	divider := accentLine(50, ColorDim)

	inner := strings.Join([]string{
		"",
		headingStyled,
		"",
		divider,
		"",
	}, "\n")
	for _, r := range rows {
		inner += r + "\n"
	}
	inner += strings.Join([]string{
		"",
		divider,
		"",
		subtitle,
		"",
	}, "\n")

	return panel(inner, ColorRed, width)
}

func renderVillain(s github.Stats, anim AnimState, width int) string {
	p := anim.Progress()

	heading := lipgloss.NewStyle().Foreground(ColorCrimson).Bold(true).Render("⚡ VILLAIN ARC ⚡")

	count := CounterAnimation(s.VillainCommits, p)
	countStyled := lipgloss.NewStyle().Foreground(ColorCrimson).Bold(true).Render(fmt.Sprintf("%d", count))
	countLine := countStyled + mutedText(" commits")

	repo := TypewriterAnimation(s.VillainRepo, p)
	repoStyled := dimText("to ") + lipgloss.NewStyle().Foreground(ColorWhite).Bold(true).Render(repo)

	quote := lipgloss.NewStyle().Foreground(ColorDim).Italic(true).Render("« obsessed much? »")
	footnote := dimText("based on your last 30 days")

	inner := strings.Join([]string{
		"",
		heading,
		"",
		"",
		countLine,
		"",
		repoStyled,
		"",
		"",
		quote,
		"",
		footnote,
		"",
	}, "\n")

	return panel(inner, ColorCrimson, width)
}

func renderWeekend(s github.Stats, anim AnimState, width int) string {
	p := anim.Progress()

	heading := GradientText("WEEKEND WARRIOR", ColorPink, ColorCyan)
	headingStyled := lipgloss.NewStyle().Bold(true).Render(heading)

	animPct := s.WeekendPercent * p
	pctStyled := lipgloss.NewStyle().Foreground(ColorPink).Bold(true).Render(fmt.Sprintf("%.0f%%", animPct))
	pctLine := pctStyled + mutedText(" of your commits land on weekends")

	barWidth := 50
	weekendW := int(float64(barWidth) * animPct / 100.0)
	weekdayW := barWidth - weekendW
	bar := RenderBar(weekendW, ColorPink) + RenderBarEmpty(weekdayW)
	barLabels := dimText("weekdays") + strings.Repeat(" ", barWidth-16) + lipgloss.NewStyle().Foreground(ColorPink).Render("weekends")

	var verdict string
	switch {
	case s.WeekendPercent >= 50:
		verdict = "you live for the weekend. respect."
	case s.WeekendPercent >= 25:
		verdict = "work hard, push harder."
	default:
		verdict = "strictly business. nice."
	}
	verdictStyled := lipgloss.NewStyle().Foreground(ColorDim).Italic(true).Render(verdict)

	inner := strings.Join([]string{
		"",
		headingStyled,
		"",
		"",
		pctLine,
		"",
		bar,
		barLabels,
		"",
		"",
		verdictStyled,
		"",
	}, "\n")

	return panel(inner, ColorPink, width)
}

func renderNovel(s github.Stats, anim AnimState, width int) string {
	p := anim.Progress()

	heading := GradientText("THE NOVEL", ColorYellow, ColorPurple)
	headingStyled := lipgloss.NewStyle().Bold(true).Render(heading)
	sub := mutedText("your longest commit message")

	// Message box
	msgWidth := 60
	wrapped := wordWrap(s.LongestMessage, msgWidth-4)
	if len(wrapped) > 300 {
		wrapped = wrapped[:300] + "..."
	}
	animated := TypewriterAnimation(wrapped, p)

	msgBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorYellow).
		Foreground(ColorYellow).
		Padding(1, 2).
		Width(msgWidth).
		Render(animated)

	charCount := bigText(fmt.Sprintf("%d", s.LongestMessageLen), ColorYellow) + mutedText(" characters")
	repo := dimText("in ") + lipgloss.NewStyle().Foreground(ColorWhite).Render(s.LongestMessageRepo)
	quote := lipgloss.NewStyle().Foreground(ColorDim).Italic(true).Render("that's not a commit, that's a blog post.")
	footnote := dimText("based on your last 30 days")

	inner := strings.Join([]string{
		"",
		headingStyled,
		sub,
		"",
		msgBox,
		"",
		charCount,
		repo,
		"",
		quote,
		footnote,
		"",
	}, "\n")

	return panel(inner, ColorYellow, width)
}

func renderPersonality(s github.Stats, anim AnimState, width int) string {
	p := anim.Progress()

	label := dimText("you are")

	archetype := TypewriterAnimation(s.Archetype, p)
	archetypeGrad := GradientText(strings.ToUpper(archetype), ColorCyan, ColorPink)
	archetypeStyled := lipgloss.NewStyle().Bold(true).Render(archetypeGrad)

	// Trait pills with colored backgrounds
	pillColors := [3]struct{ fg, bg lipgloss.Color }{
		{lipgloss.Color("#1a1a2e"), ColorPurple},
		{lipgloss.Color("#1a1a2e"), ColorPink},
		{lipgloss.Color("#1a1a2e"), ColorCyan},
	}
	var pills []string
	for i, trait := range s.Traits {
		t := TypewriterAnimation(trait, p)
		if t == "" {
			continue
		}
		name := strings.TrimPrefix(t, "The ")
		pills = append(pills, pill(strings.ToLower(name), pillColors[i].fg, pillColors[i].bg))
	}
	pillRow := strings.Join(pills, "  ")

	divider := accentLine(40, ColorDim)

	outro := mutedText("Your " + s.YearLabel + ", Unwrapped.")
	gifHint := lipgloss.NewStyle().Foreground(ColorCyan).Render("press ") +
		lipgloss.NewStyle().Foreground(ColorCyan).Bold(true).Render("g") +
		lipgloss.NewStyle().Foreground(ColorCyan).Render(" to export as GIF")

	inner := strings.Join([]string{
		"",
		"",
		label,
		"",
		archetypeStyled,
		"",
		pillRow,
		"",
		"",
		divider,
		"",
		outro,
		"",
		gifHint,
		"",
	}, "\n")

	return panel(inner, ColorCyan, width)
}
