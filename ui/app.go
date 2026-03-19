package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/vaibhav/gh-wrapped/github"
)

const (
	MinWidth       = 100
	MinHeight      = 30
	AutoAdvanceMs  = 3000
	AnimIntervalMs = 50
	AnimDurationMs = 1500
)

// tickMsg is the message type for animation ticks.
type tickMsg time.Time

// Model is the Bubble Tea model that drives the slide presentation.
type Model struct {
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

	// extraTicks counts ticks after animation completes, used for auto-advance delay.
	extraTicks int
}

// NewModel constructs a Model with the active slide list.
func NewModel(stats github.Stats, auto bool) Model {
	slides := ActiveSlides(stats)
	return Model{
		stats:    stats,
		slides:   slides,
		current:  0,
		anim:     NewAnimState(AnimDurationMs, AnimIntervalMs),
		autoPlay: auto,
		autoMode: auto,
	}
}

// Init returns the first tick command.
func (m Model) Init() tea.Cmd {
	return tick()
}

// tick returns a tea.Cmd that fires a tickMsg after AnimIntervalMs.
func tick() tea.Cmd {
	return tea.Tick(time.Duration(AnimIntervalMs)*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles incoming messages and updates the model state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.tooSmall = m.width < MinWidth || m.height < MinHeight
		return m, nil

	case tea.KeyMsg:
		if m.autoMode {
			// In auto mode, ignore all key input.
			return m, nil
		}
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.done = true
			return m, tea.Quit

		case "right", "l", " ":
			m.autoPlay = false
			m = m.nextSlide()

		case "left", "h":
			m = m.prevSlide()

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

		if !m.anim.Done {
			m.anim.Advance()
		} else if m.autoPlay {
			// Animation is done; count extra ticks until auto-advance delay expires.
			// Total extra ticks needed: (AutoAdvanceMs - AnimDurationMs) / AnimIntervalMs
			extraNeeded := (AutoAdvanceMs - AnimDurationMs) / AnimIntervalMs
			m.extraTicks++
			if m.extraTicks >= extraNeeded {
				// Time to advance.
				if m.current == len(m.slides)-1 {
					if m.autoMode {
						m.done = true
						return m, tea.Quit
					}
				} else {
					m = m.nextSlide()
				}
			}
		}

		return m, tick()
	}

	return m, nil
}

// nextSlide advances to the next slide and resets animation state.
func (m Model) nextSlide() Model {
	if m.current < len(m.slides)-1 {
		m.current++
	}
	m.anim = NewAnimState(AnimDurationMs, AnimIntervalMs)
	m.extraTicks = 0
	return m
}

// prevSlide moves to the previous slide and resets animation state.
func (m Model) prevSlide() Model {
	if m.current > 0 {
		m.current--
	}
	m.anim = NewAnimState(AnimDurationMs, AnimIntervalMs)
	m.extraTicks = 0
	return m
}

// View renders the current slide and bottom bar.
func (m Model) View() string {
	if m.tooSmall {
		warning := fmt.Sprintf(
			"Terminal too small. Need at least %dx%d, got %dx%d.",
			MinWidth, MinHeight, m.width, m.height,
		)
		return lipgloss.NewStyle().
			Foreground(ColorRed).
			Bold(true).
			Render(warning)
	}

	// Reserve 1 line for the bottom bar.
	slideHeight := m.height - 1
	if slideHeight < 1 {
		slideHeight = 1
	}

	slideID := m.slides[m.current]
	slideContent := RenderSlide(slideID, m.stats, m.anim, m.width, slideHeight)

	// Bottom bar: slide counter + controls hint.
	counter := fmt.Sprintf("%d/%d", m.current+1, len(m.slides))
	counterStr := LabelStyle.Render(counter)

	var hint string
	if m.autoMode {
		hint = LabelStyle.Render("auto")
	} else {
		hint = LabelStyle.Render("← → navigate  a auto  q quit")
	}

	barStyle := lipgloss.NewStyle().Width(m.width)
	bottomBar := barStyle.Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(m.width/2).Align(lipgloss.Left).Render(counterStr),
			lipgloss.NewStyle().Width(m.width-m.width/2).Align(lipgloss.Right).Render(hint),
		),
	)

	return slideContent + "\n" + bottomBar
}

// ExportRequested reports whether the user requested a GIF export.
func (m Model) ExportRequested() bool {
	return m.exportGIF
}
