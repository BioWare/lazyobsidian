# LazyObsidian

TUI productivity dashboard for Obsidian vault.

## Features

- **Keyboard-first** - Full vim-style navigation (j/k/h/l/gg/G)
- **Vault = Source of Truth** - All data stored in .md files
- **Pomodoro Timer** - Background daemon with Neovim statusline integration
- **Goals & Tasks** - Hierarchical goal tracking with progress visualization
- **Courses & Books** - Track learning progress
- **Statistics** - Activity heatmaps and focus time analytics

## Installation

```bash
go install github.com/BioWare/lazyobsidian/cmd/lazyobsidian@latest
```

## Usage

```bash
lazyobsidian --vault ~/obsidian/my-vault
```

## Configuration

Config file: `~/.config/lazyobsidian/config.yaml`

```yaml
vault:
  path: ~/obsidian/my-vault

folders:
  daily: Journal
  goals: Plan
  courses: Input/Courses
  books: Input/Books

pomodoro:
  work_minutes: 25
  short_break: 5
  daily_goal: 5
```

## Keybindings

| Key | Action |
|-----|--------|
| `j/k` | Navigate up/down |
| `h/l` | Collapse/Expand |
| `Tab` | Switch panels |
| `Enter` | Select/Action |
| `p` | Start Pomodoro |
| `/` | Global search |
| `?` | Help |
| `q` | Quit |

## Related Projects

- [lazyobsidian.nvim](https://github.com/BioWare/lazyobsidian.nvim) - Neovim plugin

## License

MIT
