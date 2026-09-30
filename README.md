<p align="center">
  <img src="docs/brand/icon.svg" width="112" alt="OpenSpec Studio" />
</p>

<h1 align="center">OpenSpec Studio</h1>

<p align="center">
  <strong>The operating system for AI software teams.</strong><br />
  Write specs, queue them up, and let AI agents implement them on a schedule.
</p>

<p align="center">
  <a href="#quick-start">Quick start</a> ·
  <a href="#usage">Usage</a> ·
  <a href="#desktop-app-macos">Desktop app</a> ·
  <a href="#configuration">Configuration</a> ·
  <a href="#troubleshooting">Troubleshooting</a>
</p>

---

OpenSpec Studio is a local authoring and project-management environment for [OpenSpec](https://github.com/Fission-AI/OpenSpec) projects. Open your real repositories, write and link specs, track changes on a kanban board, and hand fully-planned changes to an AI agent (Claude Code or any CLI agent) that runs them in order while you're away.

It runs as a **web app** (one self-contained binary) or as a **macOS desktop app**, over the same data. Your `openspec/` folders stay the source of truth: the Studio reads and writes real OpenSpec files and never copies spec content into its own database.

## Quick start

```bash
git clone https://github.com/davi-fati/openspec-studio.git
cd openspec-studio
make install   # Go modules + npm packages
make build     # one binary with the UI embedded
./backend/bin/openspec-studio-server
```

Open **http://127.0.0.1:4173** and click **Open** to add your first OpenSpec project.

### Requirements

| Tool | Version | Needed for |
|---|---|---|
| Go | 1.25+ | everything |
| Node.js | 20.19+ or 22.12+ | building the UI |
| Rust (`rustup`) | stable | the desktop app only |
| An agent CLI, e.g. [Claude Code](https://docs.anthropic.com/en/docs/claude-code) | any | AI features (optional) |

## Usage

The Studio has four tabs, in the order you'll use them: **Overview → Kanban → Specs → Specflow**. The project selector in the header scopes every tab to one project, or to **All projects**.

### 1. Add a project

| Button | Use it for |
|---|---|
| **Open** | A folder that already has an `openspec/` directory. |
| **New** | An empty folder. The Studio scaffolds the same layout `openspec init` creates. |

Projects are remembered across restarts. A project whose folder was moved or deleted shows as *unavailable* instead of disappearing.

### 2. See where everything stands (Overview)

The landing tab. One card per project with its specs, changes by stage, and whether a Specflow run is active. Click any number to jump to the matching Kanban or Specs view.

### 3. Track changes (Kanban)

Every OpenSpec change is a card, placed in a column by its lifecycle stage. Click a card for its proposal, design, tasks and delta specs. The board updates live when files change on disk, whether you edit them in your editor or an agent does.

### 4. Write specs (Specs)

- **Create:** write a spec by hand, or describe it in a sentence and let your AI provider draft the requirements and scenarios. Nothing is written to disk until you save.
- **Edit / delete:** changes go straight to the spec's `spec.md`.
- **Link:** declare that one spec depends on another and see the dependency graph.
- **Diagrams:** ` ```mermaid ` blocks in any spec or change render as diagrams, in light and dark theme.
- **Project context:** keep review notes and tech-debt items next to the project's specs.

Every spec created in the Studio records its id, creation time, author and project.

### 5. Connect an AI provider

Open **Settings** (⚙ in the header), add a provider, then click **Check** to confirm it works. Checks only run when you click them, so they never spend API credits on their own.

**CLI agent (recommended)**, via **Add CLI agent provider**. The Studio runs the executable with your arguments followed by the prompt as the last argument. For Claude Code:

| Field | Value |
|---|---|
| Executable | `claude` |
| Arguments | `-p --permission-mode acceptEdits` |

Any agent that accepts a prompt as its final argument works the same way.

**Hosted API**, via **Add hosted API provider**. An endpoint, a model and an API key. The Studio sends `POST {"model": ..., "prompt": ...}` with `Authorization: Bearer <key>`, so point it at an endpoint (or proxy) that accepts that shape.

API keys are stored in the macOS Keychain, or in an AES-encrypted file in the data directory where no keychain exists. They are never written to the database.

### 6. Ship while you sleep (Specflow)

1. Open **Specflow** → **Schedule flow**, and pick a project.
2. Add changes. Only fully-planned changes (proposal, specs and `tasks.md`) can be added, and a change can't run before a change it depends on.
3. Order them, pick a provider and a start time.
4. At that time, the agent is asked to implement each change from its `tasks.md`, one after another.

Each item shows its status and a live log. A failure stops only that flow. **Cancel** a running flow at any time (it also stops the agent process and its children), or delete a pending one. Changes being implemented are marked as such on the Overview and the Kanban.

> [!IMPORTANT]
> Scheduled runs only fire while the Studio is running; a flow whose start time passed while it was off is marked **missed**. The desktop app stays in the menu bar when you close its window with a flow pending, so an overnight run still happens.

## Desktop app (macOS)

```bash
make desktop-build
open "src-tauri/target/release/bundle/macos/OpenSpec Studio.app"
```

A `.dmg` is built next to it in `src-tauri/target/release/bundle/dmg/`.

The desktop app runs the same backend and UI as the web app, and adds:

- **Native folder picker** for Open and New.
- **Menu-bar mode:** closing the window while a flow is scheduled or running keeps the Studio alive in the menu bar (⌘ icon → **Open Studio** / **Quit**). With nothing pending, closing the window quits.
- **Your shell's `PATH`:** CLI agents installed through your shell profile (`~/.local/bin`, nvm, asdf…) are found even when the app is launched from the Dock.
- **Safe quit:** quitting with a flow running asks first, then cancels it cleanly.

**Web and desktop share everything.** Projects, providers and flows live in one data directory, and only one Studio backend can use it at a time. If a web backend is already running, the desktop app attaches to it instead of starting a second one, and leaves it running when you quit.

> [!NOTE]
> The app is not signed yet. If macOS blocks it, right-click the app → **Open**, or run `xattr -dr com.apple.quarantine "OpenSpec Studio.app"`.

## Configuration

**Server flags**

```bash
openspec-studio-server [--port <port>] [--inherit-login-path]
```

| Flag / variable | Default | Description |
|---|---|---|
| `--port` | `4173` | Port to listen on, bound to `127.0.0.1` only. `0` picks a free port. |
| `OPENSPEC_STUDIO_PORT` | — | Same as `--port`; the flag wins. |
| `--inherit-login-path` | off | Resolve `PATH` from your login shell. The desktop app sets it. |

**Data directory**

| OS | Path |
|---|---|
| macOS | `~/Library/Application Support/openspec-studio/` |
| Linux | `~/.config/openspec-studio/` |

It holds `studio.db` (projects, providers, flows) and `studio.lock`. Your specs are never stored here.

**API.** Every endpoint lives under `/api`. Interactive docs are at **http://127.0.0.1:4173/swagger/index.html**.

## Commands

| Command | What it does |
|---|---|
| `make install` | Install Go and npm dependencies |
| `make dev` | Backend on `:4173` + Vite with hot reload on `:5173` (open the Vite URL) |
| `make build` | Build the self-contained web binary: `backend/bin/openspec-studio-server` |
| `make test` | Backend tests + frontend typecheck and build |
| `make desktop-dev` | Run the desktop app in development mode |
| `make desktop-build` | Build `OpenSpec Studio.app` and the `.dmg` |
| `make desktop-rebuild` | `desktop-clean` + `desktop-build` |
| `make desktop-stop` | Quit the desktop app and stop the backend it started |
| `make desktop-clean` | Remove every desktop build artifact, installed copy and webview cache; keeps your data |
| `make icons` | Regenerate app and web icons from the SVG masters (needs `rsvg-convert`) |
| `make swagger` | Regenerate the OpenAPI docs |
| `make kill-ports` | Free ports `4173` and `5173` |

## Troubleshooting

<details>
<summary><strong>"An OpenSpec Studio backend is already running at …"</strong></summary>

Only one backend can use the data directory. Use the address in the message, or stop the other one: `make desktop-stop` for the desktop app, or <kbd>Ctrl</kbd>+<kbd>C</kbd> in its terminal.
</details>

<details>
<summary><strong>A CLI provider says "executable not found"</strong></summary>

Run **Check** in Settings. If the command works in your terminal but not in the Studio, give the full path (`which claude`) as the executable.
</details>

<details>
<summary><strong>Port 4173 is already in use</strong></summary>

Run `make kill-ports`, or start on another port: `./backend/bin/openspec-studio-server --port 4200`.
</details>

<details>
<summary><strong>The desktop app shows "could not start"</strong></summary>

The screen includes the backend's output. The most common cause is a Studio of a different version already running; stop it and reopen the app.
</details>

## Contributing

This project plans its own work with OpenSpec: capabilities live in [`openspec/specs/`](openspec/specs), and finished changes are archived in [`openspec/changes/archive/`](openspec/changes/archive). Architecture and conventions are in [`openspec/project.md`](openspec/project.md).

Branches follow git flow: open `feature/*` from `develop`, and releases go to `main`.

## License

No license has been chosen yet, so all rights are reserved by default.
