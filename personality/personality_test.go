// personality/personality_test.go
package personality

import (
	"testing"

	"github.com/vaibhav/gh-wrapped/github"
)

func TestScoreNightcrawler(t *testing.T) {
	s := github.Stats{
		TimeBlocks: [4]int{1, 1, 8, 2}, // 83% evening
	}
	result := Compute(s)
	if result.Archetype != "The Nightcrawler" {
		t.Errorf("got %q, want The Nightcrawler", result.Archetype)
	}
}

func TestScoreSpecialist(t *testing.T) {
	s := github.Stats{
		Languages: []github.LangStat{
			{Name: "Go", Percent: 85},
			{Name: "Shell", Percent: 15},
		},
	}
	result := Compute(s)
	if result.Archetype != "The Specialist" {
		t.Errorf("got %q, want The Specialist", result.Archetype)
	}
}

func TestScorePolyglot(t *testing.T) {
	s := github.Stats{
		Languages: []github.LangStat{
			{Name: "Go", Percent: 25},
			{Name: "Python", Percent: 25},
			{Name: "JS", Percent: 25},
			{Name: "Rust", Percent: 25},
		},
	}
	result := Compute(s)
	if result.Archetype != "The Polyglot" {
		t.Errorf("got %q, want The Polyglot", result.Archetype)
	}
}

func TestDegradedMode(t *testing.T) {
	s := github.Stats{
		HasCalendar: false,
		TimeBlocks:  [4]int{0, 0, 5, 5}, // 100% evening
	}
	result := Compute(s)
	if result.Archetype != "The Nightcrawler" {
		t.Errorf("got %q, want The Nightcrawler in degraded mode", result.Archetype)
	}
}

func TestTraitsAreTopThree(t *testing.T) {
	s := github.Stats{
		HasCalendar:    true,
		TimeBlocks:     [4]int{0, 0, 8, 2},
		WeekendPercent: 40,
		LongestStreak:  20,
		Languages: []github.LangStat{
			{Name: "Go", Percent: 80},
		},
	}
	result := Compute(s)
	if len(result.Traits) != 3 {
		t.Errorf("got %d traits, want 3", len(result.Traits))
	}
}
