# sesn

A tmux session manager with a beautiful TUI.

## Installation

### Easy Install (Recommended)

Run this one-liner to install sesn:

```bash
curl -fsSL https://raw.githubusercontent.com/thetanav/sesn/main/install.sh | bash
```

This will:

- Check for required dependencies (Go, tmux, git)
- Clone the repository
- Build the binary
- Install it to `/usr/local/bin` (requires sudo)

### Custom Install Location

If you want to install to a different location (e.g., user directory):

```bash
export INSTALL_DIR=$HOME/bin
curl -fsSL https://raw.githubusercontent.com/thetanav/sesn/main/install.sh | bash
```

### Building from source

If you prefer to build manually:

```bash
git clone https://github.com/thetanav/sesn.git
cd sesn
go build -o sesn .
```

## Usage

```bash
sesn              # Launch the TUI
sesn -f           # Use fuzzy finder mode
```

## Keybindings

- `c` - Create new session
- `d` - Delete selected session
- `r` - Rename selected session
- `k` - Kill selected session
- `enter` - Attach to selected session
- `/` - Fuzzy find mode
- `ctrl+c` - Quit

## Requirements

- Go 1.19+
- tmux
- git (for installation script)
