# swallow

A project-based CLI tool that helps you continue software development sessions more effectively.

Swallow does **not** write production code. It acts as a **planning and context-orchestration layer** around coding agents — collecting your project context, ingesting agent outputs, and generating a structured next-session brief you can paste directly into another coding agent.

**Swallow plans. Another coding agent executes.**

---

## Installation

```bash
go install github.com/vule022/swallow/cmd/swallow@latest
```

Requires Go 1.22+. No CGO required.

---

## Quick Start

```bash
# 1. Initialise swallow
swallow init

# 2. Create a project from your working directory
cd /path/to/your/project
swallow project init --name myproject --summary "My awesome project"

# 3. Index your codebase
swallow ingest .

# 4. Set your API key (OpenAI-compatible)
export SWALLOW_API_KEY=sk-...

# 5. Generate your next session brief
swallow spit "fix the auth flow"
```

---

## Commands

### `swallow init`
Initialise swallow configuration and storage (`~/.swallow/`).

### `swallow project init [--name NAME] [--summary TEXT]`
Create a new project rooted at the current directory.

```bash
swallow project init --name api-server --summary "REST API with auth and billing"
```

### `swallow project list`
List all projects. The active project is marked with `*`.

### `swallow project use <name-or-id>`
Set the active project by name or ID.

### `swallow status`
Show the active project, indexed file count, and recent session history.

### `swallow ingest <path>`
Ingest a file or directory into the active project.

```bash
swallow ingest .                   # ingest current directory
swallow ingest src/                # ingest a subdirectory
swallow ingest README.md           # ingest a single file
```

### `swallow ingest-output <path> [--stdin]`
Ingest a coding agent output (text/markdown) as structured context.

```bash
swallow ingest-output agent_output.md
cat output.txt | swallow ingest-output --stdin
```

Swallow will attempt to extract goals, decisions, blockers, and next actions from the text using the LLM.

### `swallow spit "<goal>"`
Generate a structured next coding session brief.

```bash
swallow spit "fix the auth flow"
swallow spit "cleanup ingest and storage layer"
swallow spit "add better export" --detail-level detailed
swallow spit "quick fix" --compact-only
swallow spit "refactor" --model claude-3-5-sonnet-20241022
```

**Flags:**
- `--detail-level compact|standard|detailed` — controls plan depth (default: standard)
- `--compact-only` — print only the copy-ready prompt block
- `--full` — generate the most detailed plan possible
- `--model <name>` — override the LLM model for this run

**Output includes:**
1. Session title
2. Primary goal
3. Why now
4. Current context
5. Relevant files
6. Execution plan
7. Constraints / non-goals
8. Validation checklist
9. Expected deliverable
10. Copy-ready prompt block

### `swallow export [--compact] [--full]`
Export a handoff document based on recent project context.

```bash
swallow export
swallow export --full
swallow export --compact
```

### `swallow doctor`
Verify configuration, API key, database, and active project.

```bash
swallow doctor
```

---

## Configuration

Configuration is stored at `~/.swallow/config.json`.

```json
{
  "provider": "openai",
  "model": "gpt-4o",
  "base_url": "https://api.openai.com/v1",
  "default_project": "",
  "copy_mode": "print",
  "storage_backend": "sqlite",
  "max_tokens": 4096,
  "temperature": 0.3
}
```

### Using a different provider

Swallow uses an OpenAI-compatible API client. Any provider with an OpenAI-compatible API works:

```json
{
  "provider": "anthropic",
  "model": "claude-3-5-sonnet-20241022",
  "base_url": "https://api.anthropic.com/v1"
}
```

Or with Ollama locally:

```json
{
  "provider": "ollama",
  "model": "llama3.2",
  "base_url": "http://localhost:11434/v1"
}
```

### API Key

Always set via environment variable — never stored in config:

```bash
export SWALLOW_API_KEY=sk-...
```

---

## How It Works

1. **Ingest** — Swallow recursively scans your project, ignoring noise (node_modules, .git, build artifacts, binaries). For each file it computes a SHA-256 hash for deduplication.

2. **Context building** — When you run `spit`, Swallow fetches your recent coding agent outputs, session history, and a relevance-scored set of document summaries. It builds a structured context prompt.

3. **Planning** — The context is sent to your configured LLM with a strict JSON schema. The model returns a structured plan with title, goal, execution steps, relevant files, and a copy-ready prompt.

4. **Output** — The plan is printed with clear section headers. The copy-ready prompt is printed inside `<<<BEGIN PROMPT ... END PROMPT>>>` delimiters for easy selection.

---

## Project Auto-Detection

If you run swallow from inside a directory that matches a known project root, swallow automatically selects that project — no need to run `swallow project use` every time.

---

## Data Storage

All data is stored locally in `~/.swallow/swallow.db` (SQLite). Nothing is sent anywhere except to your configured LLM provider for planning operations.

---

## Detail Levels

| Level | Context depth | Sections included |
|---|---|---|
| `compact` | 2 outputs, 3 docs | title, goal, execution plan, copy prompt |
| `standard` | 5 outputs, 8 docs | + why now, context, files, constraints, validation |
| `detailed` | 10 outputs, 15 docs | all sections, maximum context |

---

## Multi-Agent Auto-Ingestion

Swallow works with any coding agent. Use `swallow watch` and `swallow hooks install` to automate ingestion so you never need to copy-paste agent outputs manually.

### `swallow watch [--dir <path>] [--project <name>]`

Watches `~/.swallow/inbox/` for new files and ingests them automatically.

```bash
# In a separate terminal (or background):
swallow watch

# Now drop any agent output into the inbox:
cp my_agent_output.md ~/.swallow/inbox/
# → swallow ingests it immediately
```

Files are moved to `~/.swallow/inbox/processed/` after ingestion.

### `swallow hooks install [--agent claude-code|cursor|codex|all]`

Set up automatic ingestion for each agent:

```bash
swallow hooks install              # install for all agents
swallow hooks install --agent claude-code
swallow hooks install --agent cursor
swallow hooks install --agent codex
```

#### Claude Code

Adds a `Stop` hook to `~/.claude/settings.json`. Every time a Claude Code session ends, the transcript is automatically ingested into swallow — no manual action required.

```bash
swallow hooks install --agent claude-code
# Then start swallow watch in the background, and every Claude Code
# session is automatically captured.
```

#### Cursor

Creates `~/.swallow/hooks/cursor-export.sh`. Run it after a Cursor session:

```bash
~/.swallow/hooks/cursor-export.sh
```

#### Codex (OpenAI CLI)

Pipe directly:

```bash
codex "fix the auth flow" | swallow ingest-output --stdin
```

Or add the alias to your shell:

```bash
alias codex-ingest='codex "$@" | swallow ingest-output --stdin'
codex-ingest "fix the auth flow"
```

### Recommended workflow

```bash
# Once, after installing swallow:
swallow hooks install --agent claude-code

# In a background terminal:
swallow watch

# Now every Claude Code session → automatically available for:
swallow spit "what should I work on next"
```
