# swallow

A project-based CLI tool that helps you continue software development sessions more effectively.

Swallow does **not** write production code. It acts as a planning and context-orchestration layer around coding agents.

## What it does

- Collects and organises project context
- Ingests files, folders, and documents
- Ingests coding agent outputs
- Generates the next coding session brief using an LLM
- Exports a compact copy-ready prompt for another coding agent

**Swallow plans. Another coding agent executes.**

## Installation

```bash
go install github.com/vule022/swallow/cmd/swallow@latest
```

## Quick Start

```bash
# Initialise swallow
swallow init

# Create a project from your current directory
cd /path/to/your/project
swallow project init --name myproject

# Ingest your codebase
swallow ingest .

# Generate a session brief
export SWALLOW_API_KEY=sk-...
swallow spit "fix the auth flow"
```

## Commands

See `swallow --help` for the full command reference.
