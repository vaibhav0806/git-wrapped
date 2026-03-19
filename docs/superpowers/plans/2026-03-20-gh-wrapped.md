# gh-wrapped Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Go CLI that presents a GitHub user's public activity as a colorful, animated terminal slideshow with GIF export.

**Architecture:** Data layer (GitHub REST client + HTML scraper) feeds stats computation, which feeds a Bubble Tea TUI with 10 animated slides. VHS handles GIF export. All data fetched concurrently via errgroup.

**Tech Stack:** Go, Bubble Tea (TUI), Lip Gloss (styling), Bubbles (spinner/progress), net/html (scraping), VHS (GIF export, runtime dep)

**Spec:** `docs/superpowers/specs/2026-03-20-gh-wrapped-design.md`

---

## File Structure

```
gh-wrapped/
├── main.go                  # CLI entry: arg parsing, fetch data, boot TUI
├── go.mod
├── github/
│   ├── client.go            # REST API client (user, events, repos)
│   ├── client_test.go       # Tests with httptest server
│   ├── scraper.go           # HTML scraper for contribution calendar
│   ├── scraper_test.go      # Tests with sample HTML fixtures
│   ├── models.go            # All API response types + computed stats struct
│   └── stats.go             # Compute slide stats from raw API data
│   └── stats_test.go        # Tests for stats computation
├── personality/
│   ├── personality.go       # Archetype scoring engine
│   └── personality_test.go  # Tests for archetype scoring
├── ui/
│   ├── app.go               # Bubble Tea Model: slide state machine, auto-play, controls
│   ├── slides.go            # Per-slide render functions (View logic)
│   ├── theme.go             # Lip Gloss styles, colors, gradient helpers
│   └── animation.go         # Tick-based animation state (counters, heatmap fill, typewriter)
└── export/
    └── gif.go               # VHS tape generation + execution
```

---

### Task 1: Project Scaffold

**Files:**
- Create: `go.mod`
- Create: `main.go` (skeleton)
- Create: `github/models.go`

- [ ] **Step 1: Initialize Go module and install dependencies**

```bash
cd /Users/vaibhav/Documents/projects/gh-wrapped
go mod init github.com/vaibhav/gh-wrapped
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/charmbracelet/bubbles
go get golang.org/x/net/html
go get golang.org/x/sync/errgroup
```

- [ ] **Step 2: Create skeleton main.go**

```go
// main.go
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: gh-wrapped <username> [--auto]")
		os.Exit(1)
	}
	username := os.Args[1]
	auto := len(os.Args) > 2 && os.Args[2] == "--auto"

	_ = username
	_ = auto
	fmt.Println("gh-wrapped: not yet implemented")
}
```

- [ ] **Step 3: Create models.go with all data types**

```go
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
	Username string
	Name     string
	YearLabel string // e.g. "2025" or "2025-2026"

	// Slide 2: Year in Numbers
	TotalContributions int
	TotalRepos         int
	TotalStars         int

	// Slide 3: Heatmap
	Calendar    []ContributionDay
	LongestStreak int
	StreakStart  time.Time
	StreakEnd    time.Time

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
	Archetype string
	Traits    [3]string

	// Degraded mode flag
	HasCalendar bool
}

type LangStat struct {
	Name    string
	Count   int
	Percent float64
	Color   string // hex color from GitHub
}
```

- [ ] **Step 4: Verify it compiles**

Run: `go build ./...`
Expected: no errors

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "feat: project scaffold with module, skeleton main, and data models"
```

---

### Task 2: GitHub REST API Client

**Files:**
- Create: `github/client.go`
- Create: `github/client_test.go`

- [ ] **Step 1: Write failing tests for client**

```go
// github/client_test.go
package github

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchUser(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/octocat" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(User{Login: "octocat", Name: "The Octocat"})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "")
	user, err := c.FetchUser("octocat")
	if err != nil {
		t.Fatalf("FetchUser: %v", err)
	}
	if user.Login != "octocat" {
		t.Errorf("got login %q, want octocat", user.Login)
	}
}

func TestFetchEvents(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			events := []Event{{Type: "PushEvent", Repo: EventRepo{Name: "octocat/hello"}}}
			w.Header().Set("Link", `<`+r.URL.Path+`?page=2>; rel="next"`)
			json.NewEncoder(w).Encode(events)
		} else {
			json.NewEncoder(w).Encode([]Event{})
		}
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "")
	events, err := c.FetchEvents("octocat")
	if err != nil {
		t.Fatalf("FetchEvents: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("got %d events, want 1", len(events))
	}
}

func TestFetchRepos(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		repos := []Repo{{Name: "hello", Language: "Go", StargazersCount: 42}}
		json.NewEncoder(w).Encode(repos)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "")
	repos, err := c.FetchRepos("octocat")
	if err != nil {
		t.Fatalf("FetchRepos: %v", err)
	}
	if len(repos) != 1 || repos[0].StargazersCount != 42 {
		t.Errorf("unexpected repos: %+v", repos)
	}
}

func TestFetchUserNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "")
	_, err := c.FetchUser("nobody")
	if err == nil {
		t.Fatal("expected error for 404")
	}
}

func TestFetchRateLimited(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "")
	_, err := c.FetchUser("octocat")
	if err == nil {
		t.Fatal("expected error for 403")
	}
}
```

- [ ] **Step 2: Run tests — verify they fail**

Run: `go test ./github/ -v`
Expected: compilation errors (NewClient not defined)

- [ ] **Step 3: Implement client.go**

```go
// github/client.go
package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		http:    &http.Client{},
	}
}

func (c *Client) get(path string) (*http.Response, error) {
	req, err := http.NewRequest("GET", c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, fmt.Errorf("not found")
	}
	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == 429 {
		resp.Body.Close()
		return nil, fmt.Errorf("rate limited")
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return resp, nil
}

func (c *Client) FetchUser(username string) (User, error) {
	resp, err := c.get("/users/" + username)
	if err != nil {
		return User{}, err
	}
	defer resp.Body.Close()
	var u User
	return u, json.NewDecoder(resp.Body).Decode(&u)
}

func (c *Client) FetchEvents(username string) ([]Event, error) {
	var all []Event
	for page := 1; page <= 3; page++ {
		path := fmt.Sprintf("/users/%s/events/public?per_page=100&page=%d", username, page)
		resp, err := c.get(path)
		if err != nil {
			return nil, err
		}
		var events []Event
		err = json.NewDecoder(resp.Body).Decode(&events)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		all = append(all, events...)
		if len(events) == 0 {
			break
		}
		// Check for next page via Link header
		link := resp.Header.Get("Link")
		if !strings.Contains(link, `rel="next"`) {
			break
		}
	}
	return all, nil
}

func (c *Client) FetchRepos(username string) ([]Repo, error) {
	var all []Repo
	for page := 1; ; page++ {
		path := fmt.Sprintf("/users/%s/repos?per_page=100&type=owner&page=%d", username, page)
		resp, err := c.get(path)
		if err != nil {
			return nil, err
		}
		var repos []Repo
		err = json.NewDecoder(resp.Body).Decode(&repos)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		all = append(all, repos...)
		if len(repos) < 100 {
			break
		}
	}
	return all, nil
}
```

- [ ] **Step 4: Run tests — verify they pass**

Run: `go test ./github/ -v`
Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add github/client.go github/client_test.go
git commit -m "feat: GitHub REST API client with tests"
```

---

### Task 3: Contribution Calendar Scraper

**Files:**
- Create: `github/scraper.go`
- Create: `github/scraper_test.go`
- Create: `github/testdata/contributions.html` (fixture)

- [ ] **Step 1: Create HTML fixture**

Fetch a real contributions page to create the fixture:

```bash
curl -s "https://github.com/users/torvalds/contributions" -o github/testdata/contributions.html
```

If curl fails or returns unexpected HTML, create a minimal fixture manually with the known `<td data-date data-level>` + `<tool-tip>` structure.

- [ ] **Step 2: Write failing tests**

```go
// github/scraper_test.go
package github

import (
	"strings"
	"testing"
)

func TestParseContributions(t *testing.T) {
	// Minimal HTML fixture matching GitHub's actual structure
	html := `<table>
<tbody>
<tr>
<td data-date="2025-01-01" data-level="2" class="ContributionCalendar-day"></td>
<td data-date="2025-01-02" data-level="0" class="ContributionCalendar-day"></td>
<td data-date="2025-01-03" data-level="4" class="ContributionCalendar-day"></td>
</tr>
</tbody>
</table>
<tool-tip for="contribution-day-component-0-0">5 contributions on January 1st.</tool-tip>
<tool-tip for="contribution-day-component-0-1">No contributions on January 2nd.</tool-tip>
<tool-tip for="contribution-day-component-0-2">32 contributions on January 3rd.</tool-tip>`

	days, err := ParseContributions(strings.NewReader(html))
	if err != nil {
		t.Fatalf("ParseContributions: %v", err)
	}
	if len(days) != 3 {
		t.Fatalf("got %d days, want 3", len(days))
	}
	if days[0].Level != 2 {
		t.Errorf("day 0 level: got %d, want 2", days[0].Level)
	}
	if days[0].Date.Format("2006-01-02") != "2025-01-01" {
		t.Errorf("day 0 date: got %s, want 2025-01-01", days[0].Date.Format("2006-01-02"))
	}
	if days[2].Level != 4 {
		t.Errorf("day 2 level: got %d, want 4", days[2].Level)
	}
	if days[0].Count != 5 {
		t.Errorf("day 0 count: got %d, want 5", days[0].Count)
	}
	if days[1].Count != 0 {
		t.Errorf("day 1 count: got %d, want 0", days[1].Count)
	}
	if days[2].Count != 32 {
		t.Errorf("day 2 count: got %d, want 32", days[2].Count)
	}
}

func TestParseContributionsFromFile(t *testing.T) {
	// Test with the real fixture if it exists
	days, err := ParseContributionsFromFile("testdata/contributions.html")
	if err != nil {
		t.Skipf("fixture not found: %v", err)
	}
	if len(days) < 300 {
		t.Errorf("expected 300+ days, got %d", len(days))
	}
	// Verify dates are in the expected range
	for _, d := range days {
		if d.Level < 0 || d.Level > 4 {
			t.Errorf("invalid level %d for %s", d.Level, d.Date.Format("2006-01-02"))
		}
	}
}

func TestParseContributionsBadHTML(t *testing.T) {
	html := `<div>no table here</div>`
	days, err := ParseContributions(strings.NewReader(html))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(days) != 0 {
		t.Errorf("expected 0 days for bad HTML, got %d", len(days))
	}
}
```

- [ ] **Step 3: Run tests — verify they fail**

Run: `go test ./github/ -run TestParse -v`
Expected: compilation errors

- [ ] **Step 4: Implement scraper.go**

```go
// github/scraper.go
package github

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

func (c *Client) FetchContributions(username string) ([]ContributionDay, error) {
	url := fmt.Sprintf("https://github.com/users/%s/contributions", username)
	if c.baseURL != "https://api.github.com" && c.baseURL != "" {
		// For testing: allow overriding the base URL
		url = c.baseURL + "/users/" + username + "/contributions"
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("scraper: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("scraper: status %d", resp.StatusCode)
	}
	return ParseContributions(resp.Body)
}

func ParseContributions(r io.Reader) ([]ContributionDay, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("parse HTML: %w", err)
	}

	// First pass: extract tooltip counts (actual contribution numbers)
	tooltipCounts := extractTooltipCounts(doc)

	// Second pass: extract cells and match with tooltip counts
	var days []ContributionDay
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "td" {
			date, level, ok := extractContributionCell(n)
			if ok {
				day := ContributionDay{Date: date, Level: level}
				// Try to get actual count from tooltip
				for _, a := range n.Attr {
					if a.Key == "id" {
						if count, found := tooltipCounts[a.Val]; found {
							day.Count = count
						}
					}
				}
				// Fallback: estimate count from level if tooltip not found
				if day.Count == 0 && day.Level > 0 {
					levelEstimates := [5]int{0, 3, 8, 15, 30}
					day.Count = levelEstimates[day.Level]
				}
				days = append(days, day)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return days, nil
}

func extractContributionCell(n *html.Node) (time.Time, int, bool) {
	var dateStr, levelStr string
	isContrib := false
	for _, a := range n.Attr {
		switch a.Key {
		case "data-date":
			dateStr = a.Val
		case "data-level":
			levelStr = a.Val
		case "class":
			if strings.Contains(a.Val, "ContributionCalendar-day") {
				isContrib = true
			}
		}
	}
	if !isContrib || dateStr == "" {
		return time.Time{}, 0, false
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, 0, false
	}
	level, _ := strconv.Atoi(levelStr)
	return date, level, true
}

// extractTooltipCounts parses <tool-tip> elements to get actual contribution counts.
// Returns map of element ID -> count.
// Tooltip format: "N contributions on Month Dayth." or "No contributions on Date."
func extractTooltipCounts(doc *html.Node) map[string]int {
	counts := map[string]int{}
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tool-tip" {
			forID := ""
			for _, a := range n.Attr {
				if a.Key == "for" {
					forID = a.Val
				}
			}
			if forID != "" {
				text := extractText(n)
				count := parseContributionCount(text)
				counts[forID] = count
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return counts
}

func extractText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		sb.WriteString(extractText(c))
	}
	return sb.String()
}

func parseContributionCount(text string) int {
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "No ") {
		return 0
	}
	// Format: "N contribution(s) on ..."
	parts := strings.SplitN(text, " ", 2)
	if len(parts) < 1 {
		return 0
	}
	count, _ := strconv.Atoi(parts[0])
	return count
}

func ParseContributionsFromFile(path string) ([]ContributionDay, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ParseContributions(f)
}
```

- [ ] **Step 5: Run tests — verify they pass**

Run: `go test ./github/ -run TestParse -v`
Expected: all PASS (TestParseContributionsFromFile may skip if fixture missing)

- [ ] **Step 6: Commit**

```bash
mkdir -p github/testdata
git add github/scraper.go github/scraper_test.go github/testdata/
git commit -m "feat: contribution calendar HTML scraper with tests"
```

---

### Task 4: Stats Computation

**Files:**
- Create: `github/stats.go`
- Create: `github/stats_test.go`

- [ ] **Step 1: Write failing tests**

```go
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
		{Date: makeDate("2025-01-01"), Level: 2},
		{Date: makeDate("2025-01-02"), Level: 3},
		{Date: makeDate("2025-12-31"), Level: 1},
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
```

- [ ] **Step 2: Run tests — verify they fail**

Run: `go test ./github/ -run TestComputeStats -v`
Expected: compilation errors (ComputeStats not defined)

- [ ] **Step 3: Implement stats.go**

```go
// github/stats.go
package github

import (
	"sort"
	"time"
)

// GitHub language colors — top languages
var LanguageColors = map[string]string{
	"Go":         "#00ADD8",
	"JavaScript": "#f1e05a",
	"TypeScript": "#3178c6",
	"Python":     "#3572A5",
	"Rust":       "#dea584",
	"Java":       "#b07219",
	"C":          "#555555",
	"C++":        "#f34b7d",
	"C#":         "#178600",
	"Ruby":       "#701516",
	"PHP":        "#4F5D95",
	"Swift":      "#F05138",
	"Kotlin":     "#A97BFF",
	"Shell":      "#89e051",
	"HTML":       "#e34c26",
	"CSS":        "#563d7c",
	"Dart":       "#00B4AB",
	"Lua":        "#000080",
	"Zig":        "#ec915c",
	"Elixir":     "#6e4a7e",
	"Haskell":    "#5e5086",
	"Scala":      "#c22d40",
	"R":          "#198CE7",
	"Vue":        "#41b883",
	"Svelte":     "#ff3e00",
}

func ComputeStats(user User, events []Event, repos []Repo, calendar []ContributionDay) Stats {
	s := Stats{
		Username:    user.Login,
		Name:        user.Name,
		HasCalendar: len(calendar) > 0,
	}

	// Year label
	s.YearLabel = computeYearLabel(calendar)

	// Repos + Stars
	s.TotalRepos = len(repos)
	for _, r := range repos {
		s.TotalStars += r.StargazersCount
	}

	// Calendar-based stats
	if len(calendar) > 0 {
		computeCalendarStats(&s, calendar)
	}

	// Events-based stats
	if len(events) > 0 {
		computeEventStats(&s, events)
	}

	// Language stats
	if len(repos) > 0 {
		computeLanguageStats(&s, repos)
	}

	return s
}

func computeYearLabel(calendar []ContributionDay) string {
	if len(calendar) == 0 {
		return time.Now().Format("2006")
	}
	first := calendar[0].Date.Year()
	last := calendar[len(calendar)-1].Date.Year()
	if first == last {
		return time.Date(first, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006")
	}
	return time.Date(first, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006") + "-" + time.Date(last, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006")
}

func computeCalendarStats(s *Stats, calendar []ContributionDay) {
	// Total contributions (sum actual counts from tooltip data)
	for _, d := range calendar {
		s.TotalContributions += d.Count
	}

	// Longest streak
	streak, bestStreak := 0, 0
	var bestStart, bestEnd, streakStart time.Time
	for _, d := range calendar {
		if d.Count > 0 {
			if streak == 0 {
				streakStart = d.Date
			}
			streak++
			if streak > bestStreak {
				bestStreak = streak
				bestStart = streakStart
				bestEnd = d.Date
			}
		} else {
			streak = 0
		}
	}
	s.LongestStreak = bestStreak
	s.StreakStart = bestStart
	s.StreakEnd = bestEnd

	// Busiest day (highest actual count)
	for _, d := range calendar {
		if d.Count > s.BusiestCount {
			s.BusiestCount = d.Count
			s.BusiestDate = d.Date
		}
	}

	// Weekend percent (weighted by contribution count)
	weekendContribs, totalContribs := 0, 0
	for _, d := range calendar {
		if d.Count > 0 {
			totalContribs += d.Count
			wd := d.Date.Weekday()
			if wd == time.Saturday || wd == time.Sunday {
				weekendContribs += d.Count
			}
		}
	}
	if totalContribs > 0 {
		s.WeekendPercent = float64(weekendContribs) / float64(totalContribs) * 100
	}
}

func computeEventStats(s *Stats, events []Event) {
	// Time blocks (UTC)
	for _, e := range events {
		if e.Type != "PushEvent" {
			continue
		}
		h := e.CreatedAt.Hour()
		switch {
		case h >= 6 && h < 12:
			s.TimeBlocks[0]++
		case h >= 12 && h < 18:
			s.TimeBlocks[1]++
		case h >= 18:
			s.TimeBlocks[2]++
		default: // 0-5
			s.TimeBlocks[3]++
		}
	}
	s.TimeLabel = computeTimeLabel(s.TimeBlocks)

	// Villain arc (most-pushed repo)
	repoCounts := map[string]int{}
	for _, e := range events {
		if e.Type == "PushEvent" {
			repoCounts[e.Repo.Name] += len(e.Payload.Commits)
		}
	}
	for repo, count := range repoCounts {
		if count > s.VillainCommits {
			s.VillainCommits = count
			s.VillainRepo = repo
		}
	}

	// Longest commit message
	for _, e := range events {
		if e.Type != "PushEvent" {
			continue
		}
		for _, c := range e.Payload.Commits {
			if len(c.Message) > s.LongestMessageLen {
				s.LongestMessageLen = len(c.Message)
				s.LongestMessage = c.Message
				s.LongestMessageRepo = e.Repo.Name
			}
		}
	}
}

func computeTimeLabel(blocks [4]int) string {
	// blocks: [06-12, 12-18, 18-00, 00-06]
	evening := blocks[2] + blocks[3] // after 6pm + before 6am
	morning := blocks[0]             // 6am-12pm
	total := blocks[0] + blocks[1] + blocks[2] + blocks[3]
	if total == 0 {
		return "Ghost"
	}
	eveningPct := float64(evening) / float64(total)
	morningPct := float64(morning) / float64(total)
	if eveningPct > 0.5 {
		return "Night Owl"
	}
	if morningPct > 0.35 {
		return "Early Bird"
	}
	return "9-to-5er"
}

func computeLanguageStats(s *Stats, repos []Repo) {
	counts := map[string]int{}
	for _, r := range repos {
		if r.Language != "" {
			counts[r.Language]++
		}
	}
	total := 0
	for _, c := range counts {
		total += c
	}

	var langs []LangStat
	for name, count := range counts {
		color := LanguageColors[name]
		if color == "" {
			color = "#888888"
		}
		langs = append(langs, LangStat{
			Name:    name,
			Count:   count,
			Percent: float64(count) / float64(total) * 100,
			Color:   color,
		})
	}
	sort.Slice(langs, func(i, j int) bool {
		return langs[i].Count > langs[j].Count
	})
	if len(langs) > 8 {
		langs = langs[:8]
	}
	s.Languages = langs
}
```

- [ ] **Step 4: Run tests — verify they pass**

Run: `go test ./github/ -run TestComputeStats -v`
Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add github/stats.go github/stats_test.go
git commit -m "feat: stats computation from API data with tests"
```

---

### Task 5: Personality Engine

**Files:**
- Create: `personality/personality.go`
- Create: `personality/personality_test.go`

- [ ] **Step 1: Write failing tests**

```go
// personality/personality_test.go
package personality

import (
	"testing"

	"github.com/vaibhav/gh-wrapped/github"
)

func TestScoreNightcrawler(t *testing.T) {
	s := github.Stats{
		TimeBlocks: [4]int{1, 1, 8, 2}, // 06-12:1, 12-18:1, 18-00:8, 00-06:2 => 83% evening
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
	// Calendar-dependent archetypes should score 0
	// Should fall back to events-based archetype
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
```

- [ ] **Step 2: Run tests — verify they fail**

Run: `go test ./personality/ -v`
Expected: compilation errors

- [ ] **Step 3: Implement personality.go**

```go
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

type archetype struct {
	Name  string
	Score float64
}

func Compute(s github.Stats) Result {
	archetypes := scoreAll(s)
	sort.SliceStable(archetypes, func(i, j int) bool {
		return archetypes[i].Score > archetypes[j].Score
	})

	result := Result{
		Archetype: archetypes[0].Name,
	}
	for i := 0; i < 3 && i < len(archetypes); i++ {
		result.Traits[i] = archetypes[i].Name
	}
	return result
}

func scoreAll(s github.Stats) []archetype {
	var a []archetype

	// Events-based archetypes
	a = append(a, archetype{"The Nightcrawler", scoreNightcrawler(s)})
	a = append(a, archetype{"The Obsessed", scoreObsessed(s)})
	a = append(a, archetype{"The Novelist", scoreNovelist(s)})

	// Repos-based archetypes
	a = append(a, archetype{"The Polyglot", scorePolyglot(s)})
	a = append(a, archetype{"The Specialist", scoreSpecialist(s)})

	// Calendar-based archetypes (0 in degraded mode)
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
	// Count total push commits from time blocks as proxy
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
```

- [ ] **Step 4: Run tests — verify they pass**

Run: `go test ./personality/ -v`
Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add personality/
git commit -m "feat: personality archetype engine with scoring and tests"
```

---

### Task 6: Theme & Styling

**Files:**
- Create: `ui/theme.go`

- [ ] **Step 1: Implement theme.go**

No TDD for this task — it's purely visual configuration, not testable logic.

```go
// ui/theme.go
package ui

import "github.com/charmbracelet/lipgloss"

// Slide color palettes
var (
	ColorRed     = lipgloss.Color("#ff6b6b")
	ColorYellow  = lipgloss.Color("#feca57")
	ColorGreen   = lipgloss.Color("#0be881")
	ColorCyan    = lipgloss.Color("#48dbfb")
	ColorPurple  = lipgloss.Color("#a29bfe")
	ColorPink    = lipgloss.Color("#ff9ff3")
	ColorCrimson = lipgloss.Color("#ff4757")
	ColorWhite   = lipgloss.Color("#e8e8e8")
	ColorDim     = lipgloss.Color("#555555")
	ColorMuted   = lipgloss.Color("#888888")
	ColorBg      = lipgloss.Color("#08080c")
)

// Slide accent colors (indexed by slide number 0-9)
var SlideAccents = [10]lipgloss.Color{
	ColorPink,    // 0: Title
	ColorYellow,  // 1: Year in Numbers
	ColorGreen,   // 2: Heatmap
	ColorRed,     // 3: Chaotic Day
	ColorPurple,  // 4: Night Owl
	ColorCyan,    // 5: Languages
	ColorCrimson, // 6: Villain Arc
	ColorPink,    // 7: Weekend Warrior
	ColorYellow,  // 8: The Novel
	ColorCyan,    // 9: Personality
}

// Reusable styles
var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Align(lipgloss.Center)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Align(lipgloss.Center)

	BigNumberStyle = lipgloss.NewStyle().
			Bold(true).
			Align(lipgloss.Center)

	LabelStyle = lipgloss.NewStyle().
			Foreground(ColorDim).
			Align(lipgloss.Center)

	FootnoteStyle = lipgloss.NewStyle().
			Foreground(ColorDim).
			Italic(true).
			Align(lipgloss.Center)

	PillStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Bold(true)

	BarStyle = lipgloss.NewStyle()
)

// GradientText renders text with a color gradient between two colors
func GradientText(text string, from, to lipgloss.Color) string {
	if len(text) == 0 {
		return ""
	}
	colors := lipgloss.Blend1D(len(text), from, to)
	result := ""
	for i, ch := range text {
		style := lipgloss.NewStyle().Foreground(colors[i])
		result += style.Render(string(ch))
	}
	return result
}

// RenderBar renders a horizontal bar chart segment
func RenderBar(width int, color lipgloss.Color) string {
	bar := ""
	for i := 0; i < width; i++ {
		bar += "█"
	}
	return lipgloss.NewStyle().Foreground(color).Render(bar)
}

// RenderBarEmpty renders empty bar segment
func RenderBarEmpty(width int) string {
	bar := ""
	for i := 0; i < width; i++ {
		bar += "░"
	}
	return lipgloss.NewStyle().Foreground(ColorDim).Render(bar)
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./ui/...`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add ui/theme.go
git commit -m "feat: Lip Gloss theme with slide colors and gradient helpers"
```

---

### Task 7: Animation System

**Files:**
- Create: `ui/animation.go`

- [ ] **Step 1: Implement animation.go**

```go
// ui/animation.go
package ui

import "time"

// AnimState tracks animation progress for a single slide
type AnimState struct {
	Tick     int           // current tick count
	MaxTicks int           // total ticks for animation to complete
	Done     bool
	Interval time.Duration // tick interval
}

func NewAnimState(durationMs int, intervalMs int) AnimState {
	interval := time.Duration(intervalMs) * time.Millisecond
	maxTicks := durationMs / intervalMs
	return AnimState{
		Tick:     0,
		MaxTicks: maxTicks,
		Interval: interval,
	}
}

func (a *AnimState) Advance() {
	if a.Tick < a.MaxTicks {
		a.Tick++
	}
	if a.Tick >= a.MaxTicks {
		a.Done = true
	}
}

// Progress returns 0.0 to 1.0
func (a *AnimState) Progress() float64 {
	if a.MaxTicks == 0 {
		return 1.0
	}
	p := float64(a.Tick) / float64(a.MaxTicks)
	if p > 1.0 {
		return 1.0
	}
	return p
}

// CounterAnimation returns the current display value for a counter that ticks up to target
func CounterAnimation(target int, progress float64) int {
	return int(float64(target) * progress)
}

// TypewriterAnimation returns the substring to display for a typewriter effect
func TypewriterAnimation(text string, progress float64) string {
	chars := int(float64(len(text)) * progress)
	if chars > len(text) {
		chars = len(text)
	}
	return text[:chars]
}

// HeatmapAnimation returns how many cells to reveal (left-to-right fill)
func HeatmapAnimation(totalCells int, progress float64) int {
	return int(float64(totalCells) * progress)
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./ui/...`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add ui/animation.go
git commit -m "feat: tick-based animation system for slides"
```

---

### Task 8: Slide Renderers

**Files:**
- Create: `ui/slides.go`

This is the largest task — each slide gets a render function. Build all 10 slides.

- [ ] **Step 1: Implement slides.go**

```go
// ui/slides.go
package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/vaibhav/gh-wrapped/github"
)

// SlideID identifies which slide to render
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

// ActiveSlides returns the slide IDs to show, skipping those with insufficient data
func ActiveSlides(s github.Stats) []SlideID {
	slides := []SlideID{SlideTitle}

	if s.HasCalendar {
		slides = append(slides, SlideNumbers)
		slides = append(slides, SlideHeatmap)
		if s.BusiestCount > 0 {
			slides = append(slides, SlideChaos)
		}
	}

	total := s.TimeBlocks[0] + s.TimeBlocks[1] + s.TimeBlocks[2] + s.TimeBlocks[3]
	if total > 0 {
		slides = append(slides, SlideClock)
	}

	if len(s.Languages) > 0 {
		slides = append(slides, SlideLanguages)
	}

	if s.VillainCommits > 0 {
		slides = append(slides, SlideVillain)
	}

	if s.HasCalendar && s.WeekendPercent > 0 {
		slides = append(slides, SlideWeekend)
	}

	if s.LongestMessageLen > 0 {
		slides = append(slides, SlideNovel)
	}

	slides = append(slides, SlidePersonality)
	return slides
}

// RenderSlide renders a slide at the given animation progress
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
	lines := strings.Count(content, "\n") + 1
	padTop := (height - lines) / 2
	if padTop < 0 {
		padTop = 0
	}
	return strings.Repeat("\n", padTop) + content
}

func renderTitle(s github.Stats, anim AnimState, w int) string {
	title := "DEV WRAPPED " + s.YearLabel
	gradTitle := GradientText(title, ColorRed, ColorCyan)
	centered := lipgloss.NewStyle().Width(w).Align(lipgloss.Center)

	lines := []string{
		centered.Render(SubtitleStyle.Render("~ ~ ~")),
		"",
		centered.Render(lipgloss.NewStyle().Bold(true).Render(gradTitle)),
		"",
		centered.Render(SubtitleStyle.Render("@" + s.Username)),
		"",
		centered.Render(SubtitleStyle.Render("~ ~ ~")),
	}
	return strings.Join(lines, "\n")
}

func renderNumbers(s github.Stats, anim AnimState, w int) string {
	p := anim.Progress()
	centered := lipgloss.NewStyle().Width(w).Align(lipgloss.Center)

	contribs := CounterAnimation(s.TotalContributions, p)
	repos := CounterAnimation(s.TotalRepos, p)
	stars := CounterAnimation(s.TotalStars, p)

	header := centered.Render(LabelStyle.Render("YOUR YEAR IN NUMBERS"))

	numStyle := func(c lipgloss.Color) lipgloss.Style {
		return lipgloss.NewStyle().Foreground(c).Bold(true)
	}
	labelSt := lipgloss.NewStyle().Foreground(ColorDim)

	col1 := fmt.Sprintf("%s\n%s", numStyle(ColorYellow).Render(fmt.Sprintf("%d", contribs)), labelSt.Render("CONTRIBUTIONS"))
	col2 := fmt.Sprintf("%s\n%s", numStyle(ColorCyan).Render(fmt.Sprintf("%d", repos)), labelSt.Render("REPOS"))
	col3 := fmt.Sprintf("%s\n%s", numStyle(ColorGreen).Render(fmt.Sprintf("%d", stars)), labelSt.Render("STARS"))

	row := lipgloss.JoinHorizontal(lipgloss.Center,
		lipgloss.NewStyle().Width(w/3).Align(lipgloss.Center).Render(col1),
		lipgloss.NewStyle().Width(w/3).Align(lipgloss.Center).Render(col2),
		lipgloss.NewStyle().Width(w/3).Align(lipgloss.Center).Render(col3),
	)

	return header + "\n\n" + centered.Render(row)
}

func renderHeatmap(s github.Stats, anim AnimState, w int) string {
	centered := lipgloss.NewStyle().Width(w).Align(lipgloss.Center)
	header := centered.Render(LabelStyle.Render("YOUR CONTRIBUTION MAP"))

	revealed := HeatmapAnimation(len(s.Calendar), anim.Progress())

	// Render 7 rows (days of week) x N weeks
	weeks := (len(s.Calendar) + 6) / 7
	maxWeeks := w - 10 // leave margin
	if weeks > maxWeeks {
		weeks = maxWeeks
	}

	greenStyle := lipgloss.NewStyle().Foreground(ColorGreen)
	dimStyle := lipgloss.NewStyle().Foreground(ColorDim)

	var rows []string
	for row := 0; row < 7; row++ {
		line := ""
		for col := 0; col < weeks; col++ {
			idx := col*7 + row
			if idx >= len(s.Calendar) {
				break
			}
			if idx >= revealed {
				line += dimStyle.Render("░")
			} else {
				switch s.Calendar[idx].Level {
				case 0:
					line += dimStyle.Render("░")
				case 1:
					line += lipgloss.NewStyle().Foreground(ColorGreen).Faint(true).Render("▒")
				case 2:
					line += greenStyle.Render("▓")
				case 3:
					line += greenStyle.Render("█")
				case 4:
					line += lipgloss.NewStyle().Foreground(ColorYellow).Render("█")
				}
			}
		}
		rows = append(rows, line)
	}

	streakLine := fmt.Sprintf("%s %s",
		lipgloss.NewStyle().Foreground(ColorGreen).Bold(true).Render(fmt.Sprintf("%d-day streak", s.LongestStreak)),
		SubtitleStyle.Render(fmt.Sprintf("(%s to %s)", s.StreakStart.Format("Jan 2"), s.StreakEnd.Format("Jan 2"))),
	)

	return header + "\n\n" + centered.Render(strings.Join(rows, "\n")) + "\n\n" + centered.Render(streakLine)
}

func renderChaos(s github.Stats, anim AnimState, w int) string {
	centered := lipgloss.NewStyle().Width(w).Align(lipgloss.Center)

	snarky := []string{"were you okay?", "deadline energy detected", "absolute chaos", "touch grass, maybe?"}
	quip := snarky[s.BusiestDate.YearDay()%len(snarky)]

	lines := []string{
		centered.Render(lipgloss.NewStyle().Foreground(ColorRed).Render("YOUR MOST CHAOTIC DAY")),
		"",
		centered.Render(lipgloss.NewStyle().Foreground(ColorWhite).Bold(true).Render(s.BusiestDate.Format("January 2, 2006"))),
		"",
		centered.Render(lipgloss.NewStyle().Foreground(ColorRed).Bold(true).Render(fmt.Sprintf("%d contributions", s.BusiestCount))),
		"",
		centered.Render(FootnoteStyle.Render(fmt.Sprintf(`"%s"`, quip))),
	}
	return strings.Join(lines, "\n")
}

func renderClock(s github.Stats, anim AnimState, w int) string {
	centered := lipgloss.NewStyle().Width(w).Align(lipgloss.Center)
	header := centered.Render(LabelStyle.Render("WHEN YOU CODE"))

	total := s.TimeBlocks[0] + s.TimeBlocks[1] + s.TimeBlocks[2] + s.TimeBlocks[3]
	labels := [4]string{"06-12", "12-18", "18-00", "00-06"}
	barWidth := 30

	var rows []string
	for i := 0; i < 4; i++ {
		pct := float64(s.TimeBlocks[i]) / float64(total) * 100
		filled := int(float64(barWidth) * pct / 100)
		if filled > barWidth {
			filled = barWidth
		}
		bar := RenderBar(filled, ColorPurple) + RenderBarEmpty(barWidth-filled)
		row := fmt.Sprintf("  %s  %s  %s",
			lipgloss.NewStyle().Foreground(ColorMuted).Width(5).Render(labels[i]),
			bar,
			lipgloss.NewStyle().Foreground(ColorDim).Width(4).Align(lipgloss.Right).Render(fmt.Sprintf("%.0f%%", pct)),
		)
		rows = append(rows, row)
	}

	label := lipgloss.NewStyle().Foreground(ColorPurple).Bold(true).Render(s.TimeLabel)
	footnote := centered.Render(FootnoteStyle.Render("Based on your last 30 days"))

	return header + "\n\n" + centered.Render(strings.Join(rows, "\n")) + "\n\n" + centered.Render(label) + "\n" + footnote
}

func renderLanguages(s github.Stats, anim AnimState, w int) string {
	centered := lipgloss.NewStyle().Width(w).Align(lipgloss.Center)
	header := centered.Render(LabelStyle.Render("YOUR TOP LANGUAGES"))

	barWidth := 30
	maxLangs := 5
	if len(s.Languages) < maxLangs {
		maxLangs = len(s.Languages)
	}

	var rows []string
	for i := 0; i < maxLangs; i++ {
		l := s.Languages[i]
		filled := int(float64(barWidth) * l.Percent / 100)
		color := lipgloss.Color(l.Color)
		bar := RenderBar(filled, color) + RenderBarEmpty(barWidth-filled)
		row := fmt.Sprintf("  %s  %s  %s",
			lipgloss.NewStyle().Foreground(color).Width(10).Render(l.Name),
			bar,
			lipgloss.NewStyle().Foreground(ColorDim).Width(4).Align(lipgloss.Right).Render(fmt.Sprintf("%.0f%%", l.Percent)),
		)
		rows = append(rows, row)
	}

	// Fun subtitle
	subtitle := ""
	if len(s.Languages) >= 4 && s.Languages[0].Percent < 40 {
		subtitle = "polyglot energy"
	} else if len(s.Languages) > 0 && s.Languages[0].Percent > 70 {
		subtitle = s.Languages[0].Name + " purist"
	}
	sub := ""
	if subtitle != "" {
		sub = "\n" + centered.Render(FootnoteStyle.Render(fmt.Sprintf(`"%s"`, subtitle)))
	}

	return header + "\n\n" + centered.Render(strings.Join(rows, "\n")) + sub
}

func renderVillain(s github.Stats, anim AnimState, w int) string {
	centered := lipgloss.NewStyle().Width(w).Align(lipgloss.Center)

	lines := []string{
		centered.Render(lipgloss.NewStyle().Foreground(ColorCrimson).Render("VILLAIN ARC")),
		"",
		centered.Render(lipgloss.NewStyle().Foreground(ColorCrimson).Bold(true).Render(fmt.Sprintf("%d commits", s.VillainCommits))),
		"",
		centered.Render(fmt.Sprintf("%s %s",
			SubtitleStyle.Render("to"),
			lipgloss.NewStyle().Foreground(ColorWhite).Bold(true).Render(s.VillainRepo),
		)),
		"",
		centered.Render(FootnoteStyle.Render(`"obsessed much?"`)),
		"",
		centered.Render(FootnoteStyle.Render("Based on your last 30 days")),
	}
	return strings.Join(lines, "\n")
}

func renderWeekend(s github.Stats, anim AnimState, w int) string {
	centered := lipgloss.NewStyle().Width(w).Align(lipgloss.Center)
	header := centered.Render(LabelStyle.Render("WEEKEND WARRIOR SCORE"))

	weekdayPct := 100 - s.WeekendPercent
	barWidth := 40
	weekdayW := int(float64(barWidth) * weekdayPct / 100)
	weekendW := barWidth - weekdayW

	bar := RenderBarEmpty(weekdayW) + RenderBar(weekendW, ColorPink)
	barLine := fmt.Sprintf("  weekdays %s %.0f%%", bar, s.WeekendPercent)

	var verdict string
	if s.WeekendPercent > 30 {
		verdict = "work-life balance is a myth"
	} else if s.WeekendPercent < 10 {
		verdict = "touch grass? already on it"
	} else {
		verdict = "balanced, for a developer"
	}

	lines := []string{
		header,
		"",
		centered.Render(barLine),
		"",
		centered.Render(lipgloss.NewStyle().Foreground(ColorPink).Bold(true).Render(fmt.Sprintf(`"%s"`, verdict))),
	}
	return strings.Join(lines, "\n")
}

func renderNovel(s github.Stats, anim AnimState, w int) string {
	centered := lipgloss.NewStyle().Width(w).Align(lipgloss.Center)
	header := centered.Render(LabelStyle.Render("YOUR LONGEST COMMIT MESSAGE"))

	msg := TypewriterAnimation(s.LongestMessage, anim.Progress())
	// Truncate display to max 300 chars
	if len(msg) > 300 {
		msg = msg[:300] + "..."
	}
	// Word wrap at ~60 chars
	wrapped := wordWrap(msg, 60)

	msgBox := lipgloss.NewStyle().
		Foreground(ColorYellow).
		Width(64).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorDim).
		Render(wrapped)

	info := fmt.Sprintf("%s in %s",
		lipgloss.NewStyle().Foreground(ColorYellow).Bold(true).Render(fmt.Sprintf("%d characters", s.LongestMessageLen)),
		lipgloss.NewStyle().Foreground(ColorWhite).Render(s.LongestMessageRepo),
	)

	lines := []string{
		header,
		"",
		centered.Render(msgBox),
		"",
		centered.Render(info),
		centered.Render(FootnoteStyle.Render(`"that's not a commit message, that's a blog post"`)),
		"",
		centered.Render(FootnoteStyle.Render("Based on your last 30 days")),
	}
	return strings.Join(lines, "\n")
}

func renderPersonality(s github.Stats, anim AnimState, w int) string {
	centered := lipgloss.NewStyle().Width(w).Align(lipgloss.Center)

	archGrad := GradientText(s.Archetype, ColorRed, ColorCyan)

	// Trait pills
	traitColors := [3]lipgloss.Color{ColorPurple, ColorGreen, ColorPink}
	var pills []string
	for i, t := range s.Traits {
		if t == "" {
			continue
		}
		pill := PillStyle.
			Foreground(traitColors[i]).
			Render(strings.ToLower(strings.TrimPrefix(t, "The ")))
		pills = append(pills, pill)
	}
	pillRow := strings.Join(pills, "  ")

	lastSlideHint := lipgloss.NewStyle().Foreground(ColorCyan).Render("press ") +
		lipgloss.NewStyle().Foreground(ColorCyan).Bold(true).Render("g") +
		lipgloss.NewStyle().Foreground(ColorCyan).Render(" to export as GIF")

	yearLabel := s.YearLabel

	lines := []string{
		centered.Render(SubtitleStyle.Render("you are")),
		"",
		centered.Render(lipgloss.NewStyle().Bold(true).Render(archGrad)),
		"",
		centered.Render(pillRow),
		"",
		"",
		centered.Render(SubtitleStyle.Render("Your " + yearLabel + ", Unwrapped.")),
		"",
		centered.Render(lastSlideHint),
	}
	return strings.Join(lines, "\n")
}

func wordWrap(s string, maxWidth int) string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return s
	}
	var lines []string
	line := words[0]
	for _, w := range words[1:] {
		if len(line)+1+len(w) > maxWidth {
			lines = append(lines, line)
			line = w
		} else {
			line += " " + w
		}
	}
	lines = append(lines, line)
	return strings.Join(lines, "\n")
}

```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./ui/...`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add ui/slides.go
git commit -m "feat: all 10 slide renderers with animations and styling"
```

---

### Task 9: Bubble Tea App

**Files:**
- Create: `ui/app.go`

- [ ] **Step 1: Implement app.go**

```go
// ui/app.go
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

type tickMsg time.Time

type Model struct {
	stats    github.Stats
	slides   []SlideID
	current  int
	anim     AnimState
	autoPlay bool
	autoMode bool // --auto flag (no keyboard, exit after last)
	width    int
	height   int
	tooSmall bool
	done     bool
	exportGIF bool
}

func NewModel(stats github.Stats, auto bool) Model {
	slides := ActiveSlides(stats)
	return Model{
		stats:    stats,
		slides:   slides,
		current:  0,
		anim:     NewAnimState(AnimDurationMs, AnimIntervalMs),
		autoPlay: true,
		autoMode: auto,
		width:    120,
		height:   40,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tick(),
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
		if m.width < MinWidth || m.height < MinHeight {
			m.tooSmall = true
		} else {
			m.tooSmall = false
		}
		return m, nil

	case tea.KeyMsg:
		if m.autoMode {
			return m, nil // ignore keyboard in auto mode
		}
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "right", "l", " ":
			m.autoPlay = false
			return m.nextSlide(), nil
		case "left", "h":
			m.autoPlay = false
			return m.prevSlide(), nil
		case "a":
			m.autoPlay = true
			return m, nil
		case "g":
			if m.current == len(m.slides)-1 {
				m.exportGIF = true
				return m, tea.Quit
			}
		}
		return m, nil

	case tickMsg:
		// Advance animation
		if !m.anim.Done {
			m.anim.Advance()
		}

		// Auto-advance slide when animation is done
		if m.autoPlay && m.anim.Done {
			// Wait a beat after animation completes before advancing
			if m.anim.Tick > m.anim.MaxTicks+(AutoAdvanceMs-AnimDurationMs)/AnimIntervalMs {
				if m.current < len(m.slides)-1 {
					m = m.nextSlide()
				} else if m.autoMode {
					return m, tea.Quit
				}
			}
		}

		return m, tick()
	}

	return m, nil
}

func (m Model) nextSlide() Model {
	if m.current < len(m.slides)-1 {
		m.current++
		m.anim = NewAnimState(AnimDurationMs, AnimIntervalMs)
	}
	return m
}

func (m Model) prevSlide() Model {
	if m.current > 0 {
		m.current--
		m.anim = NewAnimState(AnimDurationMs, AnimIntervalMs)
	}
	return m
}

func (m Model) View() string {
	if m.tooSmall {
		return "Terminal too small. Please resize to at least 100x30."
	}
	if m.done {
		return ""
	}

	slideID := m.slides[m.current]
	content := RenderSlide(slideID, m.stats, m.anim, m.width, m.height)

	// Bottom bar: slide counter + controls hint
	counter := LabelStyle.Render(
		fmt.Sprintf("%d/%d", m.current+1, len(m.slides)),
	)
	var hint string
	if !m.autoMode {
		if m.autoPlay {
			hint = LabelStyle.Render("← → navigate  q quit")
		} else {
			hint = LabelStyle.Render("← → navigate  a auto-play  q quit")
		}
	}
	bar := lipgloss.JoinHorizontal(lipgloss.Bottom,
		lipgloss.NewStyle().Width(m.width/2).Align(lipgloss.Left).Render(counter),
		lipgloss.NewStyle().Width(m.width/2).Align(lipgloss.Right).Render(hint),
	)

	return content + "\n\n" + bar
}

// ExportRequested returns true if the user pressed 'g' to export
func (m Model) ExportRequested() bool {
	return m.exportGIF
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./ui/...`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add ui/app.go
git commit -m "feat: Bubble Tea app with slide state machine and auto-play"
```

---

### Task 10: VHS/GIF Export

**Files:**
- Create: `export/gif.go`

- [ ] **Step 1: Implement gif.go**

```go
// export/gif.go
package export

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const tapeTemplate = `# gh-wrapped VHS tape
Output %s
Set FontSize 14
Set FontFamily "JetBrains Mono"
Set Width 1200
Set Height 800
Set Theme { "name": "gh-wrapped", "black": "#08080c", "red": "#ff6b6b", "green": "#0be881", "yellow": "#feca57", "blue": "#48dbfb", "magenta": "#a29bfe", "cyan": "#48dbfb", "white": "#e8e8e8", "brightBlack": "#555555", "brightRed": "#ff4757", "brightGreen": "#0be881", "brightYellow": "#feca57", "brightBlue": "#48dbfb", "brightMagenta": "#ff9ff3", "brightCyan": "#48dbfb", "brightWhite": "#ffffff", "background": "#08080c", "foreground": "#e8e8e8", "selectionBackground": "#333333", "cursorColor": "#e8e8e8" }

Type "gh-wrapped %s --auto" Enter
Sleep 500ms
Wait
`

func GenerateGIF(username string) error {
	// Check VHS is installed
	if _, err := exec.LookPath("vhs"); err != nil {
		return fmt.Errorf("VHS not found. Install with: brew install vhs")
	}

	outputFile := fmt.Sprintf("gh-wrapped-%s.gif", username)
	absOutput, _ := filepath.Abs(outputFile)

	// Write tape file
	tapeContent := fmt.Sprintf(tapeTemplate, absOutput, username)
	tapeFile := fmt.Sprintf("gh-wrapped-%s.tape", username)
	if err := os.WriteFile(tapeFile, []byte(tapeContent), 0644); err != nil {
		return fmt.Errorf("write tape: %w", err)
	}
	defer os.Remove(tapeFile)

	// Run VHS
	fmt.Printf("Recording GIF...\n")
	cmd := exec.Command("vhs", tapeFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("vhs failed: %w", err)
	}

	fmt.Printf("GIF saved to %s\n", absOutput)
	return nil
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./export/...`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add export/gif.go
git commit -m "feat: VHS tape generation for GIF export"
```

---

### Task 11: Wire Everything in main.go

**Files:**
- Modify: `main.go`

- [ ] **Step 1: Implement main.go**

```go
// main.go
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/sync/errgroup"

	"github.com/vaibhav/gh-wrapped/export"
	"github.com/vaibhav/gh-wrapped/github"
	"github.com/vaibhav/gh-wrapped/personality"
	"github.com/vaibhav/gh-wrapped/ui"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: gh-wrapped <username> [--auto]")
		os.Exit(1)
	}
	username := os.Args[1]
	auto := len(os.Args) > 2 && os.Args[2] == "--auto"

	token := os.Getenv("GITHUB_TOKEN")
	client := github.NewClient("https://api.github.com", token)

	// Fetch all data concurrently
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
		if err != nil {
			return fmt.Errorf("user: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		var err error
		events, err = client.FetchEvents(username)
		if err != nil {
			// Non-fatal: events are nice-to-have
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
			// Degraded mode: calendar scraping failed
			calendar = nil
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Check for zero activity
	if len(events) == 0 && len(calendar) == 0 && len(repos) == 0 {
		fmt.Printf("Looks like @%s had a quiet year. Nothing to wrap!\n", username)
		os.Exit(0)
	}

	// Compute stats
	stats := github.ComputeStats(user, events, repos, calendar)

	// Compute personality
	p := personality.Compute(stats)
	stats.Archetype = p.Archetype
	stats.Traits = p.Traits

	// Run TUI
	model := ui.NewModel(stats, auto)
	program := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := program.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Check if GIF export was requested
	if m, ok := finalModel.(ui.Model); ok && m.ExportRequested() {
		if err := export.GenerateGIF(username); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	}
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build -o gh-wrapped .`
Expected: binary created successfully

- [ ] **Step 3: Quick smoke test**

Run: `./gh-wrapped torvalds`
Expected: TUI launches with slides (may have visual bugs — that's fine, the wiring works)

- [ ] **Step 4: Commit**

```bash
git add main.go
git commit -m "feat: wire all components in main.go with concurrent data fetching"
```

---

### Task 12: Polish & Integration Test

**Files:**
- Verify all components work end-to-end
- Fix any compilation or runtime issues

- [ ] **Step 1: Run all tests**

Run: `go test ./... -v`
Expected: all PASS

- [ ] **Step 2: Run the tool against a real user**

Run: `go run . octocat`
Expected: TUI launches, slides render, auto-advance works

- [ ] **Step 3: Test auto mode**

Run: `go run . octocat --auto`
Expected: slides play automatically, program exits after last slide

- [ ] **Step 4: Test edge case — non-existent user**

Run: `go run . thisuserdoesnotexist123456`
Expected: "Error: user: not found" and exit 1

- [ ] **Step 5: Run go vet and fix any warnings**

Run: `go vet ./...`
Expected: no warnings

- [ ] **Step 6: Final commit**

```bash
git add -A
git commit -m "fix: integration fixes and polish"
```
