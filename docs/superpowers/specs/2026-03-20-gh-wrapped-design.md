# gh-wrapped — Design Spec

> Spotify Wrapped, but for your GitHub year.

## Overview

A Go CLI tool that fetches a GitHub user's public activity and presents it as a colorful, animated terminal slideshow using Charm libraries (Bubble Tea, Lip Gloss, Bubbles). Optionally exports the presentation as a GIF via Charm's VHS.

## CLI Interface

```
gh-wrapped <username>              # wrap a GitHub user's year
gh-wrapped <username> --auto       # headless auto-play (used by VHS for GIF export)
```

No subcommands, no config files. One positional arg. Always wraps the last 12 months (rolling window based on contribution calendar).

## Data Source

GitHub public API only. No authentication required.

**Endpoints:**
- `GET /users/{username}` — profile info
- `GET /users/{username}/events/public` — recent public events (paginated, max 300 events, **30-day window**)
- `GET /users/{username}/repos` — repos + language breakdown
- **HTML scrape:** `https://github.com/users/{username}/contributions` — contribution calendar for the last 12 months. Returns an SVG with `<rect>` elements containing `data-date`, `data-level` (0-4), and `data-count` attributes. This is the only unauthenticated way to get full-year contribution data — GitHub's GraphQL API requires a token. **If scraping fails** (GitHub changes HTML structure), the tool enters degraded mode: shows only slides sourced from the REST APIs and skips calendar-dependent slides (2, 3, 4, 8).

**Data coverage per slide:**
| Data Source | Window | Used By |
|---|---|---|
| Contribution calendar (HTML) | 12 months | Slides 2 (total contributions), 3 (heatmap), 4 (chaotic day), 8 (weekend %) |
| Events API | 30 days | Slides 5 (commit times), 7 (most-pushed repo), 9 (commit messages) |
| Repos API | All time | Slides 2 (repos count, total stars), 6 (languages) |
| User API | Current | Slide 1 (profile) |

Slides sourced from the events API (30-day window) display "Based on your last 30 days" as a subtle footnote.

**Rate limits:** 60 requests/hour unauthenticated. We need ~10-20 requests per user (events API is paginated at up to 100/page, repos can span multiple pages). Running the tool more than 3x in quick succession may hit the limit. If `GITHUB_TOKEN` is set in the environment, we silently use it for higher rate limits (5,000/hr) and private repo access — but we never prompt for it.

## Presentation Flow

Hybrid auto-play with manual controls:
- Slides auto-advance every 3-4 seconds with transitions
- Arrow keys take manual control at any time
- `a` resumes auto-play
- `q` / `Esc` quits
- `g` on the final slide triggers GIF export

## Slides (10 total)

### Slide 1: Title Card — "The Opening"
Big animated "DEV WRAPPED" title with gradient text and the year range (e.g., "2025-2026" if spanning two calendar years, or just "2025" if Jan-Dec). Username fades in below. Loading spinner plays while data fetches from GitHub API concurrently (all endpoints in parallel via errgroup).

### Slide 2: Your Year in Numbers — "The Overview"
Animated counters tick up from zero: total contributions (from calendar), repos contributed to (repos API), total stars across repos (sum of `stargazers_count` from repos API). Numbers land with a snap. Only shows metrics we can actually fetch without auth — no PRs/issues/reviews counts (those require GraphQL).

### Slide 3: Contribution Heatmap — "The Map"
Terminal-rendered contribution graph using block characters (░▓█). Cells fill in left-to-right with animation. Longest streak gets highlighted with a callout.

### Slide 4: Most Chaotic Day — "The Chaos"
The day with the most contributions. Shows the date, the count in a huge number, and a snarky one-liner ("were you okay?", "deadline energy detected").

### Slide 5: Night Owl or Early Bird? — "The Clock"
Commit time distribution across 4 time blocks (06-12, 12-18, 18-00, 00-06) rendered as horizontal bars. Crowns you "Night Owl", "Early Bird", or "9-to-5er" based on distribution. Data sourced from events API (30-day window). Footnote: "Based on your last 30 days."

### Slide 6: Top Languages — "The Stack"
Languages ranked by number of repos using that language as primary (the `language` field from the repos API — no extra API calls needed). Each bar in its actual GitHub language color. Fun subtitle based on distribution ("Go purist", "Polyglot energy", etc.).

### Slide 7: Villain Arc — "The Villain Arc"
The repo with the most `PushEvent` commits in the events API (30-day window), framed as your obsession. "You pushed 47 commits to legacy-api in the last month. Obsessed much?" Menacing red vibe. Skipped if no PushEvents exist in the events window.

### Slide 8: Weekend Warrior Score — "The Schedule"
Percentage of contributions on weekends. Split bar visualization. Snarky verdict: >30% = "work-life balance is a myth", <10% = "touch grass? already on it."

### Slide 9: Longest Commit Message — "The Novel"
The actual commit message (truncated if needed), displayed with typewriter effect. Shows the repo and character count. "That's not a commit message, that's a blog post." Data sourced from `PushEvent` payloads in the events API (30-day window).

### Slide 10: Developer Personality — "The Finale"
Grand finale. An archetype derived from all stats displayed in big gradient text. Examples: "The Nightcrawler" (late-night coder), "The Obsessed" (one-repo focus), "The Machine" (long streaks), "The Sprinter" (burst activity). Shows top 3 traits as pill badges. Ends with "Your Year, Unwrapped." and a prompt to press `g` for GIF export.

**Slide skipping:** If a user has insufficient data for a slide (e.g., no PushEvents for Villain Arc, scraper failure for calendar slides), that slide is skipped entirely rather than showing boring zeros.

## Personality Archetypes

Based on stat analysis, assign one of:
- **The Nightcrawler** — >50% of commits after 6pm (events API)
- **The Obsessed** — >60% of PushEvents to a single repo (events API)
- **The Machine** — longest streak >= 14 days and < 5 zero-contribution days per month (calendar)
- **The Weekender** — weekend contribution % > 30% (calendar)
- **The Novelist** — longest commit message > 500 characters (events API)
- **The Polyglot** — 4+ languages, none > 40% (repos API)
- **The Specialist** — one dominant language > 70% (repos API)
- **The Sprinter** — busiest day has > 5x the daily average (calendar)

Archetype selection priority: score each archetype 0-1 based on how strongly the user matches, pick the highest. Ties broken by the order above.

Top 3 traits (from the same pool) are shown as pill badges — these are the top 3 scoring archetypes regardless of which one is primary.

## GIF Export

On pressing `g` at the final slide:
1. Generate a VHS `.tape` file that runs `gh-wrapped <username> --auto`
2. The tape configures: 120x40 terminal, JetBrains Mono font, dark background
3. Shell out to `vhs` to render the GIF
4. Print the output path on completion

**`--auto` mode:** Plays all slides with 3-second auto-advance (30s total GIF for 10 slides), no keyboard interaction, exits cleanly after the last slide. Animations complete within the 3-second window. This is what VHS records.

**VHS dependency:** VHS must be installed separately. If not found on `g` press, print: "VHS not found. Install with: brew install vhs" — no crash.

## Architecture

```
gh-wrapped/
├── main.go              # CLI entry point, arg parsing
├── github/
│   ├── client.go        # GitHub REST API client (unauthenticated, respects GITHUB_TOKEN)
│   ├── scraper.go       # HTML scraper for contribution calendar
│   ├── models.go        # API response types
│   └── stats.go         # Fetch + compute all stats from raw API data
├── stats/
│   └── personality.go   # Archetype engine — maps stats → personality
├── ui/
│   ├── app.go           # Bubble Tea main model (slide state machine)
│   ├── slides.go        # Slide definitions + per-slide render logic
│   ├── theme.go         # Lip Gloss styles, colors, gradients
│   └── animation.go     # Tick-based animations (counters, heatmap fill)
└── export/
    └── gif.go           # VHS tape generation + execution
```

**Minimum terminal size:** 100x30. If the terminal is smaller, print a warning and exit: "Terminal too small. Please resize to at least 100x30." Bubble Tea's `tea.WindowSizeMsg` is used to detect this.

**Data flow:**
1. `main.go` parses args, calls `github/` to fetch all data concurrently (errgroup — profile, events, repos, contribution scrape in parallel)
2. `github/stats.go` computes all 10 slide stats from raw API responses
3. Stats struct gets passed to `ui/app.go` which boots Bubble Tea
4. Bubble Tea runs the slide state machine — auto-advances on timer, responds to keypresses
5. On `g` press, `export/gif.go` generates a VHS `.tape` file and shells out to `vhs`

**Key dependencies:**
- `github.com/charmbracelet/bubbletea` — TUI framework
- `github.com/charmbracelet/lipgloss` — styling
- `github.com/charmbracelet/bubbles` — spinners, progress bars
- `github.com/charmbracelet/vhs` — GIF export (runtime dependency, not compiled in)

## Error Handling

- **User not found:** "No GitHub user found for @typo123" — exit 1
- **Zero public activity:** "Looks like @ghost had a quiet year. Nothing to wrap!" — exit 0
- **Sparse data:** Skip slides that lack meaningful data
- **Rate limited:** "GitHub rate limit reached. Try again in a few minutes, or set GITHUB_TOKEN for higher limits." — exit 1
- **No VHS installed:** On `g` press: "VHS not found. Install with: brew install vhs" — no crash, presentation continues
- **Scraper failure:** Contribution calendar HTML changed or unreachable — enter degraded mode, skip calendar-dependent slides (2, 3, 4, 8), show remaining slides
- **Terminal too small:** Print "Terminal too small. Please resize to at least 100x30." and exit
