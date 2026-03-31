# NeuroPanopticon Implementation Plan

## Phase 1: Project Initialization [COMPLETE]
- [x] Install Go 1.24.2 (user-local at ~/.local/go)
- [x] Install Wails v2.12.0
- [x] Initialize Wails project with React+TypeScript template
- [x] Create backend directory structure (agent/, skills/, scanner/, models/, config/)
- [x] Create data models (`backend/models/types.go`)
- [x] Create config package (`backend/config/config.go`)
- [x] Create LLM provider wrapper (`backend/agent/provider.go`) - supports Ollama + Cloud
- [x] Create agent loop (`backend/agent/agent.go`) - picoclaw-based tool calling
- [x] Implement all 5 security skills:
  - [x] `scan_network` - port/connection scanner using gopsutil
  - [x] `detect_lateral_movement` - process tree anomaly detection
  - [x] `analyze_blast_radius` - network environment risk analysis
  - [x] `inspect_sbom` - shared library/dependency extraction
  - [x] `remediate` - safe OS command execution (firewall, kill, service)
- [x] Wire agent into Wails app.go with IPC methods
- [x] All Go packages compile successfully

### System Dependencies Note
Missing system packages (need sudo): `libgtk-3-dev`, `libwebkit2gtk-4.0-dev`
These are required for `wails build` but not for Go compilation.
Install with: `sudo apt install libgtk-3-dev libwebkit2gtk-4.0-dev`

## Phase 2: Backend Enhancement [COMPLETE]
- [x] Add background scanner (`backend/scanner/`) - periodic system monitoring
- [x] Implement security scoring algorithm (`backend/scanner/scoring.go`)
- [x] Wire scanner into app.go with real SecurityScore + GetFindings endpoint
- [x] Write tests for scoring algorithm (7 tests passing)
- [x] Add proper error handling and logging (`backend/logging/`, slog throughout app/scanner/agent)
- [x] Write tests for skills (scan_network: 9 tests, lateral_movement: 5 tests)

## Phase 3: Frontend Implementation [COMPLETE]
- [x] Set up TailwindCSS v4 + shadcn/ui utilities (Vite plugin, cn(), CVA, lucide-react)
- [x] Cyberpunk dark theme with neon accents (`src/index.css` theme tokens)
- [x] Generate Wails bindings for frontend TypeScript (manual bindings matching Go API)
- [x] Build Dashboard view (security score gauge, finding summary, system metrics, findings list)
- [x] Build Chat/Audit view (message bubbles, tool call cards, inline code formatting, empty state with suggestions)
- [x] Build Settings view (LLM backend toggle, cloud API key, security options, scan interval)

## Phase 4: Integration & Polish
- [ ] IPC: Real-time tool execution indicators in chat
- [ ] Background monitoring with event notifications
- [ ] Cross-platform testing
- [ ] End-to-end build with `wails build`

## Architecture Notes
- Using picoclaw v0.2.4 for tool registry and LLM provider interfaces
- Agent loop is custom (simpler than picoclaw's full AgentLoop which is designed for standalone bot)
- Ollama support via OpenAI-compatible API at localhost:11434/v1
- Cloud support via HTTP provider (Anthropic, OpenAI compatible)
- All skills implement picoclaw's `tools.Tool` interface
- Configuration stored at platform-appropriate paths (~/.config/neuropanopticon/)
