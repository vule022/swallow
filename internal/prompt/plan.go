package prompt

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/sashabaranov/go-openai"
	"github.com/vule022/swallow/internal/storage"
)

// PlanRequest holds all inputs for building a plan prompt.
type PlanRequest struct {
	Goal        string
	Context     *storage.RecentContext
	DetailLevel DetailLevel
}

// BuildPlanPrompt assembles the full planning prompt.
func BuildPlanPrompt(req PlanRequest) []openai.ChatCompletionMessage {
	system := buildSystemPrompt(req.DetailLevel)
	user := buildUserPrompt(req)
	return []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: system},
		{Role: openai.ChatMessageRoleUser, Content: user},
	}
}

func buildSystemPrompt(level DetailLevel) string {
	schema := planJSONSchema(level)
	return fmt.Sprintf(`You are an expert at writing prompts for AI coding agents (Claude Code, Cursor, Codex, etc.).

A developer will give you a task and project context. Your PRIMARY job is to write the "copy_ready_prompt" field — a detailed, self-contained prompt they can paste directly into a coding agent to get the task done without any back-and-forth.

## What makes a great coding agent prompt

A great prompt leaves nothing ambiguous. The coding agent should be able to read it and immediately start implementing without asking clarifying questions. It must include:

1. **Project context** — what the project does, tech stack, relevant architectural decisions
2. **Current state** — what exists today, what's already been done, relevant file paths and their roles
3. **The task** — stated imperatively and precisely ("Implement X in file Y using pattern Z")
4. **Technical specifications** — exact function signatures, data shapes, API contracts where known
5. **Edge cases to handle** — enumerate them explicitly (empty inputs, concurrent access, auth failures, etc.)
6. **Patterns to follow** — how existing code in the project solves similar problems (so the agent stays consistent)
7. **Constraints** — what NOT to do, what NOT to change, what NOT to break
8. **Acceptance criteria** — how to verify the implementation is correct

## Format of copy_ready_prompt

Write it as clean markdown. Use headings, bullet points, and code blocks freely. It should read like a detailed engineering spec written for a capable coding agent. Length should match complexity — a simple task gets 200 words, a complex one gets 800+.

Start the prompt with the task stated directly. Do not open with "You are a..." or meta-commentary.

## Rules

- Do NOT invent file paths, function names, or facts not present in the context. State assumptions explicitly inline (e.g. "assuming the auth middleware is in internal/middleware/").
- If context is sparse, still write a detailed prompt — just make assumptions explicit.
- The copy_ready_prompt is the most important output. Make it exceptional.
- Respond with ONLY valid JSON — no markdown fences, no extra text.

Required JSON schema:
%s`, schema)
}

func planJSONSchema(level DetailLevel) string {
	copyPromptSpec := copyReadyPromptSpec(level)
	switch level {
	case DetailCompact:
		return fmt.Sprintf(`{
  "title": "short task title, 5 words max",
  "goal": "the task in one sentence",
  "execution_plan": ["concrete step 1", "concrete step 2"],
  "relevant_files": [{"path": "exact/file/path", "reason": "role in this task"}],
  "copy_ready_prompt": %s
}`, copyPromptSpec)
	case DetailDetailed:
		return fmt.Sprintf(`{
  "title": "short task title, 5 words max",
  "goal": "the task in one sentence",
  "why_now": "why tackle this now given project state",
  "current_context": ["relevant fact about current state", "another fact"],
  "relevant_files": [{"path": "exact/file/path", "reason": "role in this task"}],
  "execution_plan": ["precise implementation step 1", "step 2"],
  "constraints": ["do NOT change X", "preserve existing Y behaviour"],
  "validation": ["run test X", "verify behaviour Y", "check edge case Z"],
  "expected_output": "concrete description of the finished implementation",
  "copy_ready_prompt": %s
}`, copyPromptSpec)
	default: // standard
		return fmt.Sprintf(`{
  "title": "short task title, 5 words max",
  "goal": "the task in one sentence",
  "why_now": "why tackle this now given project state",
  "current_context": ["relevant fact about current state"],
  "relevant_files": [{"path": "exact/file/path", "reason": "role in this task"}],
  "execution_plan": ["precise implementation step 1", "step 2"],
  "constraints": ["do NOT change X", "preserve existing Y"],
  "validation": ["how to verify correctness"],
  "expected_output": "concrete description of the finished implementation",
  "copy_ready_prompt": %s
}`, copyPromptSpec)
	}
}

func copyReadyPromptSpec(level DetailLevel) string {
	switch level {
	case DetailCompact:
		return `"A focused, self-contained prompt the developer pastes into a coding agent. Include: task statement, relevant files, key implementation steps, main constraints. 150-300 words. Markdown formatting."`
	case DetailDetailed:
		return `"A comprehensive, self-contained prompt the developer pastes into a coding agent. MUST include all of: (1) project/task context paragraph, (2) current state of relevant code, (3) imperative task statement, (4) relevant files with their roles, (5) step-by-step implementation guide with technical details, (6) edge cases to handle explicitly (empty inputs, concurrent access, error states, auth/permission boundaries, etc.), (7) existing patterns in the codebase to follow for consistency, (8) hard constraints — what not to touch, what not to break, (9) acceptance criteria / how to verify. 500-1000 words. Use markdown with ## headings, bullet points, and inline code. Start directly with the task — no preamble."`
	default:
		return `"A detailed, self-contained prompt the developer pastes into a coding agent. MUST include: (1) brief project/task context, (2) imperative task statement with relevant file paths, (3) implementation steps with technical specifics, (4) key edge cases to handle, (5) constraints and patterns to follow, (6) how to verify success. 300-600 words. Markdown formatting. Start directly with the task."`
	}
}

func buildUserPrompt(req PlanRequest) string {
	var sb strings.Builder

	ctx := req.Context
	if ctx != nil && ctx.Project != nil {
		p := ctx.Project
		sb.WriteString(fmt.Sprintf("PROJECT: %s\n", p.Name))
		if p.RootPath != "" {
			sb.WriteString(fmt.Sprintf("ROOT: %s\n", p.RootPath))
		}
		if p.Summary != "" {
			sb.WriteString(fmt.Sprintf("DESCRIPTION: %s\n", p.Summary))
		}
		if len(p.ActiveGoals) > 0 {
			sb.WriteString("ACTIVE GOALS:\n")
			for _, g := range p.ActiveGoals {
				sb.WriteString(fmt.Sprintf("  - %s\n", g))
			}
		}
		sb.WriteString("\n")
	}

	if ctx != nil && len(ctx.RecentSessions) > 0 {
		sb.WriteString("RECENT SESSION HISTORY:\n")
		for _, s := range ctx.RecentSessions {
			sb.WriteString(fmt.Sprintf("  [%s] %s\n", s.Type, s.Summary))
			if s.NextAction != "" {
				sb.WriteString(fmt.Sprintf("    → next: %s\n", s.NextAction))
			}
		}
		sb.WriteString("\n")
	}

	if ctx != nil && len(ctx.RecentOutputs) > 0 {
		sb.WriteString("RECENT CODING AGENT OUTPUTS:\n")
		for _, o := range ctx.RecentOutputs {
			if o.Goal != "" {
				sb.WriteString(fmt.Sprintf("  Goal: %s\n", o.Goal))
			}
			if len(o.NextActions) > 0 {
				sb.WriteString("  Next actions:\n")
				for _, a := range o.NextActions {
					sb.WriteString(fmt.Sprintf("    - %s\n", a))
				}
			}
			if len(o.Blockers) > 0 {
				sb.WriteString("  Blockers:\n")
				for _, b := range o.Blockers {
					sb.WriteString(fmt.Sprintf("    - %s\n", b))
				}
			}
			if len(o.Decisions) > 0 {
				sb.WriteString("  Decisions:\n")
				for _, d := range o.Decisions {
					sb.WriteString(fmt.Sprintf("    - %s\n", d))
				}
			}
		}
		sb.WriteString("\n")
	}

	if ctx != nil && len(ctx.DocumentSummaries) > 0 {
		relevant := scoreAndFilter(ctx.DocumentSummaries, req.Goal)
		if len(relevant) > 0 {
			sb.WriteString("RELEVANT PROJECT FILES:\n")
			for _, d := range relevant {
				sb.WriteString(fmt.Sprintf("  [%s] %s\n", d.RelativePath, d.Summary))
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString(fmt.Sprintf("DEVELOPER REQUEST: %s\n", req.Goal))
	return sb.String()
}

// scoreAndFilter does a lightweight keyword relevance sort on doc summaries.
func scoreAndFilter(docs []storage.DocumentSummary, goal string) []storage.DocumentSummary {
	tokens := tokenize(goal)
	if len(tokens) == 0 {
		return docs
	}

	type scored struct {
		doc   storage.DocumentSummary
		score int
	}

	var candidates []scored
	for _, d := range docs {
		score := 0
		text := strings.ToLower(d.RelativePath + " " + d.Summary)
		for _, t := range tokens {
			if strings.Contains(text, t) {
				score++
			}
		}
		candidates = append(candidates, scored{d, score})
	}

	// Sort descending by score (simple insertion sort — small N).
	for i := 1; i < len(candidates); i++ {
		for j := i; j > 0 && candidates[j].score > candidates[j-1].score; j-- {
			candidates[j], candidates[j-1] = candidates[j-1], candidates[j]
		}
	}

	result := make([]storage.DocumentSummary, 0, len(candidates))
	for _, c := range candidates {
		result = append(result, c.doc)
	}
	return result
}

var stopWords = map[string]bool{
	"the": true, "a": true, "an": true, "is": true, "in": true,
	"on": true, "at": true, "to": true, "and": true, "or": true,
	"for": true, "of": true, "with": true, "i": true, "my": true,
	"need": true, "want": true, "should": true,
}

func tokenize(s string) []string {
	var tokens []string
	var current strings.Builder
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(r)
		} else if current.Len() > 0 {
			t := current.String()
			if len(t) > 2 && !stopWords[t] {
				tokens = append(tokens, t)
			}
			current.Reset()
		}
	}
	if current.Len() > 2 {
		t := current.String()
		if !stopWords[t] {
			tokens = append(tokens, t)
		}
	}
	return tokens
}
