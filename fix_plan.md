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
Required: `libgtk-3-dev`, `libwebkit2gtk-4.1-dev` (Ubuntu 24.04+)
Install with: `sudo apt install libgtk-3-dev libwebkit2gtk-4.1-dev`
Build with: `wails build -tags webkit2_41`
Binary output: `build/bin/neuropanopticon` (~19MB)

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

## Phase 4: Integration & Polish [COMPLETE]
- [x] IPC: Real-time tool execution indicators in chat (Wails EventsEmit, live tool cards in Chat)
- [x] Background monitoring with event notifications
- [x] Cross-platform testing
- [x] End-to-end build with `wails build` (19MB binary, webkit2_41 tag for Ubuntu 24.04+)

## Phase 5: Code Quality & Maintenance [IN PROGRESS]
- [x] Fix agent concurrency bug: mutex held during entire LLM call (up to 2min), blocking Reset()
  - Refactored Chat() to use fine-grained locking around message slice access only
  - LLM calls and tool executions now run without holding the lock
  - Added 3 agent tests: TestResetNotBlockedDuringChat, TestChatAddsMessages, TestResetClearsMessages
- [ ] Populate NetworkIn/NetworkOut fields in SystemStatus (defined in models but never set)
- [x] Add config validation (LLM backend, temperature range, max_tokens bounds, scan interval)
  - Added Validate() method checking backend type, URL/key requirements, temperature [0,2], max_tokens [1,128000], scan interval >= 10
  - Wired into Load() and Save() so invalid configs are rejected at both entry points
  - Added 8 config validation tests with table-driven subtests
- [x] Replace custom string utils in scanner.go (toLower, contains) with strings stdlib
  - Removed toLower(), contains(), searchString() — replaced with strings.ToLower/strings.Contains
  - Kept sanitizeID() (no stdlib equivalent)
- [x] Fix log file handle leak in backend/logging/logging.go (file opened but never closed)
  - Setup() now returns a close function alongside the logger
  - App stores and calls logClose in shutdown() to properly release the file descriptor
- [x] Add negative/zero PID validation in sbom_inspector.go
  - Returns clear error for pid <= 0 before attempting process lookup
  - Added 5 tests: zero, negative, missing, wrong type, and metadata

## Architecture Notes
- Using picoclaw v0.2.4 for tool registry and LLM provider interfaces
- Agent loop is custom (simpler than picoclaw's full AgentLoop which is designed for standalone bot)
- Ollama support via OpenAI-compatible API at localhost:11434/v1
- Cloud support via HTTP provider (Anthropic, OpenAI compatible)
- All skills implement picoclaw's `tools.Tool` interface
- Configuration stored at platform-appropriate paths (~/.config/neuropanopticon/)
