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
