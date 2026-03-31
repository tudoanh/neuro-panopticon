# Ralph Agent Configuration

## Build Instructions

```bash
# Ensure Go is in PATH
export PATH=$HOME/.local/go/bin:$HOME/go/bin:$PATH

# Build Go packages (compiles backend, no GUI deps needed)
go build ./...

# Full Wails build (requires libgtk-3-dev and libwebkit2gtk-4.0-dev)
# wails build
```

## Test Instructions

```bash
export PATH=$HOME/.local/go/bin:$HOME/go/bin:$PATH
go test ./...
```

## Run Instructions

```bash
export PATH=$HOME/.local/go/bin:$HOME/go/bin:$PATH

# Development mode (requires system GUI deps)
# wails dev

# Or just verify compilation
go build ./...
```

## Notes
- Go 1.24.2 installed at ~/.local/go (no sudo available)
- Wails v2.12.0 installed at ~/go/bin/wails
- Missing system deps for GUI: libgtk-3-dev, libwebkit2gtk-4.0-dev
- picoclaw v0.2.4 used for tool/provider abstractions
- gopsutil v4 used for system monitoring
