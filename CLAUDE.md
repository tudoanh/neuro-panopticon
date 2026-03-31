# CLAUDE.md — NeuroPanopticon Project Instructions

## Project Overview

NeuroPanopticon is an AI-powered local security auditor. It is a **watcher** — it detects
threats but does NOT fix them directly. When it finds something suspicious, it tells the
user to summon `claude-code` with bypass permissions to remediate.

**Open source project** — keep code clean, well-documented, and contributor-friendly.

## Build & Run

```bash
# Prerequisites (Ubuntu 24.04+)
sudo apt install libgtk-3-dev libwebkit2gtk-4.1-dev

# Start local LLM (llama-cpp on port 8999)
llama-server -hf Jackrong/Qwen3.5-4B-Claude-4.6-Opus-Reasoning-Distilled-GGUF:Q4_K_M \
  -ngl 99 --jinja --temp 0.7 --port 8999 --alias "Qwen-3.5-9b" --ctx-size 32768

# Development
wails dev

# Production build (Ubuntu 24.04+ with webkit2gtk 4.1)
CGO_ENABLED=1 go build -tags "webkit2_41 production" -o build/bin/neuropanopticon .

# Run
./build/bin/neuropanopticon
```

## Test

```bash
go test ./... -v
```

## Architecture

```
backend/
  agent/       — AI agent loop (picoclaw), LLM provider factory
  skills/      — 5 security tools (scan_network, lateral_movement, blast_radius, sbom, remediate)
  scanner/     — Background periodic scanner + security scoring
  config/      — JSON config with validation, platform-specific paths
  models/      — Shared types (Finding, SystemStatus, ToolCall, etc.)
  logging/     — Structured logging (slog)

frontend/     — React 18 + TypeScript + TailwindCSS (cyberpunk theme)
  components/  — Dashboard, Chat, Settings views

main.go       — Wails app entry, asset embedding
app.go        — IPC bridge (SendMessage, GetSystemStatus, GetFindings, etc.)
```

## Key Design Decisions

- **Watcher-only agent**: The NeuroPanopticon agent detects and reports threats.
  It does NOT auto-fix. For remediation, it should instruct the user to run:
  `claude --dangerously-skip-permissions` to fix the detected issue.
- **Local-first**: All scanning runs locally. Cloud LLM is opt-in.
- **llama-cpp backend**: Default LLM is Qwen-3.5-9b via llama-cpp on port 8999.
  The provider uses OpenAI-compatible API at `/v1/chat/completions`.
- **Tool calls need Type+Function**: llama-cpp requires `"type": "function"` and
  the `function` object on tool calls. See the normalization in `agent.go`.

## Telegram Integration

The autonomous loop communicates bidirectionally with Telegram.
Credentials are in `.env` (never committed). See `.env.example` for the format.

```bash
source .env
# TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID are used by the Ralph loop
```

### Telegram Protocol
- **Receive**: Check for new messages/instructions from the Telegram channel before each loop
- **Send**: Post status updates, findings, and recommendations after each loop
- **Format**: Use markdown formatting for Telegram messages

## Autonomous Loop Behavior

When running autonomously (via Ralph):

1. **Check Telegram** for new instructions at the start of each loop
2. **Execute instructions** if any are pending
3. **If no instructions**: Research and improve the project:
   - Debate with Gemini and Codex on security architecture decisions
   - Research emerging threats in the AI age (prompt injection, model poisoning,
     AI-assisted lateral movement, deepfake social engineering, etc.)
   - Implement improvements that increase user security
   - Keep code clean and contributor-friendly (this is open source)
4. **Update this CLAUDE.md** with any new learnings, patterns, or decisions
5. **Report to Telegram** what was done, what was found, what to do next

## Security Research Focus Areas

- AI-age attack vectors: prompt injection, model supply chain, AI-assisted recon
- LLM-powered lateral movement detection (behavioral analysis, not signatures)
- Supply chain integrity (SBOM analysis, dependency risk scoring)
- Zero-trust local posture (principle of least privilege enforcement)
- Privacy-preserving threat intelligence sharing

## Code Standards (Open Source)

- Go: `gofmt`, meaningful error messages, table-driven tests
- Frontend: TypeScript strict mode, functional components, TailwindCSS utilities
- Commits: conventional commits (`feat:`, `fix:`, `refactor:`, `docs:`)
- No secrets in code — use `.env` or platform config files
- Document public APIs and non-obvious design choices
- Keep dependencies minimal and well-justified

## Config

Stored at platform-specific paths (Linux: `~/.config/neuropanopticon/config.json`).
Defaults in `backend/config/config.go`. Validation enforces boundaries on all fields.

## Last Updated

2026-03-31 — Switched to llama-cpp backend, added Telegram integration,
established watcher-only architecture with claude-code remediation pattern.
