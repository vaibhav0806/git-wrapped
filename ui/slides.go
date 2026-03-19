package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/vaibhav/gh-wrapped/github"
)

type SlideID int

const (
	SlideTitle SlideID = iota
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

func panel(content string, accent lipgloss.Color, width int) string {
	boxWidth := width - 4
	if boxWidth > 80 {
		boxWidth = 80
	}
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accent).
		Background(ColorPanel).
		Padding(1, 4).
		Width(boxWidth).
		Align(lipgloss.Center).
		Render(content)
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(box)
}

func heading(text string, color lipgloss.Color) string {
	return lipgloss.NewStyle().Foreground(color).Bold(true).Render(text)
}

func bold(text string, color lipgloss.Color) string {
	return lipgloss.NewStyle().Foreground(color).Bold(true).Render(text)
}

func dim(text string) string {
	return lipgloss.NewStyle().Foreground(ColorDim).Render(text)
}

func muted(text string) string {
	return lipgloss.NewStyle().Foreground(ColorMuted).Render(text)
}

func italic(text string) string {
	return lipgloss.NewStyle().Foreground(ColorDim).Italic(true).Render(text)
}

func divider(w int) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#21262d")).Render(strings.Repeat("─", w))
}

func pill(text string, fg, bg lipgloss.Color) string {
	return lipgloss.NewStyle().
		Foreground(fg).
		Background(bg).
		Bold(true).
		Padding(0, 2).
		Render(text)
}

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
// Slide 1: Title
// ---------------------------------------------------------------------------

var asciiDev = []string{
	"██████╗ ███████╗██╗   ██╗",
	"██╔══██╗██╔════╝██║   ██║",
	"██║  ██║█████╗  ██║   ██║",
	"██║  ██║██╔══╝  ╚██╗ ██╔╝",
	"██████╔╝███████╗ ╚████╔╝ ",
	"╚═════╝ ╚══════╝  ╚═══╝  ",
}

var asciiWrapped = []string{
	"██╗    ██╗██████╗  █████╗ ██████╗ ██████╗ ███████╗██████╗ ",
	"██║    ██║██╔══██╗██╔══██╗██╔══██╗██╔══██╗██╔════╝██╔══██╗",
	"██║ █╗ ██║██████╔╝███████║██████╔╝██████╔╝█████╗  ██║  ██║",
	"██║███╗██║██╔══██╗██╔══██║██╔═══╝ ██╔═══╝ ██╔══╝  ██║  ██║",
	"╚███╔███╔╝██║  ██║██║  ██║██║     ██║     ███████╗██████╔╝",
	" ╚══╝╚══╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝     ╚═╝     ╚══════╝╚═════╝ ",
}

func renderTitle(s github.Stats, anim AnimState, width int) string {
	p := anim.Progress()

	var devLines []string
	for _, line := range asciiDev {
		runes := []rune(line)
		chars := int(float64(len(runes)) * p)
		if chars > len(runes) {
			chars = len(runes)
		}
		devLines = append(devLines, bold(string(runes[:chars]), ColorCyan))
	}

	var wrappedLines []string
	for _, line := range asciiWrapped {
		runes := []rune(line)
		chars := int(float64(len(runes)) * p)
		if chars > len(runes) {
			chars = len(runes)
		}
		wrappedLines = append(wrappedLines, bold(string(runes[:chars]), ColorPurple))
	}

	handle := TypewriterAnimation("@"+s.Username, p)

	inner := strings.Join([]string{
		"",
		strings.Join(devLines, "\n"),
		strings.Join(wrappedLines, "\n"),
		"",
		bold(s.YearLabel, ColorPink),
		"",
		muted(handle),
		dim("your year in code"),
		"",
	}, "\n")

	return panel(inner, ColorPurple, width)
}

// ---------------------------------------------------------------------------
// Slide 2: Year in Numbers
// ---------------------------------------------------------------------------

func renderNumbers(s github.Stats, anim AnimState, width int) string {
	p := anim.Progress()

	contribs := CounterAnimation(s.TotalContributions, p)
	repos := CounterAnimation(s.TotalRepos, p)
	stars := CounterAnimation(s.TotalStars, p)

	h := heading("YOUR YEAR IN NUMBERS", ColorCyan)

	// Each stat: big number on top, label below, inside a mini box
	statStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#21262d")).
		Padding(1, 2).
		Width(20).
		Align(lipgloss.Center)

	stat := func(val string, label string, color lipgloss.Color) string {
		return statStyle.Render(
			bold(val, color) + "\n" + dim(label),
		)
	}

	row := lipgloss.JoinHorizontal(lipgloss.Center,
		stat(fmt.Sprintf("%d", contribs), "contributions", ColorCyan),
		"  ",
		stat(fmt.Sprintf("%d", repos), "repositories", ColorPurple),
		"  ",
		stat(fmt.Sprintf("%d", stars), "stars earned", ColorYellow),
	)

	inner := strings.Join([]string{
		"",
		h,
		"",
		divider(60),
		"",
		row,
		"",
		divider(60),
		"",
	}, "\n")

	return panel(inner, ColorCyan, width)
}

// ---------------------------------------------------------------------------
// Slide 3: Contribution Heatmap
// ---------------------------------------------------------------------------

func renderHeatmap(s github.Stats, anim AnimState, width int) string {
	p := anim.Progress()

	h := heading("CONTRIBUTION HEATMAP", ColorGreen)

	totalCells := len(s.Calendar)
	revealed := HeatmapAnimation(totalCells, p)

	// GitHub-style green shades
	shades := [5]lipgloss.Color{
		lipgloss.Color("#161b22"), // 0: empty
		lipgloss.Color("#0e4429"), // 1: low
		lipgloss.Color("#006d32"), // 2: med
		lipgloss.Color("#26a641"), // 3: high
		lipgloss.Color("#39d353"), // 4: max
	}

	cols := 53
	boxWidth := width - 12
	if boxWidth > 76 {
		boxWidth = 76
	}
	maxCols := (boxWidth - 4) / 2
	if cols > maxCols {
		cols = maxCols
	}

	dayLabels := [7]string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}

	var gridLines []string
	for row := 0; row < 7; row++ {
		label := dim(dayLabels[row]) + " "
		if row%2 == 0 {
			label = dim(dayLabels[row]) + " "
		} else {
			label = "    " // only show every other day label
		}

		var rowSb strings.Builder
		for col := 0; col < cols; col++ {
			idx := row*53 + col
			if idx >= totalCells {
				rowSb.WriteString("  ")
				continue
			}
			if idx >= revealed {
				rowSb.WriteString(lipgloss.NewStyle().Foreground(shades[0]).Render("██"))
				continue
			}
			day := s.Calendar[idx]
			lvl := day.Level
			if lvl == 0 && day.Count > 0 {
				switch {
				case day.Count >= 20:
					lvl = 4
				case day.Count >= 10:
					lvl = 3
				case day.Count >= 5:
					lvl = 2
				default:
					lvl = 1
				}
			}
			if lvl > 4 {
				lvl = 4
			}
			rowSb.WriteString(lipgloss.NewStyle().Foreground(shades[lvl]).Render("██"))
		}
		gridLines = append(gridLines, label+rowSb.String())
	}

	grid := strings.Join(gridLines, "\n")

	// Legend
	legend := dim("less ") +
		lipgloss.NewStyle().Foreground(shades[0]).Render("██") + " " +
		lipgloss.NewStyle().Foreground(shades[1]).Render("██") + " " +
		lipgloss.NewStyle().Foreground(shades[2]).Render("██") + " " +
		lipgloss.NewStyle().Foreground(shades[3]).Render("██") + " " +
		lipgloss.NewStyle().Foreground(shades[4]).Render("██") +
		dim(" more")

	streak := ""
	if s.LongestStreak > 0 {
		streak = bold(fmt.Sprintf("%d", s.LongestStreak), ColorGreen) +
			muted(" day streak  ") +
			dim(s.StreakStart.Format("Jan 2")+" → "+s.StreakEnd.Format("Jan 2"))
	}

	inner := strings.Join([]string{
		"",
		h,
		"",
		grid,
		"",
		legend,
		"",
		divider(60),
		"",
		streak,
		"",
	}, "\n")

	return panel(inner, ColorGreen, width)
}

// ---------------------------------------------------------------------------
// Slide 4: Most Chaotic Day
// ---------------------------------------------------------------------------

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

	h := heading("YOUR MOST CHAOTIC DAY", ColorRed)

	date := TypewriterAnimation(s.BusiestDate.Format("January 2, 2006"), p)

	count := CounterAnimation(s.BusiestCount, p)

	// Big dramatic ASCII art number
	bigNum := RenderBigNumber(fmt.Sprintf("%d", count), ColorRed)
	contribLabel := muted("contributions in a single day")

	inner := strings.Join([]string{
		"",
		h,
		"",
		bold(date, ColorWhite),
		"",
		bigNum,
		contribLabel,
		"",
		italic("« " + quote + " »"),
		"",
	}, "\n")

	return panel(inner, ColorRed, width)
}

// ---------------------------------------------------------------------------
// Slide 5: When You Code
// ---------------------------------------------------------------------------

func renderClock(s github.Stats, anim AnimState, width int) string {
	p := anim.Progress()

	h := heading("WHEN YOU CODE", ColorPurple)

	type timeSlot struct {
		label string
		icon  string
		color lipgloss.Color
	}
	slots := [4]timeSlot{
		{"Morning    06─12", "◑", ColorYellow},
		{"Afternoon  12─18", "◉", ColorCyan},
		{"Evening    18─00", "◗", ColorPurple},
		{"Night      00─06", "○", ColorPink},
	}

	maxVal := 1
	total := 0
	for _, v := range s.TimeBlocks {
		total += v
		if v > maxVal {
			maxVal = v
		}
	}

	barWidth := 25

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

		icon := lipgloss.NewStyle().Foreground(slots[i].color).Render(slots[i].icon)
		label := lipgloss.NewStyle().Foreground(slots[i].color).Width(20).Render(slots[i].label)
		bar := RenderBar(filled, slots[i].color) + RenderBarEmpty(empty)
		pctStr := dim(fmt.Sprintf("  %4.0f%%", pct))

		rows = append(rows, " "+icon+" "+label+" "+bar+pctStr)
	}

	// Verdict pill
	verdictPill := pill(" "+s.TimeLabel+" ", lipgloss.Color("#0d1117"), ColorPurple)

	footnote := dim("based on your last 30 days")

	inner := strings.Join([]string{
		"",
		h,
		"",
		divider(60),
		"",
	}, "\n")
	for _, r := range rows {
		inner += r + "\n"
	}
	inner += strings.Join([]string{
		"",
		divider(60),
		"",
		verdictPill,
		"",
		footnote,
		"",
	}, "\n")

	return panel(inner, ColorPurple, width)
}

// ---------------------------------------------------------------------------
// Slide 6: Top Languages
// ---------------------------------------------------------------------------

func renderLanguages(s github.Stats, anim AnimState, width int) string {
	p := anim.Progress()

	h := heading("YOUR TOP LANGUAGES", ColorYellow)

	barWidth := 25
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

		rank := dim(fmt.Sprintf("#%d ", i+1))
		name := lipgloss.NewStyle().Foreground(color).Bold(true).Width(12).Render(lang.Name)
		bar := RenderBar(filled, color) + RenderBarEmpty(empty)
		pctStr := dim(fmt.Sprintf("  %5.1f%%", lang.Percent))

		rows = append(rows, " "+rank+name+" "+bar+pctStr)
	}

	// Subtitle
	subtitle := ""
	if len(s.Languages) >= 4 && s.Languages[0].Percent < 40 {
		subtitle = italic("polyglot energy")
	} else if len(s.Languages) > 0 && s.Languages[0].Percent > 70 {
		subtitle = italic(s.Languages[0].Name + " loyalist")
	}

	inner := strings.Join([]string{
		"",
		h,
		"",
		divider(60),
		"",
	}, "\n")
	for _, r := range rows {
		inner += r + "\n"
	}
	inner += strings.Join([]string{
		"",
		divider(60),
		"",
		subtitle,
		"",
	}, "\n")

	return panel(inner, ColorYellow, width)
}

// ---------------------------------------------------------------------------
// Slide 7: Villain Arc
// ---------------------------------------------------------------------------

func renderVillain(s github.Stats, anim AnimState, width int) string {
	p := anim.Progress()

	h := heading("VILLAIN ARC", ColorCrimson)

	count := CounterAnimation(s.VillainCommits, p)
	repo := TypewriterAnimation(s.VillainRepo, p)

	// Big dramatic commit count with background
	countPill := lipgloss.NewStyle().
		Background(ColorCrimson).
		Foreground(lipgloss.Color("#0d1117")).
		Bold(true).
		Padding(0, 3).
		Render(fmt.Sprintf(" %d commits ", count))

	repoLine := dim("to ") + bold(repo, ColorWhite)

	inner := strings.Join([]string{
		"",
		h,
		dim("your most pushed repo"),
		"",
		"",
		countPill,
		"",
		repoLine,
		"",
		"",
		italic("« obsessed much? »"),
		"",
		dim("based on your last 30 days"),
		"",
	}, "\n")

	return panel(inner, ColorCrimson, width)
}

// ---------------------------------------------------------------------------
// Slide 8: Weekend Warrior
// ---------------------------------------------------------------------------

func renderWeekend(s github.Stats, anim AnimState, width int) string {
	p := anim.Progress()

	h := heading("WEEKEND WARRIOR", ColorPink)

	animPct := s.WeekendPercent * p

	// Big percentage with background
	pctPill := lipgloss.NewStyle().
		Background(ColorPink).
		Foreground(lipgloss.Color("#0d1117")).
		Bold(true).
		Padding(0, 3).
		Render(fmt.Sprintf(" %.0f%% ", animPct))

	pctLine := muted("of your commits land on weekends")

	barWidth := 50
	weekendW := int(float64(barWidth) * animPct / 100.0)
	weekdayW := barWidth - weekendW
	bar := RenderBarEmpty(weekdayW) + RenderBar(weekendW, ColorPink)

	labelsLine := dim("weekdays") +
		strings.Repeat(" ", barWidth-16) +
		lipgloss.NewStyle().Foreground(ColorPink).Render("weekends")

	var verdict string
	switch {
	case s.WeekendPercent >= 50:
		verdict = "you live for the weekend."
	case s.WeekendPercent >= 25:
		verdict = "work hard, push harder."
	default:
		verdict = "strictly business."
	}

	inner := strings.Join([]string{
		"",
		h,
		"",
		"",
		pctPill,
		pctLine,
		"",
		bar,
		labelsLine,
		"",
		divider(50),
		"",
		italic(verdict),
		"",
	}, "\n")

	return panel(inner, ColorPink, width)
}

// ---------------------------------------------------------------------------
// Slide 9: The Novel
// ---------------------------------------------------------------------------

func renderNovel(s github.Stats, anim AnimState, width int) string {
	p := anim.Progress()

	h := heading("THE NOVEL", ColorYellow)
	sub := dim("your longest commit message")

	msgWidth := 58
	wrapped := wordWrap(s.LongestMessage, msgWidth-6)
	if len(wrapped) > 300 {
		wrapped = wrapped[:300] + "..."
	}
	animated := TypewriterAnimation(wrapped, p)

	msgBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#21262d")).
		Foreground(ColorWhite).
		Padding(1, 2).
		Width(msgWidth).
		Render(animated)

	// Stats line
	charPill := pill(fmt.Sprintf(" %d chars ", s.LongestMessageLen), lipgloss.Color("#0d1117"), ColorYellow)
	repoLine := dim("in ") + bold(s.LongestMessageRepo, ColorWhite)

	inner := strings.Join([]string{
		"",
		h,
		sub,
		"",
		msgBox,
		"",
		charPill + "  " + repoLine,
		"",
		italic("that's not a commit, that's a blog post."),
		dim("based on your last 30 days"),
		"",
	}, "\n")

	return panel(inner, ColorYellow, width)
}

// ---------------------------------------------------------------------------
// Slide 10: Developer Personality (Finale)
// ---------------------------------------------------------------------------

func renderPersonality(s github.Stats, anim AnimState, width int) string {
	p := anim.Progress()

	// Small "you are" label
	label := dim("you are")

	// BIG archetype name with background highlight
	archetype := TypewriterAnimation(s.Archetype, p)
	archetypeName := strings.ToUpper(archetype)
	archetypeStyled := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#0d1117")).
		Background(ColorCyan).
		Bold(true).
		Padding(0, 3).
		Render(" " + archetypeName + " ")

	// Trait pills with colored backgrounds
	pillDefs := [3]struct{ fg, bg lipgloss.Color }{
		{lipgloss.Color("#0d1117"), ColorPurple},
		{lipgloss.Color("#0d1117"), ColorPink},
		{lipgloss.Color("#0d1117"), ColorCyan},
	}
	var pills []string
	for i, trait := range s.Traits {
		t := TypewriterAnimation(trait, p)
		if t == "" {
			continue
		}
		name := strings.TrimPrefix(t, "The ")
		pills = append(pills, pill(" "+strings.ToLower(name)+" ", pillDefs[i].fg, pillDefs[i].bg))
	}
	pillRow := strings.Join(pills, "  ")

	outro := muted("Your " + s.YearLabel + ", Unwrapped.")

	gifBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#21262d")).
		Padding(0, 2).
		Render(
			dim("press ") + bold("g", ColorCyan) + dim(" to export as GIF"),
		)

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
		divider(40),
		"",
		outro,
		"",
		gifBox,
		"",
	}, "\n")

	return panel(inner, ColorCyan, width)
}
