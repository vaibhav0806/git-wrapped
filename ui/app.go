package ui

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/vaibhav0806/git-wrapped/github"
	"github.com/vaibhav0806/git-wrapped/personality"
	"golang.org/x/sync/errgroup"
)

const (
	MinWidth       = 100
	MinHeight      = 30
	AutoAdvanceMs  = 3500
	AnimIntervalMs = 50
	AnimDurationMs = 1500
	TransitionMs   = 300
)

type tickMsg time.Time

type phase int

const (
	phaseLoading phase = iota
	phaseTransitionIn
	phasePresenting
	phaseTransitionOut
)

// DataLoadedMsg is sent when background data fetching completes.
type DataLoadedMsg struct {
	User     github.User
	Events   []github.Event
	Repos    []github.Repo
	Calendar []github.ContributionDay
	Err      error
}

// Model is the Bubble Tea model.
type Model struct {
	// Loading phase
	phase    phase
	spinner  spinner.Model
	username string
	token    string

	// Presenting phase
	stats     github.Stats
	slides    []SlideID
	current   int
	anim      AnimState
	autoPlay  bool
	autoMode  bool
	width     int
	height    int
	tooSmall  bool
	done      bool
	exportGIF bool

	extraTicks     int
	transitionTick int
	transitionMax  int

	// Particles (pre-generated positions)
	particles []particle
}

type particle struct {
	x, y int
	ch   rune
}

// NewModel creates a model that starts in loading phase.
func NewModel(username, token string, auto bool) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorCyan)

	return Model{
		phase:         phaseLoading,
		spinner:       s,
		username:      username,
		token:         token,
		autoPlay:      true,
		autoMode:      auto,
		transitionMax: TransitionMs / AnimIntervalMs,
	}
}

// FetchData returns a tea.Cmd that fetches all GitHub data concurrently.
func FetchData(username, token string) tea.Cmd {
	return func() tea.Msg {
		client := github.NewClient("https://api.github.com", token)

		var (
			user     github.User
			events   []github.Event
			repos    []github.Repo
			calendar []github.ContributionDay
		)

		g := new(errgroup.Group)

		g.Go(func() error {
			var err error
			user, err = client.FetchUser(username)
			return err
		})
		g.Go(func() error {
			var err error
			events, err = client.FetchEvents(username)
			if err != nil {
				events = nil
			}
			return nil
		})
		g.Go(func() error {
			var err error
			repos, err = client.FetchRepos(username)
			if err != nil {
				repos = nil
			}
			return nil
		})
		g.Go(func() error {
			var err error
			calendar, err = client.FetchContributions(username)
			if err != nil {
				calendar = nil
			}
			return nil
		})

		if err := g.Wait(); err != nil {
			return DataLoadedMsg{Err: err}
		}
		return DataLoadedMsg{User: user, Events: events, Repos: repos, Calendar: calendar}
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		FetchData(m.username, m.token),
	)
}

func tick() tea.Cmd {
	return tea.Tick(time.Duration(AnimIntervalMs)*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.tooSmall = m.width < MinWidth || m.height < MinHeight
		m.particles = generateParticles(m.width, m.height, 40)
		return m, nil

	case DataLoadedMsg:
		if msg.Err != nil {
			m.done = true
			return m, tea.Quit
		}

		if len(msg.Events) == 0 && len(msg.Calendar) == 0 && len(msg.Repos) == 0 {
			m.done = true
			return m, tea.Quit
		}

		stats := github.ComputeStats(msg.User, msg.Events, msg.Repos, msg.Calendar)
		p := personality.Compute(stats)
		stats.Archetype = p.Archetype
		stats.ArchetypeDescription = personality.Descriptions[p.Archetype]
		stats.Traits = p.Traits
		for i, t := range p.Traits {
			if t != "" {
				stats.TraitDescriptions[i] = personality.Descriptions[t]
			}
		}

		m.stats = stats
		m.slides = ActiveSlides(stats)
		m.phase = phaseTransitionIn
		m.transitionTick = 0
		m.anim = NewAnimState(AnimDurationMs, AnimIntervalMs)
		return m, tick()

	case spinner.TickMsg:
		if m.phase == phaseLoading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case tea.KeyMsg:
		if m.phase == phaseLoading || m.autoMode {
			if msg.String() == "q" || msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			return m, nil
		}
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.done = true
			return m, tea.Quit
		case "right", "l", " ":
			m.autoPlay = false
			if m.current < len(m.slides)-1 {
				m.phase = phaseTransitionOut
				m.transitionTick = 0
			}
		case "left", "h":
			if m.current > 0 {
				m.autoPlay = false
				m.current--
				m.phase = phaseTransitionIn
				m.transitionTick = 0
				m.anim = NewAnimState(AnimDurationMs, AnimIntervalMs)
			}
		case "a":
			m.autoPlay = true
		case "g":
			if m.current == len(m.slides)-1 {
				m.exportGIF = true
				m.done = true
				return m, tea.Quit
			}
		}
		return m, nil

	case tickMsg:
		if m.done {
			return m, nil
		}

		switch m.phase {
		case phaseTransitionIn:
			m.transitionTick++
			if m.transitionTick >= m.transitionMax {
				m.phase = phasePresenting
			}
			return m, tick()

		case phaseTransitionOut:
			m.transitionTick++
			if m.transitionTick >= m.transitionMax {
				m.current++
				m.phase = phaseTransitionIn
				m.transitionTick = 0
				m.anim = NewAnimState(AnimDurationMs, AnimIntervalMs)
				m.extraTicks = 0
			}
			return m, tick()

		case phasePresenting:
			m.anim.Advance()
			if m.anim.Done && m.autoPlay {
				extraNeeded := (AutoAdvanceMs - AnimDurationMs) / AnimIntervalMs
				m.extraTicks++
				if m.extraTicks >= extraNeeded {
					if m.current == len(m.slides)-1 {
						if m.autoMode {
							m.done = true
							return m, tea.Quit
						}
					} else {
						m.phase = phaseTransitionOut
						m.transitionTick = 0
					}
				}
			}
			return m, tick()
		}

		return m, tick()
	}

	return m, nil
}

func (m Model) View() string {
	if m.tooSmall && m.width > 0 {
		return lipgloss.NewStyle().Foreground(ColorRed).Bold(true).Render(
			fmt.Sprintf("Terminal too small (%dx%d). Need %dx%d.", m.width, m.height, MinWidth, MinHeight),
		)
	}

	var raw string

	switch m.phase {
	case phaseLoading:
		raw = m.viewLoading()
	case phaseTransitionIn, phaseTransitionOut:
		raw = m.viewTransition()
	case phasePresenting:
		raw = m.viewPresenting()
	default:
		raw = ""
	}

	return m.fillScreen(raw)
}

func (m Model) viewLoading() string {
	spinnerStr := m.spinner.View()
	loading := lipgloss.NewStyle().Foreground(ColorMuted).Render("  fetching @" + m.username + "'s year...")

	content := strings.Join([]string{
		"",
		"",
		spinnerStr + loading,
		"",
		"",
	}, "\n")

	boxWidth := 50
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorCyan).
		Background(ColorPanel).
		Padding(2, 4).
		Width(boxWidth).
		Align(lipgloss.Center).
		Render(content)

	// Center in screen
	centered := lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(box)
	lines := strings.Split(centered, "\n")
	padTop := (m.height - len(lines)) / 2
	if padTop < 0 {
		padTop = 0
	}
	return strings.Repeat("\n", padTop) + centered
}

func (m Model) viewTransition() string {
	// Blank frame with just the background — brief pause between slides
	return strings.Repeat("\n", m.height/2)
}

func (m Model) viewPresenting() string {
	slideHeight := m.height - 2
	if slideHeight < 1 {
		slideHeight = 1
	}

	slideID := m.slides[m.current]
	slideContent := RenderSlide(slideID, m.stats, m.anim, m.width, slideHeight)

	// Progress dots
	dots := m.renderProgressDots()

	// Controls hint
	var hint string
	if m.autoMode {
		hint = dim("auto")
	} else {
		hint = dim("← → · a auto · q quit")
	}

	bottomBar := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(m.width/2).Align(lipgloss.Left).Render(" "+dots),
		lipgloss.NewStyle().Width(m.width-m.width/2).Align(lipgloss.Right).Render(hint+" "),
	)

	return slideContent + "\n" + bottomBar
}

func (m Model) renderProgressDots() string {
	var sb strings.Builder
	active := lipgloss.NewStyle().Foreground(ColorCyan)
	inactive := lipgloss.NewStyle().Foreground(lipgloss.Color("#21262d"))

	for i := range m.slides {
		if i > 0 {
			sb.WriteString(" ")
		}
		if i == m.current {
			sb.WriteString(active.Render("●"))
		} else if i < m.current {
			sb.WriteString(active.Render("●"))
		} else {
			sb.WriteString(inactive.Render("●"))
		}
	}
	return sb.String()
}

func (m Model) fillScreen(raw string) string {
	if m.width == 0 || m.height == 0 {
		return raw
	}

	lines := strings.Split(raw, "\n")
	var sb strings.Builder
	bgStyle := lipgloss.NewStyle().Background(ColorBg)

	for i := 0; i < m.height; i++ {
		line := ""
		if i < len(lines) {
			line = lines[i]
		}
		visWidth := lipgloss.Width(line)
		pad := m.width - visWidth
		if pad < 0 {
			pad = 0
		}

		// Scatter particles in the padding
		if pad > 0 && len(m.particles) > 0 {
			sb.WriteString(line)
			padStr := m.renderPaddingWithParticles(i, visWidth, pad)
			sb.WriteString(padStr)
		} else {
			sb.WriteString(line)
			if pad > 0 {
				sb.WriteString(bgStyle.Render(strings.Repeat(" ", pad)))
			}
		}

		if i < m.height-1 {
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}

func (m Model) renderPaddingWithParticles(row, startCol, padWidth int) string {
	bgStyle := lipgloss.NewStyle().Background(ColorBg)
	particleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#1a1f2e")).Background(ColorBg)

	// Build the padding string, inserting particles at matching positions
	padRunes := make([]rune, padWidth)
	for j := range padRunes {
		padRunes[j] = ' '
	}

	for _, p := range m.particles {
		if p.y == row {
			col := p.x - startCol
			if col >= 0 && col < padWidth {
				padRunes[col] = p.ch
			}
		}
	}

	// Render character by character where particles exist
	var sb strings.Builder
	for _, ch := range padRunes {
		if ch == ' ' {
			sb.WriteString(bgStyle.Render(" "))
		} else {
			sb.WriteString(particleStyle.Render(string(ch)))
		}
	}
	return sb.String()
}

func generateParticles(width, height, count int) []particle {
	if width == 0 || height == 0 {
		return nil
	}
	rng := rand.New(rand.NewSource(42)) // deterministic
	chars := []rune{'·', '∗', '✦', '⋅', '·', '·', '⋅'}

	var particles []particle
	for i := 0; i < count; i++ {
		x := rng.Intn(width)
		y := rng.Intn(height)
		ch := chars[rng.Intn(len(chars))]
		particles = append(particles, particle{x: x, y: y, ch: ch})
	}
	return particles
}

// ExportRequested reports whether the user requested a GIF export.
func (m Model) ExportRequested() bool {
	return m.exportGIF
}

// Username returns the username for GIF export.
func (m Model) Username() string {
	return m.username
}
