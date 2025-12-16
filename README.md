# sesn

A tmux session manager with a beautiful TUI.

## Installation

### Easy Install (Recommended)

Clone the repository and run the install script:

```bash
git clone https://github.com/thetanav/sesn.git
cd sesn
./install.sh
```

This will:

- Check for required dependencies (Go, tmux, git)
- Build the binary from source
- Install it to `~/bin` (or `/usr/local/bin` if run as root)

### Custom Install Location

If you want to install to a different location:

```bash
git clone https://github.com/thetanav/sesn.git
cd sesn
export INSTALL_DIR=/path/to/install
./install.sh
```

Note: Ensure the install directory is in your PATH.

### Building from source

If you prefer to build manually:

```bash
git clone https://github.com/thetanav/sesn.git
cd sesn
go build -o sesn .
```

Then move `sesn` to your PATH (e.g., `sudo mv sesn /usr/local/bin/`)

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
