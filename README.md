# todo-calendar

A terminal-based (TUI) application that combines a monthly calendar view with a todo list. The left panel shows a navigable calendar with national holidays highlighted in red. The right panel displays todos for the visible month alongside undated (floating) items.

Built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Install

```
go install github.com/anttilinno/todo-calendar@latest
```

Or build from source:

```
git clone git@github.com:anttilinno/todo-calendar.git
cd todo-calendar
go build -o todo-calendar .
```

## Usage

```
./todo-calendar
```

### Keybindings

**General**

| Key | Action |
|-----|--------|
| `Tab` | Switch between calendar and todo panes |
| `q` / `Ctrl+C` | Quit |

**Calendar (left pane)**

| Key | Action |
|-----|--------|
| `h` / `Left` | Previous month |
| `l` / `Right` | Next month |

**Todo list (right pane)**

| Key | Action |
|-----|--------|
| `j` / `Down` | Move cursor down |
| `k` / `Up` | Move cursor up |
| `a` | Add floating todo |
| `A` | Add dated todo |
| `x` | Toggle complete |
| `d` | Delete todo |
| `Enter` | Confirm input |
| `Esc` | Cancel input |

## Configuration

Create `~/.config/todo-calendar/config.toml`:

```toml
country = "fi"
monday_start = true
```

| Option | Default | Description |
|--------|---------|-------------|
| `country` | `"us"` | Country code for national holidays |
| `monday_start` | `false` | Start week on Monday instead of Sunday |

### Supported countries

`de` `dk` `ee` `es` `fi` `fr` `gb` `it` `no` `se` `us`

## Data storage

Todos are stored as JSON at `~/.config/todo-calendar/todos.json`.

## License

MIT
