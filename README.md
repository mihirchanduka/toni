# toni

A clean, local-first TUI for logging restaurant visits. Single-user, fully offline, zero cloud—inspired by the Beli app.

## Features

- **Local-first**: All data stored in a single SQLite file (`~/.toni/toni.db`)
- **Restaurant Autocomplete**: Powered by Yelp Fusion API (optional, works offline without it)
- **Vim-style navigation**: Modal interface with familiar keybindings
- **Fast keyboard workflow**: Navigate, search, and add entries without touching the mouse
- **Zero dependencies**: Pure Go, no CGO, no external services
- **Beautiful TUI**: Clean design with polished tables, human-friendly dates, and color-coded ratings

## Installation

### From Source

```bash
git clone <repo-url>
cd toni
go install
```

### Run Directly

```bash
go run main.go
go run main.go --db ~/my-food.db
```

## Usage

### Database Location

By default, toni stores your data in `~/.toni/toni.db`. You can override this with the `--db` flag:

```bash
toni --db /path/to/your/database.db
```

To back up your data, simply copy the SQLite file:

```bash
cp ~/.toni/toni.db ~/backups/toni-backup.db
```

### Restaurant Autocomplete

toni integrates with the Yelp Fusion API to provide smart restaurant autocomplete when adding visits. This is **completely optional** — the app works perfectly offline without it.

#### Setup

1. Open the Yelp developer dashboard at https://www.yelp.com/developers/v3/manage_app
2. Create an API key at https://www.yelp.com/developers/v3/manage_app
3. Set your API key via the onboarding process or in `~/.toni`

#### How it Works

When entering a restaurant name in the visit form:
- Type 2+ characters to trigger autocomplete
- Results appear in a dropdown below the input field
- Navigate with `j/k` (or `↑/↓`)
- Press `Enter` or `Tab` to select a suggestion
- Press `Esc` to dismiss the dropdown

The autocomplete provides:
- Restaurant name
- City and neighborhood
- Cuisine type (auto-filled from Yelp categories)

**Free Tier**: 10,000 API calls per month, no credit card required.

**Offline Mode**: If no API key is set, toni displays a subtle startup message and the restaurant field works as a plain text input.

## Keybindings

### Navigation Mode (Default)

#### Global Movement
| Key        | Action              |
|------------|---------------------|
| j / ↓      | Move down           |
| k / ↑      | Move up             |
| h / ← / b  | Go back / parent    |
| l / → / enter | Open / select    |
| gg         | Jump to top         |
| G          | Jump to bottom      |
| ctrl+d     | Half page down      |
| ctrl+u     | Half page up        |
| /          | Search/filter       |
| esc        | Cancel / close      |
| q          | Quit                |
| ?          | Toggle help         |

#### Visits Screen (Home)
| Key   | Action            |
|-------|-------------------|
| a     | Quick-add visit   |
| r     | Go to restaurants |
| enter | Open visit detail |

#### Restaurants Screen
| Key   | Action                  |
|-------|-------------------------|
| a     | Add restaurant          |
| v     | Log visit for selected  |
| enter | Open restaurant detail  |
| b / h | Back to visits          |

#### Detail Screens
| Key      | Action     |
|----------|------------|
| h / esc  | Back       |
| e        | Edit       |
| d        | Delete     |
| v        | Add visit (restaurants only) |

### Insert/Edit Mode (Forms)

| Key         | Action         |
|-------------|----------------|
| tab         | Next field     |
| shift+tab   | Previous field |
| ctrl+s      | Save           |
| esc         | Cancel         |

**Autocomplete Dropdown** (when active in restaurant field):
- `j/k` or `↓/↑` to navigate suggestions
- `enter` or `tab` to select
- `esc` to dismiss

## Data Model

### Restaurants
- Name (required)
- City
- Neighborhood
- Cuisine
- Price Range ($, $$, $$$, $$$$)

### Visits
- Restaurant (required)
- Date (YYYY-MM-DD, defaults to today)
- Rating (1-10 scale)
- Would Return? (Yes/No)
- Notes (free text)

## Architecture

Built with a clean separation of concerns:

- `internal/db/` - Database layer with typed queries
- `internal/model/` - Domain types and Bubble Tea messages
- `internal/ui/` - TUI components and screen logic
- `internal/util/` - Formatting and validation utilities
- `cmd/` - CLI flag parsing

## Tech Stack

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Styling
- [modernc.org/sqlite](https://modernc.org/sqlite) - Pure Go SQLite driver
- [micasa](https://micasa.dev/) - Original Inspiration
## License

MIT
