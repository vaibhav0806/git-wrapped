// github/stats_test.go
package github

import (
	"testing"
	"time"
)

func makeDate(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

func makeTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

func TestComputeStats_YearInNumbers(t *testing.T) {
	calendar := []ContributionDay{
		{Date: makeDate("2025-01-01"), Level: 2, Count: 5},
		{Date: makeDate("2025-01-02"), Level: 3, Count: 12},
		{Date: makeDate("2025-12-31"), Level: 1, Count: 2},
	}
	repos := []Repo{
		{Name: "a", Language: "Go", StargazersCount: 10},
		{Name: "b", Language: "Go", StargazersCount: 5},
	}

	stats := ComputeStats(User{Login: "test", Name: "Test"}, nil, repos, calendar)
	if stats.TotalRepos != 2 {
		t.Errorf("TotalRepos: got %d, want 2", stats.TotalRepos)
	}
	if stats.TotalStars != 15 {
		t.Errorf("TotalStars: got %d, want 15", stats.TotalStars)
	}
	if stats.YearLabel != "2025" {
		t.Errorf("YearLabel: got %q, want 2025", stats.YearLabel)
	}
	if stats.TotalContributions != 19 {
		t.Errorf("TotalContributions: got %d, want 19", stats.TotalContributions)
	}
}

func TestComputeStats_LongestStreak(t *testing.T) {
	calendar := []ContributionDay{
		{Date: makeDate("2025-01-01"), Level: 1, Count: 3},
		{Date: makeDate("2025-01-02"), Level: 2, Count: 8},
		{Date: makeDate("2025-01-03"), Level: 1, Count: 2},
		{Date: makeDate("2025-01-04"), Level: 0, Count: 0}, // break
		{Date: makeDate("2025-01-05"), Level: 3, Count: 15},
		{Date: makeDate("2025-01-06"), Level: 1, Count: 4},
	}
	stats := ComputeStats(User{Login: "test"}, nil, nil, calendar)
	if stats.LongestStreak != 3 {
		t.Errorf("LongestStreak: got %d, want 3", stats.LongestStreak)
	}
}

func TestComputeStats_BusiestDay(t *testing.T) {
	calendar := []ContributionDay{
		{Date: makeDate("2025-03-14"), Level: 4, Count: 47},
		{Date: makeDate("2025-03-15"), Level: 1, Count: 3},
		{Date: makeDate("2025-06-01"), Level: 3, Count: 20},
	}
	stats := ComputeStats(User{Login: "test"}, nil, nil, calendar)
	if stats.BusiestDate.Format("2006-01-02") != "2025-03-14" {
		t.Errorf("BusiestDate: got %s, want 2025-03-14", stats.BusiestDate.Format("2006-01-02"))
	}
	if stats.BusiestCount != 47 {
		t.Errorf("BusiestCount: got %d, want 47", stats.BusiestCount)
	}
}

func TestComputeStats_TimeBlocks(t *testing.T) {
	events := []Event{
		{Type: "PushEvent", CreatedAt: makeTime("2025-03-01T08:00:00Z")},  // 06-12
		{Type: "PushEvent", CreatedAt: makeTime("2025-03-01T14:00:00Z")},  // 12-18
		{Type: "PushEvent", CreatedAt: makeTime("2025-03-01T20:00:00Z")},  // 18-00
		{Type: "PushEvent", CreatedAt: makeTime("2025-03-01T20:30:00Z")},  // 18-00
		{Type: "PushEvent", CreatedAt: makeTime("2025-03-01T03:00:00Z")},  // 00-06
	}
	stats := ComputeStats(User{Login: "test"}, events, nil, nil)
	expected := [4]int{1, 1, 2, 1}
	if stats.TimeBlocks != expected {
		t.Errorf("TimeBlocks: got %v, want %v", stats.TimeBlocks, expected)
	}
	if stats.TimeLabel != "Night Owl" {
		t.Errorf("TimeLabel: got %q, want Night Owl", stats.TimeLabel)
	}
}

func TestComputeStats_Languages(t *testing.T) {
	repos := []Repo{
		{Name: "a", Language: "Go"},
		{Name: "b", Language: "Go"},
		{Name: "c", Language: "Python"},
		{Name: "d", Language: ""},
	}
	stats := ComputeStats(User{Login: "test"}, nil, repos, nil)
	if len(stats.Languages) < 2 {
		t.Fatalf("Languages: got %d, want >= 2", len(stats.Languages))
	}
	if stats.Languages[0].Name != "Go" || stats.Languages[0].Count != 2 {
		t.Errorf("Top language: got %+v, want Go:2", stats.Languages[0])
	}
}

func TestComputeStats_VillainArc(t *testing.T) {
	events := []Event{
		{Type: "PushEvent", Repo: EventRepo{Name: "user/api"}, Payload: Payload{Commits: []Commit{{Message: "fix"}}}},
		{Type: "PushEvent", Repo: EventRepo{Name: "user/api"}, Payload: Payload{Commits: []Commit{{Message: "refactor"}}}},
		{Type: "PushEvent", Repo: EventRepo{Name: "user/web"}, Payload: Payload{Commits: []Commit{{Message: "update"}}}},
	}
	stats := ComputeStats(User{Login: "test"}, events, nil, nil)
	if stats.VillainRepo != "user/api" {
		t.Errorf("VillainRepo: got %q, want user/api", stats.VillainRepo)
	}
	if stats.VillainCommits != 2 {
		t.Errorf("VillainCommits: got %d, want 2", stats.VillainCommits)
	}
}

func TestComputeStats_WeekendPercent(t *testing.T) {
	calendar := []ContributionDay{
		{Date: makeDate("2025-01-06"), Level: 1, Count: 5}, // Monday
		{Date: makeDate("2025-01-07"), Level: 1, Count: 5}, // Tuesday
		{Date: makeDate("2025-01-11"), Level: 1, Count: 5}, // Saturday
		{Date: makeDate("2025-01-12"), Level: 1, Count: 5}, // Sunday
	}
	stats := ComputeStats(User{Login: "test"}, nil, nil, calendar)
	if stats.WeekendPercent != 50.0 {
		t.Errorf("WeekendPercent: got %.1f, want 50.0", stats.WeekendPercent)
	}
}

func TestComputeStats_LongestMessage(t *testing.T) {
	events := []Event{
		{Type: "PushEvent", Repo: EventRepo{Name: "user/api"}, Payload: Payload{
			Commits: []Commit{{Message: "short"}},
		}},
		{Type: "PushEvent", Repo: EventRepo{Name: "user/web"}, Payload: Payload{
			Commits: []Commit{{Message: "this is a much longer commit message that goes on and on"}},
		}},
	}
	stats := ComputeStats(User{Login: "test"}, events, nil, nil)
	if stats.LongestMessageRepo != "user/web" {
		t.Errorf("LongestMessageRepo: got %q, want user/web", stats.LongestMessageRepo)
	}
}
