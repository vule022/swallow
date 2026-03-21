package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vule022/swallow/internal/config"
)

func newHooksCmd(c *Container) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hooks",
		Short: "Manage coding agent hook integrations",
	}
	cmd.AddCommand(
		newHooksInstallCmd(c),
		newHooksRunCmd(c),
	)
	return cmd
}

// ─── hooks install ────────────────────────────────────────────────────────────

func newHooksInstallCmd(c *Container) *cobra.Command {
	var agent string

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install hooks for coding agent integrations",
		Long: `Install hooks so coding agents automatically ingest their sessions into swallow.

Supported agents:
  claude-code   Adds a Stop hook to ~/.claude/settings.json
  cursor        Writes a helper script to ~/.swallow/hooks/cursor-export.sh
  codex         Prints pipe instructions
  all           Installs for all agents (default)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			switch agent {
			case "claude-code":
				return installClaudeCodeHook()
			case "cursor":
				return installCursorHook()
			case "codex":
				return installCodexHook()
			case "all", "":
				var errs []string
				if err := installClaudeCodeHook(); err != nil {
					errs = append(errs, fmt.Sprintf("claude-code: %v", err))
				}
				if err := installCursorHook(); err != nil {
					errs = append(errs, fmt.Sprintf("cursor: %v", err))
				}
				if err := installCodexHook(); err != nil {
					errs = append(errs, fmt.Sprintf("codex: %v", err))
				}
				if len(errs) > 0 {
					return fmt.Errorf("some installs failed:\n  %s", strings.Join(errs, "\n  "))
				}
				return nil
			default:
				return fmt.Errorf("unknown agent %q: use claude-code, cursor, codex, or all", agent)
			}
		},
	}

	cmd.Flags().StringVar(&agent, "agent", "all", "agent to configure: claude-code, cursor, codex, or all")
	return cmd
}

func installClaudeCodeHook() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	settingsPath := filepath.Join(home, ".claude", "settings.json")

	// Read existing settings or start from scratch.
	var raw map[string]interface{}
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("read settings: %w", err)
		}
		raw = map[string]interface{}{}
	} else {
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("parse settings.json: %w", err)
		}
	}

	// Check for existing swallow hook.
	if containsSwallowHook(raw) {
		fmt.Println("claude-code: swallow hook already installed ✓")
		return nil
	}

	// Build the new hook entry.
	newHook := map[string]interface{}{
		"matcher": "",
		"hooks": []interface{}{
			map[string]interface{}{
				"type":    "command",
				"command": "swallow hooks run --agent claude-code",
			},
		},
	}

	// Merge into hooks.Stop array.
	hooks, _ := raw["hooks"].(map[string]interface{})
	if hooks == nil {
		hooks = map[string]interface{}{}
	}
	stopHooks, _ := hooks["Stop"].([]interface{})
	stopHooks = append(stopHooks, newHook)
	hooks["Stop"] = stopHooks
	raw["hooks"] = hooks

	// Write back.
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o700); err != nil {
		return err
	}
	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(settingsPath, out, 0o600); err != nil {
		return fmt.Errorf("write settings.json: %w", err)
	}

	fmt.Printf("claude-code: installed Stop hook → %s\n", settingsPath)
	fmt.Println("  Every Claude Code session will now be automatically ingested by swallow.")
	return nil
}

func containsSwallowHook(raw map[string]interface{}) bool {
	hooks, _ := raw["hooks"].(map[string]interface{})
	if hooks == nil {
		return false
	}
	stopHooks, _ := hooks["Stop"].([]interface{})
	for _, h := range stopHooks {
		hm, ok := h.(map[string]interface{})
		if !ok {
			continue
		}
		innerHooks, _ := hm["hooks"].([]interface{})
		for _, ih := range innerHooks {
			ihm, ok := ih.(map[string]interface{})
			if !ok {
				continue
			}
			if cmd, _ := ihm["command"].(string); strings.Contains(cmd, "swallow") {
				return true
			}
		}
	}
	return false
}

func installCursorHook() error {
	cfgDir, err := config.Dir()
	if err != nil {
		return err
	}
	hooksDir := filepath.Join(cfgDir, config.HooksDirName)
	if err := os.MkdirAll(hooksDir, 0o700); err != nil {
		return err
	}

	scriptPath := filepath.Join(hooksDir, "cursor-export.sh")
	script := `#!/usr/bin/env bash
# Cursor session export helper — run after a Cursor session to ingest it into swallow.
# Usage: ./cursor-export.sh [cursor-dir]
set -euo pipefail

CURSOR_DIR="${1:-$HOME/.cursor}"
MARKER="$CURSOR_DIR/.swallow-last-export"

# Find files newer than last export (or all if first run).
if [ -f "$MARKER" ]; then
  RECENT=$(find "$CURSOR_DIR" -type f \( -name "*.md" -o -name "*.json" -o -name "*.log" \) \
    -newer "$MARKER" 2>/dev/null | sort | head -5)
else
  RECENT=$(find "$CURSOR_DIR" -type f \( -name "*.md" -o -name "*.json" \) \
    2>/dev/null | sort -r | head -1)
fi

if [ -z "$RECENT" ]; then
  echo "No new Cursor files found since last export."
  exit 0
fi

for f in $RECENT; do
  echo "Ingesting: $f"
  swallow ingest-output "$f"
done

touch "$MARKER"
`
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		return fmt.Errorf("write cursor script: %w", err)
	}

	fmt.Printf("cursor: created helper script → %s\n", scriptPath)
	fmt.Printf("  Run it after a Cursor session: %s\n", scriptPath)
	fmt.Println("  Note: adjust CURSOR_DIR or the find pattern for your Cursor version.")
	return nil
}

func installCodexHook() error {
	fmt.Println("codex: pipe instructions:")
	fmt.Println("  codex <prompt> | swallow ingest-output --stdin")
	fmt.Println()
	fmt.Println("  Or add this alias to your shell (~/.zshrc or ~/.bashrc):")
	fmt.Println("    alias codex-ingest='codex \"$@\" | swallow ingest-output --stdin'")
	fmt.Println()
	fmt.Println("  Then use: codex-ingest \"fix the auth flow\"")
	return nil
}

// ─── hooks run (hidden — invoked by agent hooks, not directly by users) ───────

type claudeCodePayload struct {
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
	CWD            string `json:"cwd"`
	HookEventName  string `json:"hook_event_name"`
}

func newHooksRunCmd(c *Container) *cobra.Command {
	var agent string

	cmd := &cobra.Command{
		Use:    "run",
		Short:  "Run a hook handler (invoked by agent hooks, not directly by users)",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if agent != "claude-code" {
				return fmt.Errorf("hooks run: unsupported agent %q (only claude-code is supported)", agent)
			}

			if err := requireDB(c); err != nil {
				fmt.Fprintf(os.Stderr, "swallow hooks run: %v\n", err)
				return nil // non-fatal for hooks
			}

			stdin, err := io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Fprintf(os.Stderr, "swallow hooks run: read stdin: %v\n", err)
				return nil
			}

			var payload claudeCodePayload
			if err := json.Unmarshal(stdin, &payload); err != nil {
				fmt.Fprintf(os.Stderr, "swallow hooks run: parse payload: %v\n", err)
				return nil
			}

			if payload.TranscriptPath == "" {
				fmt.Fprintln(os.Stderr, "swallow hooks run: no transcript_path in payload, skipping")
				return nil
			}

			data, err := os.ReadFile(payload.TranscriptPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "swallow hooks run: read transcript %s: %v\n", payload.TranscriptPath, err)
				return nil
			}

			// Resolve project from the agent's working directory.
			ctx := cmd.Context()
			cwd := payload.CWD
			if cwd == "" {
				cwd, _ = os.Getwd()
			}

			p, err := c.Projects.AutoDetect(ctx, cwd)
			if err != nil {
				p, err = c.Projects.GetActive(ctx)
				if err != nil {
					fmt.Fprintf(os.Stderr, "swallow hooks run: no project found for cwd %s\n", cwd)
					return nil
				}
			}

			_, _, err = ingestRawText(ctx, c, p.ID, payload.TranscriptPath, string(data))
			if err != nil {
				fmt.Fprintf(os.Stderr, "swallow hooks run: ingest: %v\n", err)
				return nil
			}

			id := payload.SessionID
			if id == "" {
				id = "unknown"
			}
			fmt.Fprintf(os.Stderr, "swallow: ingested session %s into project %s\n", id, p.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&agent, "agent", "", "agent name (required): claude-code")
	return cmd
}
