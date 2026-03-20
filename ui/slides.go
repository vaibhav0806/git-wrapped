package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/vaibhav0806/git-wrapped/github"
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

var asciiGit = []string{
	" ██████╗ ██╗████████╗",
	"██╔════╝ ██║╚══██╔══╝",
	"██║  ███╗██║   ██║   ",
	"██║   ██║██║   ██║   ",
	"╚██████╔╝██║   ██║   ",
	" ╚═════╝ ╚═╝   ╚═╝   ",
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
	for _, line := range asciiGit {
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

	// Hero stat: contributions in big ASCII digits
	bigContribs := RenderBigNumber(fmt.Sprintf("%d", contribs), ColorCyan)
	contribLabel := dim("contributions")

	// Secondary stats on one line
	secondaryLine := bold(fmt.Sprintf("%d", repos), ColorPurple) + dim(" repos") +
		muted("    ·    ") +
		bold(fmt.Sprintf("%d", stars), ColorYellow) + dim(" stars earned")

	inner := strings.Join([]string{
		"",
		h,
		"",
		divider(50),
		"",
		bigContribs,
		contribLabel,
		"",
		divider(50),
		"",
		secondaryLine,
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

	// Total contributions
	totalCount := CounterAnimation(s.TotalContributions, p)
	totalLine := bold(fmt.Sprintf("%d", totalCount), ColorGreen) + dim(" contributions in the last year")

	totalCells := len(s.Calendar)
	revealed := HeatmapAnimation(totalCells, p)

	// GitHub-style green shades
	shades := [5]lipgloss.Color{
		lipgloss.Color("#21262d"), // 0: empty
		lipgloss.Color("#0e4429"), // 1: low
		lipgloss.Color("#006d32"), // 2: med
		lipgloss.Color("#26a641"), // 3: high
		lipgloss.Color("#39d353"), // 4: max
	}

	cols := 53

	// Render grid using background-colored spaces (consistent across all terminals).
	// Block characters like █ render at inconsistent widths on some terminals.
	cell := "  " // 2 spaces per cell
	var gridLines []string
	for row := 0; row < 7; row++ {
		var rowSb strings.Builder
		for col := 0; col < cols; col++ {
			idx := row*53 + col
			if idx >= totalCells || idx >= revealed {
				rowSb.WriteString(lipgloss.NewStyle().Background(shades[0]).Render(cell))
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
			rowSb.WriteString(lipgloss.NewStyle().Background(shades[lvl]).Render(cell))
		}
		gridLines = append(gridLines, rowSb.String())
	}

	grid := strings.Join(gridLines, "\n")

	// Legend
	legend := dim("less ") +
		lipgloss.NewStyle().Background(shades[0]).Render(cell) +
		lipgloss.NewStyle().Background(shades[1]).Render(cell) +
		lipgloss.NewStyle().Background(shades[2]).Render(cell) +
		lipgloss.NewStyle().Background(shades[3]).Render(cell) +
		lipgloss.NewStyle().Background(shades[4]).Render(cell) +
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
		totalLine,
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

	// Wider panel for heatmap to fit 53 * 2 = 106 char grid
	boxWidth := width - 4
	if boxWidth > 120 {
		boxWidth = 120
	}
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorGreen).
		Background(ColorPanel).
		Padding(1, 4).
		Width(boxWidth).
		Align(lipgloss.Center).
		Render(inner)
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(box)
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
		langName := lang.Name
		if len(langName) > 11 {
			langName = langName[:10] + "…"
		}
		name := lipgloss.NewStyle().Foreground(color).Bold(true).Width(12).Render(langName)
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

	// Compute actual counts
	weekdayCount, weekendCount := 0, 0
	for _, d := range s.Calendar {
		if d.Count > 0 {
			wd := d.Date.Weekday()
			if wd == 0 || wd == 6 {
				weekendCount += d.Count
			} else {
				weekdayCount += d.Count
			}
		}
	}

	animWeekday := CounterAnimation(weekdayCount, p)
	animWeekend := CounterAnimation(weekendCount, p)
	weekdayPct := 100 - s.WeekendPercent

	// Two side-by-side stat boxes — same size, different emphasis
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 3).
		Width(28).
		Height(7).
		Align(lipgloss.Center)

	weekdayBox := boxStyle.
		BorderForeground(lipgloss.Color("#30363d")).
		Render(
			muted(fmt.Sprintf("%d", animWeekday))+"\n"+
				dim("commits")+"\n"+
				"\n"+
				muted(fmt.Sprintf("%.0f%%", weekdayPct*p))+"\n"+
				dim("Mon – Fri"),
		)

	weekendBox := boxStyle.
		BorderForeground(ColorPink).
		Render(
			bold(fmt.Sprintf("%d", animWeekend), ColorPink)+"\n"+
				bold("commits", ColorPink)+"\n"+
				"\n"+
				bold(fmt.Sprintf("%.0f%%", s.WeekendPercent*p), ColorPink)+"\n"+
				muted("Sat – Sun"),
		)

	boxes := lipgloss.JoinHorizontal(lipgloss.Center, weekdayBox, "  ", weekendBox)

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
		boxes,
		"",
		divider(60),
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

// All archetype names for the lottery effect.
var allArchetypes = []string{
	"The Nightcrawler",
	"The Obsessed",
	"The Novelist",
	"The Polyglot",
	"The Specialist",
	"The Machine",
	"The Weekender",
	"The Sprinter",
}

func renderPersonality(s github.Stats, anim AnimState, width int) string {
	// This slide uses a longer animation (3x normal) for the lottery effect.
	// We remap p from the normal 0-1 range across a longer timeline.
	// The anim runs 0-1 over AnimDurationMs, but we use raw tick count
	// to extend the lottery phase beyond the normal animation window.
	// Tick count gives us ~30 ticks per second (50ms interval).
	// Lottery: 60 ticks (2s), settle: 10 ticks, reveal: staggered after.
	tick := anim.Tick
	lotteryEnd := 40   // 2 seconds of cycling
	settleAt := lotteryEnd + 3
	descAt := settleAt + 10
	pillsAt := descAt + 8
	outroAt := pillsAt + 6

	label := dim("you are")

	var displayName string
	var nameStyle lipgloss.Style
	landed := tick >= lotteryEnd

	if !landed {
		// Cycling phase — slow down as we approach lotteryEnd
		progress := float64(tick) / float64(lotteryEnd)
		cycleSpeed := 1 + int((1.0-progress)*6) // fast → slow
		if cycleSpeed < 1 {
			cycleSpeed = 1
		}
		idx := (tick / cycleSpeed) % len(allArchetypes)
		displayName = strings.ToUpper(allArchetypes[idx])
		nameStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Bold(true)
	} else {
		displayName = strings.ToUpper(s.Archetype)
		nameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#0d1117")).
			Background(ColorCyan).
			Bold(true).
			Padding(0, 3)
	}

	archetypeStyled := nameStyle.Render(" " + displayName + " ")

	// Description
	archetypeDesc := ""
	if tick >= descAt {
		archetypeDesc = italic(s.ArchetypeDescription)
	}

	// Trait pills
	pillDefs := [3]struct{ fg, bg lipgloss.Color }{
		{lipgloss.Color("#0d1117"), ColorPurple},
		{lipgloss.Color("#0d1117"), ColorPink},
		{lipgloss.Color("#0d1117"), ColorCyan},
	}
	pillRow := ""
	if tick >= pillsAt {
		var pills []string
		for i, trait := range s.Traits {
			if trait == "" {
				continue
			}
			name := strings.TrimPrefix(trait, "The ")
			pills = append(pills, pill(" "+strings.ToLower(name)+" ", pillDefs[i].fg, pillDefs[i].bg))
		}
		pillRow = strings.Join(pills, "  ")
	}

	// Outro + GIF hint
	outro := ""
	gifBox := ""
	if tick >= outroAt {
		outro = muted("Your " + s.YearLabel + ", Unwrapped.")
		gifBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#21262d")).
			Padding(0, 2).
			Render(
				dim("press ") + bold("g", ColorCyan) + dim(" to export as GIF"),
			)
	}

	var parts []string
	parts = append(parts, "", "", label, "", archetypeStyled)
	if archetypeDesc != "" {
		parts = append(parts, archetypeDesc)
	}
	if pillRow != "" {
		parts = append(parts, "", pillRow)
	}
	if outro != "" {
		parts = append(parts, "", divider(50), "", outro, "", gifBox)
	}
	parts = append(parts, "", "")

	inner := strings.Join(parts, "\n")
	return panel(inner, ColorCyan, width)
}
