# 🛡️ PRD & BOOTSTRAP GUIDE: NEUROPANOPTICON

**Subject:** Build "NeuroPanopticon" - The AI-Driven Local Security Auditor

## 1. PROJECT OVERVIEW & CONTEXT

You are tasked with building **NeuroPanopticon** (short: Panopti). It is an open-source, ultra-lightweight, cross-platform (Windows, macOS, Linux) active defense security agent.

**The Problem:** Traditional AVs use static signatures. Attackers (and AI-driven offensive agents) now use lateral movement, supply chain vulnerabilities (e.g., the *Cuckoo's Egg* Emacs scenario), and "Living off the Land" (LotL) techniques.
**The Solution:** NeuroPanopticon acts as an "all-seeing neural eye." It uses a local AI agent to continuously audit system configurations, analyze blast radii, detect lateral movement anomalies, and provide actionable 1-click remediations.

**Monetization Strategy (Open-Core):**

* **Free Tier:** 100% open-source, relies on Local LLMs (via Ollama), manual threat updates.
* **Subscription Tier (Panopti Pro):** Cloud AI fallback (Claude 4.6/GPT-5) for complex malware analysis, auto-sync threat intelligence, centralized dashboard, and 1-click compiled binaries.

---

## 2. TECH STACK MANDATE

Do not deviate from this stack. We need maximum performance, minimal RAM usage, and seamless cross-platform compilation.

* **Backend & System Logic:** Golang (Go 1.21+)
* **Frontend/GUI:** Wails v2 (Native OS Webview, no Electron) + React 18 + TailwindCSS + shadcn/ui.
* **AI Agentic Framework:** `picoclaw` (github.com/sipeed/picoclaw) - a lightweight Go-based AI agent loop.
* **System Interaction Libs:** `shirou/gopsutil` (for cross-platform process/port monitoring), standard OS execution libraries.
* **LLM Backend:** Agnostic (Must support `localhost:11434` for Ollama and standard REST API for Cloud AI).

---

## 3. ARCHITECTURE & DIRECTORY STRUCTURE

Initialize the project using Wails and structure the Go backend as follows:

```text
neuropanopticon/
├── build/                 # Wails build assets (icons, etc.)
├── frontend/              # React + Tailwind SPA
├── backend/
│   ├── agent/             # PicoClaw initialization and Agentic Loop setup
│   ├── skills/            # Go functions (Tools) injected into PicoClaw
│   ├── scanner/           # Background system monitoring (Syslog, Ports)
│   ├── models/            # Data structs for IPC between Go and React
│   └── config/            # User settings, License/Subscription validation
├── main.go                # Wails app lifecycle & Backend entry point
└── PRD.md                 # This file
```

---

## 4. CORE FEATURES & PICO-CLAW "SKILLS"

The genius of this app is that we don't hardcode logic. We provide Go functions ("Skills") to `picoclaw` and let the AI deduce the threats. You must implement the following Go interfaces and wrap them as PicoClaw tools:

### Skill 1: `Skill_ScanNetwork()`

* **Action:** Uses `gopsutil/net` to list all active listening ports and established connections.
* **Use Case for AI:** When the user asks "Is my computer exposing any dangerous services?", the AI calls this tool, sees Port 3389 (RDP) or 445 (SMB) open, and warns the user.

### Skill 2: `Skill_SBOM_Inspector(processID)`

* **Action:** Analyzes a running process or executable to extract its Software Bill of Materials (dependencies).
* **Use Case for AI:** To detect supply-chain attacks. If a chat app is running, the AI checks if it uses a vulnerable `libwebp` version and flags it.

### Skill 3: `Skill_LateralMovementDetector()`

* **Action:** Reads recent process execution logs (e.g., parsing PowerShell/Bash history or process arguments).
* **Use Case for AI:** To detect the *Cuckoo's Egg* scenario. If a PDF reader spawns a PowerShell process attempting to ping the local subnet, the AI identifies the context mismatch and blocks/alerts.

### Skill 4: `Skill_BlastRadiusAnalyzer()`

* **Action:** Checks network interfaces (e.g., Is a VPN active? Are we on a public WiFi?).
* **Use Case for AI:** Contextual threat scoring. A vulnerability is scored higher if the machine is actively connected to a corporate VPN, preventing the local machine from being a pivot point.

### Skill 5: `Skill_Remediate(actionType, target)`

* **Action:** Executes a safe, predefined OS command (e.g., adding a firewall rule to block an IP, killing a process).
* **Use Case for AI:** The AI uses this tool to provide 1-click fixes to the user via the Wails frontend.

---

## 5. SYSTEM PROMPT FOR NEUROPANOPTICON

Inject this system prompt into the `picoclaw` engine initialization:

> "You are NeuroPanopticon, an elite, locally-hosted cybersecurity auditor. Your mission is to protect this machine from lateral movement, supply chain attacks, and misconfigurations. You have access to specialized system tools (Skills). Do not guess system states; use your tools to fetch real-time data. If you detect an anomaly, explain the blast radius (how it could spread) and suggest a remediation tool. Prioritize user privacy: never upload sensitive file contents unless explicitly authorized via the Pro Cloud Fallback."

---

## 6. FRONTEND (UI/UX) REQUIREMENTS

Build the React frontend with a cyberpunk yet clean aesthetic (Dark mode default, neon accents).

* **Dashboard View:**
  * Overall Security Score (0-100).
  * Active monitoring status (CPU, RAM, Network footprint).
* **Chat/Audit View:**
  * A chat interface where the user can talk to the PicoClaw agent.
  * Example: User types "Audit my system for lateral movement risks." -> UI shows a loading state -> Agent uses skills -> Renders a markdown response with action buttons.
* **Subscription View:**
  * Toggle between "Local Engine (Ollama)" and "Panopti Pro Engine (Cloud)".
  * License key input field.

---

## 7. EXECUTION PLAN FOR `claude-code`

Please execute the following steps sequentially. Ask for my confirmation if you encounter cross-platform dependency issues.

1. **Phase 1: Project Initialization**
   * Run `wails init -n neuropanopticon -t react-ts`.
   * Set up TailwindCSS and `shadcn/ui` in the `frontend` folder.
2. **Phase 2: Backend Core & Agent Setup**
   * Install Go dependencies: `gopsutil` and `picoclaw` (`go get github.com/sipeed/picoclaw`).
   * Create the `agent.go` file to initialize the PicoClaw instance.
3. **Phase 3: Skill Development**
   * Implement the 5 Skills defined in Section 4 using standard Go libraries. Ensure cross-platform compatibility (use conditional compilation `_windows.go`, `_darwin.go` if absolutely necessary, but prefer agnostic libs).
   * Register these skills with the PicoClaw agent.
4. **Phase 4: IPC (Inter-Process Communication)**
   * Bind the Go Agent methods to the Wails runtime so the React frontend can send chat messages to PicoClaw and receive tool-call updates in real-time.
5. **Phase 5: Frontend Implementation**
   * Build the Dashboard and Chat UI. Ensure tool executions (like scanning ports) trigger a visual loading indicator in the chat.

**Let's build the ultimate AI shield. Acknowledge this PRD and begin Phase 1.**


For local models on ollama, use qwen 3.5 models that suit to users graphic cards. You can use Q4_K_M or Q4_0 for most models without much quality differ.
