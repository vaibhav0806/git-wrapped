// personality/personality.go
package personality

import (
	"sort"

	"github.com/vaibhav/gh-wrapped/github"
)

type Result struct {
	Archetype string
	Traits    [3]string
}

// Descriptions maps archetype names to short human-readable explanations.
var Descriptions = map[string]string{
	"The Nightcrawler": "you come alive after dark — most of your code is written when others sleep",
	"The Obsessed":     "one repo owns your soul — you can't stop pushing to it",
	"The Novelist":     "your commit messages read like short stories",
	"The Polyglot":     "you speak many languages fluently — no loyalty, just vibes",
	"The Specialist":   "one language to rule them all — deep expertise, laser focus",
	"The Machine":      "relentless consistency — long streaks, no days off",
	"The Weekender":    "weekends aren't for rest — they're for shipping",
	"The Sprinter":     "calm, calm, calm, then BOOM — explosive bursts of activity",
}

type archetype struct {
	Name  string
	Score float64
}

func Compute(s github.Stats) Result {
	archetypes := scoreAll(s)
	sort.SliceStable(archetypes, func(i, j int) bool {
		return archetypes[i].Score > archetypes[j].Score
	})

	result := Result{Archetype: archetypes[0].Name}
	for i := 0; i < 3 && i < len(archetypes); i++ {
		result.Traits[i] = archetypes[i].Name
	}
	return result
}

func scoreAll(s github.Stats) []archetype {
	var a []archetype

	// Events-based
	a = append(a, archetype{"The Nightcrawler", scoreNightcrawler(s)})
	a = append(a, archetype{"The Obsessed", scoreObsessed(s)})
	a = append(a, archetype{"The Novelist", scoreNovelist(s)})

	// Repos-based
	a = append(a, archetype{"The Polyglot", scorePolyglot(s)})
	a = append(a, archetype{"The Specialist", scoreSpecialist(s)})

	// Calendar-based (0 in degraded mode)
	if s.HasCalendar {
		a = append(a, archetype{"The Machine", scoreMachine(s)})
		a = append(a, archetype{"The Weekender", scoreWeekender(s)})
		a = append(a, archetype{"The Sprinter", scoreSprinter(s)})
	} else {
		a = append(a, archetype{"The Machine", 0})
		a = append(a, archetype{"The Weekender", 0})
		a = append(a, archetype{"The Sprinter", 0})
	}
	return a
}

func scoreNightcrawler(s github.Stats) float64 {
	total := s.TimeBlocks[0] + s.TimeBlocks[1] + s.TimeBlocks[2] + s.TimeBlocks[3]
	if total == 0 {
		return 0
	}
	evening := float64(s.TimeBlocks[2]+s.TimeBlocks[3]) / float64(total)
	if evening > 0.5 {
		return clamp(evening)
	}
	return 0
}

func scoreObsessed(s github.Stats) float64 {
	if s.VillainCommits == 0 {
		return 0
	}
	totalPush := s.TimeBlocks[0] + s.TimeBlocks[1] + s.TimeBlocks[2] + s.TimeBlocks[3]
	if totalPush == 0 {
		return 0
	}
	ratio := float64(s.VillainCommits) / float64(totalPush)
	if ratio > 0.6 {
		return clamp(ratio)
	}
	return 0
}

func scoreNovelist(s github.Stats) float64 {
	if s.LongestMessageLen > 500 {
		return clamp(float64(s.LongestMessageLen) / 1000)
	}
	return 0
}

func scorePolyglot(s github.Stats) float64 {
	if len(s.Languages) < 4 {
		return 0
	}
	maxPct := 0.0
	for _, l := range s.Languages {
		if l.Percent > maxPct {
			maxPct = l.Percent
		}
	}
	if maxPct < 40 {
		return clamp(float64(len(s.Languages)) / 10)
	}
	return 0
}

func scoreSpecialist(s github.Stats) float64 {
	if len(s.Languages) == 0 {
		return 0
	}
	if s.Languages[0].Percent > 70 {
		return clamp(s.Languages[0].Percent / 100)
	}
	return 0
}

func scoreMachine(s github.Stats) float64 {
	if s.LongestStreak >= 14 {
		return clamp(float64(s.LongestStreak) / 30)
	}
	return 0
}

func scoreWeekender(s github.Stats) float64 {
	if s.WeekendPercent > 30 {
		return clamp(s.WeekendPercent / 100)
	}
	return 0
}

func scoreSprinter(s github.Stats) float64 {
	if s.BusiestCount > 0 && s.TotalContributions > 0 {
		avg := float64(s.TotalContributions) / 365.0
		if avg > 0 {
			ratio := float64(s.BusiestCount) / avg
			if ratio > 5 {
				return clamp(ratio / 20)
			}
		}
	}
	return 0
}

func clamp(v float64) float64 {
	if v > 1 {
		return 1
	}
	if v < 0 {
		return 0
	}
	return v
}
