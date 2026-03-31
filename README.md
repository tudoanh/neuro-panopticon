# NeuroPanopticon

**AI-powered local security auditor for your machine.**

NeuroPanopticon is a desktop application that continuously scans your system for security misconfigurations, lateral movement risks, and supply chain threats. It uses an AI agent with real-time tool execution to analyze findings, explain blast radius, and suggest one-click remediations — all running locally on your machine.

Built with [Wails](https://wails.io/) (Go + React), powered by local LLMs via [llama.cpp](https://github.com/ggerganov/llama.cpp) or [Ollama](https://ollama.com/).

## Features

- **Real-time security scanning** — Background scanner monitors ports, processes, and network environment on a configurable interval
- **AI-powered analysis** — Chat with an AI agent that uses specialized security tools instead of guessing
- **Tool-calling agent loop** — The agent can chain multiple tools together to investigate threats end-to-end
- **Security score** — Live 0-100 score based on active findings, weighted by severity
- **Safe remediation** — Block ports, kill processes, disable services — all whitelisted and sanitized, requiring explicit approval
- **Privacy-first** — Everything runs locally by default. Cloud LLM is opt-in
- **Cross-platform** — Works on Linux, macOS, and Windows

## Security Tools

| Tool | Description |
|------|-------------|
| `scan_network` | Lists listening ports and established connections, flags risky services |
| `detect_lateral_movement` | Detects suspicious parent-child process relationships and recon patterns |
| `analyze_blast_radius` | Assesses network environment (VPN, exposed ports) for contextual threat scoring |
| `inspect_sbom` | Extracts loaded libraries from running processes for supply chain analysis |
| `remediate` | Executes safe, pre-approved actions (block port, kill process, disable service) |

## Screenshots

The UI uses a dark cyberpunk theme with three main views:

- **Dashboard** — Security score gauge, system metrics (CPU, RAM, Network I/O), real-time findings list
- **Chat** — Interactive conversation with the AI agent, showing tool execution in real-time
- **Settings** — Configure LLM backend, model, security options, and scan interval

## Getting Started

### Prerequisites

- [Go](https://go.dev/) 1.22+
- [Node.js](https://nodejs.org/) 18+
- [Wails CLI](https://wails.io/docs/gettingstarted/installation) v2
- A local LLM server — either:
  - [llama.cpp](https://github.com/ggerganov/llama.cpp) (recommended)
  - [Ollama](https://ollama.com/)

#### Linux only

```bash
sudo apt install libgtk-3-dev libwebkit2gtk-4.1-dev
```

### Running a local LLM

Using llama.cpp (recommended):

```bash
llama-server \
  -hf Jackrong/Qwen3.5-4B-Claude-4.6-Opus-Reasoning-Distilled-GGUF:Q4_K_M \
  -ngl 99 --jinja --temp 0.7 \
  --port 8999 --alias "Qwen-3.5-9b" \
  --ctx-size 32768
```

Or using Ollama:

```bash
ollama serve
ollama pull llama3.2
```

If using Ollama, update the URL in Settings to `http://localhost:11434`.

### Build & Run

```bash
# Development (hot-reload)
wails dev

# Production build
# On Ubuntu 24.04+ (webkit2gtk 4.1):
CGO_ENABLED=1 go build -tags "webkit2_41 production" -o build/bin/neuropanopticon .

# On older distros (webkit2gtk 4.0):
wails build

# Run
./build/bin/neuropanopticon
```

## Configuration

Configuration is stored as JSON and editable via the Settings UI:

| Setting | Default | Description |
|---------|---------|-------------|
| `llm.backend` | `ollama` | `ollama` (local) or `cloud` (API) |
| `llm.ollama_url` | `http://localhost:8999` | LLM server endpoint |
| `llm.ollama_model` | `Qwen-3.5-9b` | Model name or alias |
| `llm.temperature` | `0.3` | Response randomness (0-2) |
| `llm.max_tokens` | `4096` | Max output tokens (1-128000) |
| `security.allow_remediation` | `false` | Enable remediation actions |
| `security.scan_interval_seconds` | `300` | Background scan interval |

Config file locations:
- Linux: `~/.config/neuropanopticon/config.json`
- macOS: `~/Library/Application Support/NeuroPanopticon/config.json`
- Windows: `%APPDATA%\NeuroPanopticon\config.json`

## Architecture

```
backend/
  agent/       # AI agent loop with multi-turn tool execution
  skills/      # Security tools (network scan, lateral movement, SBOM, etc.)
  scanner/     # Background periodic scanner + scoring algorithm
  config/      # Configuration management with validation
  models/      # Shared data types
  logging/     # Structured logging (slog)

frontend/
  src/
    components/  # React views (Dashboard, Chat, Settings)
    types/       # TypeScript interfaces
    lib/         # Utilities
```

**Tech stack:**
- **Backend:** Go, Wails v2, [picoclaw](https://github.com/sipeed/picoclaw) (AI agent framework), gopsutil (system metrics)
- **Frontend:** React 18, TypeScript, TailwindCSS, Vite, Lucide icons

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests.

## License

MIT License. See [LICENSE](LICENSE) for details.
