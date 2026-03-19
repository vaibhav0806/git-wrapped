# gh-wrapped — Design Spec

> Spotify Wrapped, but for your GitHub year.

## Overview

A Go CLI tool that fetches a GitHub user's public activity and presents it as a colorful, animated terminal slideshow using Charm libraries (Bubble Tea, Lip Gloss, Bubbles). Optionally exports the presentation as a GIF via Charm's VHS.

## CLI Interface

```
gh-wrapped <username>              # wrap a GitHub user's year
gh-wrapped <username> --year 2024  # specific year (default: current)
gh-wrapped --self                  # uses git config user
gh-wrapped <username> --auto       # headless auto-play (used by VHS for GIF export)
```

No subcommands, no config files. One positional arg.

## Data Source

GitHub public API only. No authentication required.

**Endpoints:**
- `GET /users/{username}` — profile info
- `GET /users/{username}/events/public` — recent public events (paginated, max 300 events / 90 days)
- `GET /users/{username}/repos` — repos + language breakdown
- GraphQL contribution calendar — full year heatmap (public, no auth)

**Rate limits:** 60 requests/hour unauthenticated. We need ~5-10 requests per user, well within limits. If `GITHUB_TOKEN` is set in the environment, we silently use it for higher rate limits (5,000/hr) and private repo access — but we never prompt for it.

## Presentation Flow

Hybrid auto-play with manual controls:
- Slides auto-advance every 3-4 seconds with transitions
- Arrow keys take manual control at any time
- `a` resumes auto-play
- `q` / `Esc` quits
- `g` on the final slide triggers GIF export

## Slides (10 total)

### Slide 1: Title Card — "The Opening"
Big animated "DEV WRAPPED 2025" title with gradient text. Username fades in below. Loading spinner plays while data fetches from GitHub API.

### Slide 2: Your Year in Numbers — "The Overview"
Animated counters tick up from zero: total commits, PRs merged, issues closed, PRs reviewed, repos contributed to, stars earned. Numbers land with a snap.

### Slide 3: Contribution Heatmap — "The Map"
Terminal-rendered contribution graph using block characters (░▓█). Cells fill in left-to-right with animation. Longest streak gets highlighted with a callout.

### Slide 4: Most Chaotic Day — "The Chaos"
The day with the most contributions. Shows the date, the count in a huge number, and a snarky one-liner ("were you okay?", "deadline energy detected").

### Slide 5: Night Owl or Early Bird? — "The Clock"
Commit time distribution across 4 time blocks (06-12, 12-18, 18-00, 00-06) rendered as horizontal bars. Crowns you "Night Owl", "Early Bird", or "9-to-5er" based on distribution.

### Slide 6: Top Languages — "The Stack"
Languages ranked by repo activity, each bar in its actual GitHub language color. Fun subtitle based on distribution ("Go purist", "Polyglot energy", etc.).

### Slide 7: Villain Arc — "The Villain Arc"
The repo where the user deleted the most code. Dramatic red number, menacing vibe. "You mass-deleted 4,200 lines from legacy-api. No regrets?"

### Slide 8: Weekend Warrior Score — "The Schedule"
Percentage of contributions on weekends. Split bar visualization. Snarky verdict: >30% = "work-life balance is a myth", <10% = "touch grass? already on it."

### Slide 9: Longest Commit Message — "The Novel"
The actual commit message (truncated if needed), displayed with typewriter effect. Shows the repo and character count. "That's not a commit message, that's a blog post."

### Slide 10: Developer Personality — "The Finale"
Grand finale. An archetype derived from all stats displayed in big gradient text. Examples: "The Nightcrawler" (late-night coder), "The Destroyer" (massive deletions), "The Machine" (long streaks), "The Socialite" (lots of PRs/reviews). Shows top 3 traits as pill badges. Ends with "Your 2025, Unwrapped." and a prompt to press `g` for GIF export.

**Slide skipping:** If a user has insufficient data for a slide (e.g., no deletions for Villain Arc), that slide is skipped entirely rather than showing boring zeros.

## Personality Archetypes

Based on stat analysis, assign one of:
- **The Nightcrawler** — majority of commits after 6pm
- **The Destroyer** — heavy deletion activity
- **The Machine** — longest streaks, high consistency
- **The Socialite** — lots of PRs and reviews
- **The Weekender** — high weekend warrior score
- **The Novelist** — unusually long commit messages
- **The Polyglot** — many languages, no strong majority
- **The Specialist** — one dominant language >70%
- **The Sprinter** — high burst activity, chaotic days

Top 3 traits (from the same pool) are shown as pill badges.

## GIF Export

On pressing `g` at the final slide:
1. Generate a VHS `.tape` file that runs `gh-wrapped <username> --auto`
2. The tape configures terminal dimensions, font, and theme
3. Shell out to `vhs` to render the GIF
4. Print the output path on completion

**`--auto` mode:** Plays all slides with auto-advance, no keyboard interaction, exits cleanly after the last slide. This is what VHS records.

**VHS dependency:** VHS must be installed separately. If not found on `g` press, print: "VHS not found. Install with: brew install vhs" — no crash.

## Architecture

```
gh-wrapped/
├── main.go              # CLI entry point, arg parsing
├── github/
│   ├── client.go        # GitHub API client (unauthenticated, respects GITHUB_TOKEN)
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

**Data flow:**
1. `main.go` parses args, calls `github/` to fetch data
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
