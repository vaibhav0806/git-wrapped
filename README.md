# gh-wrapped

Spotify Wrapped, but for your GitHub year. A beautiful terminal slideshow of your GitHub activity.

```
 ██████╗ ██╗████████╗
██╔════╝ ██║╚══██╔══╝
██║  ███╗██║   ██║
██║   ██║██║   ██║
╚██████╔╝██║   ██║
 ╚═════╝ ╚═╝   ╚═╝
██╗    ██╗██████╗  █████╗ ██████╗ ██████╗ ███████╗██████╗
██║    ██║██╔══██╗██╔══██╗██╔══██╗██╔══██╗██╔════╝██╔══██╗
██║ █╗ ██║██████╔╝███████║██████╔╝██████╔╝█████╗  ██║  ██║
██║███╗██║██╔══██╗██╔══██║██╔═══╝ ██╔═══╝ ██╔══╝  ██║  ██║
╚███╔███╔╝██║  ██║██║  ██║██║     ██║     ███████╗██████╔╝
 ╚══╝╚══╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝     ╚═╝     ╚══════╝╚═════╝
```

## What is this?

Run `git-wrapped <username>` and get a cinematic, animated terminal presentation of any GitHub user's year — contributions, streaks, languages, coding habits, and a personality archetype. No authentication required.

## Install

**Homebrew** (macOS/Linux):

```bash
brew install vaibhav0806/tap/git-wrapped
```

**Go**:

```bash
go install github.com/vaibhav0806/git-wrapped@latest
```

**Binary**: download from [Releases](https://github.com/vaibhav0806/git-wrapped/releases)

**Build from source**:

```bash
git clone https://github.com/vaibhav0806/git-wrapped.git
cd git-wrapped
go build -o git-wrapped .
```

## Usage

```bash
git-wrapped <username>        # wrap any GitHub user
git-wrapped <username> --auto # auto-play mode (no keyboard)
```

Slides auto-advance. Use arrow keys to navigate manually, `a` to resume auto-play, `q` to quit.

Press `g` on the last slide to export as GIF (requires [VHS](https://github.com/charmbracelet/vhs)).

## The Slides

| # | Slide | What it shows |
|---|-------|---------------|
| 1 | **Title** | GIT WRAPPED with ASCII art + your username |
| 2 | **Year in Numbers** | Total contributions, repos, stars |
| 3 | **Contribution Heatmap** | GitHub-style green grid with streak callout |
| 4 | **Most Chaotic Day** | Your busiest day with a snarky comment |
| 5 | **When You Code** | Commit time distribution — Night Owl, Early Bird, or 9-to-5er |
| 6 | **Top Languages** | Ranked bars with actual GitHub language colors |
| 7 | **Villain Arc** | Your most-pushed repo — "obsessed much?" |
| 8 | **Weekend Warrior** | What % of your commits land on weekends |
| 9 | **The Novel** | Your longest commit message |
| 10 | **Personality** | Lottery reveal of your developer archetype |

## Personality Archetypes

Based on your stats, you get assigned one of:

- **The Nightcrawler** — you come alive after dark
- **The Machine** — relentless consistency, long streaks, no days off
- **The Obsessed** — one repo owns your soul
- **The Sprinter** — calm, calm, calm, then BOOM
- **The Specialist** — one language to rule them all
- **The Polyglot** — many languages, no loyalty
- **The Weekender** — weekends aren't for rest
- **The Novelist** — your commit messages read like short stories

## Data Sources

All data comes from GitHub's public API. No authentication needed.

| Source | What it provides |
|--------|-----------------|
| REST API `/users/{user}` | Profile info |
| REST API `/users/{user}/events/public` | Recent events (30-day window) |
| REST API `/users/{user}/repos` | Repos + languages |
| HTML scrape of contribution calendar | Full-year contribution data |

Set `GITHUB_TOKEN` env var for higher rate limits (optional).

## GIF Export

Requires [VHS](https://github.com/charmbracelet/vhs) from Charm:

```bash
brew install vhs
```

Press `g` on the personality slide to generate a GIF of the full presentation.

## Built With

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — Terminal styling
- [Bubbles](https://github.com/charmbracelet/bubbles) — TUI components
- [VHS](https://github.com/charmbracelet/vhs) — Terminal GIF recorder

## License

MIT
