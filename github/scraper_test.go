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
<td id="contribution-day-component-0-0" data-date="2025-01-01" data-level="2" class="ContributionCalendar-day"></td>
<td id="contribution-day-component-0-1" data-date="2025-01-02" data-level="0" class="ContributionCalendar-day"></td>
<td id="contribution-day-component-0-2" data-date="2025-01-03" data-level="4" class="ContributionCalendar-day"></td>
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
