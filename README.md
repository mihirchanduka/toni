# toni

A clean, local-first TUI for logging restaurant visits. Single-user, fully offline, zero cloud—inspired by the Beli app.

## Features

- **Local-first:** All data stored in a single SQLite file (`~/.toni/toni.db`).
- **Restaurant Autocomplete:** Powered by Yelp Fusion API (optional, works perfectly offline without it).
- **Vim-style navigation:** Modal interface with familiar keybindings (e.g., `h`, `j`, `k`, `l`).
- **Fast keyboard workflow:** Navigate, search, and add entries without touching the mouse.
- **Zero dependencies:** Pure Go, no CGO, no external services required.
- **Beautiful TUI:** Clean design with polished tables, human-friendly dates, and color-coded ratings.

## Prerequisites

The project is built in pure Go and requires Go 1.22 or higher.

## Build & Install

From the project root, you can install the binary directly using Go:

```bash
git clone https://github.com/github_username/toni.git
cd toni
go install
```

Alternatively, you can run it directly from the source:

```bash
go run main.go
```

## Run

Execute the compiled binary from your terminal:

```bash
toni
```

### Usage Instructions
1. **Database Location:** By default, toni stores your data in `~/.toni/toni.db`. You can override this using the `--db` flag (e.g., `toni --db ~/my-food.db`).
2. **Yelp Autocomplete (Optional):** If you want smart restaurant autocomplete, create a Yelp API key at the [Yelp Developer Dashboard](https://www.yelp.com/developers/v3/manage_app). You can set it during toni's first-run onboarding or via the `YELP_API_KEY` environment variable.
3. **Adding a Visit:** Press `a` from the home screen. If Yelp is configured, type 2+ characters to trigger autocomplete and use `j/k` to select a suggestion.
4. **Navigation:** Use `j/k` (or arrows) to move up and down, `enter` to view details, and `esc` or `h` to go back. Press `?` at any time to toggle the help menu.

## Data Model

toni's local database captures:

- **Restaurants:** Name (required), City, Neighborhood, Cuisine, and Price Range ($, $$, $$$, $$$$).
- **Visits:** Restaurant (required), Date, Rating (1-10 scale), Would Return? (Yes/No), and Notes.

## Keybindings

**Global Movement:**
- `j` / `↓` : Move down
- `k` / `↑` : Move up
- `h` / `←` / `b` : Go back / parent
- `l` / `→` / `enter` : Open / select
- `/` : Search/filter
- `esc` : Cancel / close
- `q` : Quit
- `?` : Toggle help

**Actions (Visits & Restaurants Screens):**
- `a` : Quick-add visit / Add restaurant
- `r` : Go to restaurants view
- `v` : Log visit for selected restaurant
- `e` : Edit entry (in Detail View)
- `d` : Delete entry (in Detail View)

## Notes

- **Backups:** To back up your data, simply copy the SQLite file: `cp ~/.toni/toni.db ~/backups/toni-backup.db`.
- **Architecture:** Built with Bubble Tea, Bubbles, Lip Gloss, and modernc.org/sqlite (pure Go SQLite driver) with a clean separation of concerns (`db`, `model`, `ui`).
- **Acknowledgments:** Original inspiration from [micasa](https://micasa.dev/).
