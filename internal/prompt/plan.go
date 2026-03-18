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
	return fmt.Sprintf(`You are a senior software architect helping a developer plan their next coding session.

Your job is to produce a structured, actionable session brief based on the provided project context and goal.

Rules:
- Do NOT hallucinate code, file contents, or facts not present in the context
- Clearly distinguish assumptions from known facts (prefix assumptions with "Assumption:")
- Prioritise concrete, immediately actionable next steps
- Be specific about file paths when you mention them
- Respond with ONLY valid JSON — no markdown fences, no extra text

Required JSON schema:
%s`, schema)
}

func planJSONSchema(level DetailLevel) string {
	switch level {
	case DetailCompact:
		return `{
  "title": "short session title",
  "goal": "primary goal in one sentence",
  "execution_plan": ["step 1", "step 2", "..."],
  "relevant_files": [{"path": "file/path", "reason": "why it matters"}],
  "copy_ready_prompt": "compact prompt ready to paste into a coding agent"
}`
	case DetailDetailed:
		return `{
  "title": "short session title",
  "goal": "primary goal in one sentence",
  "why_now": "why this is the right thing to work on now",
  "current_context": ["key context point 1", "key context point 2"],
  "relevant_files": [{"path": "file/path", "reason": "why it matters"}],
  "execution_plan": ["detailed step 1", "detailed step 2", "..."],
  "constraints": ["constraint or non-goal 1", "..."],
  "validation": ["validation check 1", "..."],
  "expected_output": "what success looks like",
  "copy_ready_prompt": "full detailed prompt ready to paste into a coding agent"
}`
	default: // standard
		return `{
  "title": "short session title",
  "goal": "primary goal in one sentence",
  "why_now": "why this is the right thing to work on now",
  "current_context": ["key context point"],
  "relevant_files": [{"path": "file/path", "reason": "why it matters"}],
  "execution_plan": ["step 1", "step 2", "..."],
  "constraints": ["constraint or non-goal"],
  "validation": ["validation check"],
  "expected_output": "what success looks like",
  "copy_ready_prompt": "prompt ready to paste into a coding agent"
}`
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
