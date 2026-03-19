// github/stats.go
package github

import (
	"sort"
	"time"
)

// GitHub language colors — top languages
var LanguageColors = map[string]string{
	"Go": "#00ADD8", "JavaScript": "#f1e05a", "TypeScript": "#3178c6",
	"Python": "#3572A5", "Rust": "#dea584", "Java": "#b07219",
	"C": "#555555", "C++": "#f34b7d", "C#": "#178600",
	"Ruby": "#701516", "PHP": "#4F5D95", "Swift": "#F05138",
	"Kotlin": "#A97BFF", "Shell": "#89e051", "HTML": "#e34c26",
	"CSS": "#563d7c", "Dart": "#00B4AB", "Lua": "#000080",
	"Zig": "#ec915c", "Elixir": "#6e4a7e", "Haskell": "#5e5086",
	"Scala": "#c22d40", "R": "#198CE7", "Vue": "#41b883", "Svelte": "#ff3e00",
}

func ComputeStats(user User, events []Event, repos []Repo, calendar []ContributionDay) Stats {
	s := Stats{
		Username:    user.Login,
		Name:        user.Name,
		HasCalendar: len(calendar) > 0,
	}
	s.YearLabel = computeYearLabel(calendar)
	s.TotalRepos = len(repos)
	for _, r := range repos {
		s.TotalStars += r.StargazersCount
	}
	if len(calendar) > 0 {
		s.Calendar = calendar
		computeCalendarStats(&s, calendar)
	}
	if len(events) > 0 {
		computeEventStats(&s, events)
	}
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
	evening := blocks[2] + blocks[3]
	morning := blocks[0]
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
			Name: name, Count: count,
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
