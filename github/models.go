// github/models.go
package github

import "time"

// --- Raw API response types ---

type User struct {
	Login     string `json:"login"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

type Event struct {
	Type      string    `json:"type"`
	Repo      EventRepo `json:"repo"`
	CreatedAt time.Time `json:"created_at"`
	Payload   Payload   `json:"payload"`
}

type EventRepo struct {
	Name string `json:"name"`
}

type Payload struct {
	Commits []Commit `json:"commits,omitempty"`
}

type Commit struct {
	Message string `json:"message"`
	SHA     string `json:"sha"`
}

type Repo struct {
	Name            string `json:"name"`
	FullName        string `json:"full_name"`
	Language        string `json:"language"`
	StargazersCount int    `json:"stargazers_count"`
	Fork            bool   `json:"fork"`
}

// --- Contribution calendar (scraped) ---

type ContributionDay struct {
	Date  time.Time
	Count int
	Level int // 0-4
}

// --- Computed stats for slides ---

type Stats struct {
	// Slide 1: Title
	Username  string
	Name      string
	YearLabel string // e.g. "2025" or "2025-2026"

	// Slide 2: Year in Numbers
	TotalContributions int
	TotalRepos         int
	TotalStars         int

	// Slide 3: Heatmap
	Calendar      []ContributionDay
	LongestStreak int
	StreakStart    time.Time
	StreakEnd      time.Time

	// Slide 4: Most Chaotic Day
	BusiestDate  time.Time
	BusiestCount int

	// Slide 5: Night Owl / Early Bird
	TimeBlocks [4]int // [06-12, 12-18, 18-00, 00-06] commit counts
	TimeLabel  string // "Night Owl", "Early Bird", "9-to-5er"

	// Slide 6: Top Languages
	Languages []LangStat

	// Slide 7: Villain Arc (most-pushed repo)
	VillainRepo    string
	VillainCommits int

	// Slide 8: Weekend Warrior
	WeekendPercent float64

	// Slide 9: Longest Commit Message
	LongestMessage     string
	LongestMessageRepo string
	LongestMessageLen  int

	// Slide 10: Personality
	Archetype            string
	ArchetypeDescription string
	Traits               [3]string
	TraitDescriptions    [3]string

	// Degraded mode flag
	HasCalendar bool
}

type LangStat struct {
	Name    string
	Count   int
	Percent float64
	Color   string // hex color from GitHub
}
