package export

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/vule022/swallow/internal/model"
	"github.com/vule022/swallow/internal/storage"
)

// Options controls export behaviour.
type Options struct {
	Compact bool
	Full    bool
}

// Exporter generates handoff documents.
type Exporter struct {
	repos *storage.Repos
}

// New creates a new Exporter.
func New(repos *storage.Repos) *Exporter {
	return &Exporter{repos: repos}
}

// Export builds a structured handoff string for the project.
func (e *Exporter) Export(ctx context.Context, projectID string, opts Options) (string, *model.PlanResult, error) {
	maxOutputs, maxSessions, maxDocs := 5, 5, 10
	if opts.Full {
		maxOutputs, maxSessions, maxDocs = 10, 10, 20
	} else if opts.Compact {
		maxOutputs, maxSessions, maxDocs = 2, 2, 5
	}

	recentCtx, err := e.repos.Context.FetchRecentContext(ctx, projectID, storage.ContextOptions{
		MaxOutputs:      maxOutputs,
		MaxSessions:     maxSessions,
		MaxDocSummaries: maxDocs,
	})
	if err != nil {
		return "", nil, fmt.Errorf("export: fetch context: %w", err)
	}

	result := buildExportResult(recentCtx, opts)
	doc := renderExportDoc(result, recentCtx, opts)
	return doc, result, nil
}

func buildExportResult(ctx *storage.RecentContext, opts Options) *model.PlanResult {
	result := &model.PlanResult{
		CurrentContext: []string{},
		RelevantFiles:  []model.FileReference{},
		ExecutionPlan:  []string{},
		Constraints:    []string{},
		Validation:     []string{},
	}

	if ctx.Project != nil {
		result.Title = fmt.Sprintf("Handoff: %s", ctx.Project.Name)
		result.Goal = "Continue development of " + ctx.Project.Name
		if ctx.Project.Summary != "" {
			result.WhyNow = ctx.Project.Summary
		}
		result.CurrentContext = append(result.CurrentContext, ctx.Project.ActiveGoals...)
	}

	// Add recent session summaries as context.
	for _, s := range ctx.RecentSessions {
		result.CurrentContext = append(result.CurrentContext,
			fmt.Sprintf("[%s] %s", s.Type, s.Summary))
		if s.NextAction != "" {
			result.ExecutionPlan = append(result.ExecutionPlan, s.NextAction)
		}
	}

	// Add recent output next-actions.
	for _, o := range ctx.RecentOutputs {
		result.ExecutionPlan = append(result.ExecutionPlan, o.NextActions...)
		for _, f := range o.Files {
			result.RelevantFiles = append(result.RelevantFiles, model.FileReference{
				Path:   f,
				Reason: "mentioned in recent coding output",
			})
		}
	}

	// Add doc summaries as relevant files.
	for _, d := range ctx.DocumentSummaries {
		result.RelevantFiles = append(result.RelevantFiles, model.FileReference{
			Path:   d.RelativePath,
			Reason: d.Summary,
		})
	}

	result.CopyReadyPrompt = buildCopyPrompt(ctx, opts)
	result.ExpectedOutput = "Continued development on " + result.Goal
	return result
}

func buildCopyPrompt(ctx *storage.RecentContext, opts Options) string {
	var sb strings.Builder

	if ctx.Project != nil {
		p := ctx.Project
		sb.WriteString(fmt.Sprintf("Project: %s\n", p.Name))
		if p.Summary != "" {
			sb.WriteString(fmt.Sprintf("Description: %s\n", p.Summary))
		}
		if len(p.ActiveGoals) > 0 {
			sb.WriteString("Goals:\n")
			for _, g := range p.ActiveGoals {
				sb.WriteString(fmt.Sprintf("  - %s\n", g))
			}
		}
	}

	if len(ctx.RecentSessions) > 0 {
		sb.WriteString("\nRecent progress:\n")
		for _, s := range ctx.RecentSessions {
			sb.WriteString(fmt.Sprintf("  - %s\n", s.Summary))
		}
	}

	if len(ctx.RecentOutputs) > 0 {
		nextActions := []string{}
		for _, o := range ctx.RecentOutputs {
			nextActions = append(nextActions, o.NextActions...)
		}
		if len(nextActions) > 0 {
			sb.WriteString("\nSuggested next actions:\n")
			for _, a := range nextActions {
				sb.WriteString(fmt.Sprintf("  - %s\n", a))
			}
		}
	}

	if !opts.Compact && len(ctx.DocumentSummaries) > 0 {
		sb.WriteString("\nKey files:\n")
		for _, d := range ctx.DocumentSummaries {
			sb.WriteString(fmt.Sprintf("  - %s: %s\n", d.RelativePath, d.Summary))
		}
	}

	return sb.String()
}

func renderExportDoc(result *model.PlanResult, ctx *storage.RecentContext, opts Options) string {
	var sb strings.Builder
	now := time.Now().Format("2006-01-02 15:04")

	sb.WriteString(fmt.Sprintf("# %s\n\nGenerated: %s\n\n", result.Title, now))

	if ctx.Project != nil {
		sb.WriteString(fmt.Sprintf("## Project: %s\n\n", ctx.Project.Name))
		if ctx.Project.Summary != "" {
			sb.WriteString(ctx.Project.Summary + "\n\n")
		}
	}

	if len(result.CurrentContext) > 0 {
		sb.WriteString("## Current Context\n\n")
		for _, c := range result.CurrentContext {
			sb.WriteString(fmt.Sprintf("- %s\n", c))
		}
		sb.WriteString("\n")
	}

	if len(result.ExecutionPlan) > 0 {
		sb.WriteString("## Next Steps\n\n")
		for i, step := range result.ExecutionPlan {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, step))
		}
		sb.WriteString("\n")
	}

	if len(result.RelevantFiles) > 0 && !opts.Compact {
		sb.WriteString("## Relevant Files\n\n")
		for _, f := range result.RelevantFiles {
			if f.Reason != "" {
				sb.WriteString(fmt.Sprintf("- `%s`: %s\n", f.Path, f.Reason))
			} else {
				sb.WriteString(fmt.Sprintf("- `%s`\n", f.Path))
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Copy-Ready Prompt\n\n```\n")
	sb.WriteString(result.CopyReadyPrompt)
	sb.WriteString("\n```\n")

	return sb.String()
}
