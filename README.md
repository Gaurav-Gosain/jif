# JIF - Terminal GIF Viewer

A modern, high-performance GIF viewer for your terminal built with Bubbletea and Lipgloss.

## Installation

### Homebrew (macOS/Linux)

```bash
brew tap Gaurav-Gosain/tap
brew install jif
```

### Arch Linux (AUR)

```bash
yay -S jif-bin
# or
paru -S jif-bin
```

### Go Install

```bash
go install github.com/Gaurav-Gosain/jif/cmd/jif@latest
```

### Build from Source

```bash
git clone https://github.com/Gaurav-Gosain/jif
cd jif
make build
./bin/jif animation.gif
```

## Usage

```bash
# View a local GIF
jif animation.gif

# View a remote GIF
jif https://example.com/animation.gif

# Show help
jif --help

# Show version
jif --version
```

## Keybindings

| Key            | Action         |
| -------------- | -------------- |
| `Space`        | Pause/Resume   |
| `n` / `→`      | Next frame     |
| `p` / `←`      | Previous frame |
| `?`            | Toggle help    |
| `q` / `Ctrl+C` | Quit           |

Press `?` while viewing to see the help overlay.

## Features

- Halfblock rendering for 2x vertical resolution
- High-quality Lanczos3 image scaling
- Progressive loading animation
- Proper GIF disposal method handling
- Remote URL support (HTTP/HTTPS)
- Automatic terminal resize handling
- Full-screen alternate buffer mode

## Technical Details

### Rendering

Uses Unicode halfblock characters (▀ and ▄) to achieve double vertical
resolution. Each terminal cell displays two pixels vertically:

- Top pixel: foreground color
- Bottom pixel: background color

### GIF Support

Properly handles all GIF disposal methods:

- DisposalNone (0): Keep previous frame
- DisposalBackground (1): Clear to background
- DisposalPrevious (2): Restore previous state

This ensures accurate rendering of complex animated GIFs.

## Development

### Run Tests

```bash
make test
```

### Build Binary

```bash
make build
```

### Generate Test Data

```bash
make generate-testdata
```

## License

MIT License - see LICENSE file for details

## Author

[@Gaurav-Gosain](https://github.com/Gaurav-Gosain)
