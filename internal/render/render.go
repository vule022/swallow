package render

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/vule022/swallow/internal/model"
)

// Renderer writes formatted output to a writer.
type Renderer struct {
	w     io.Writer
	color bool
}

// New creates a Renderer writing to stdout with auto-detected color support.
func New() *Renderer {
	return &Renderer{
		w:     os.Stdout,
		color: color.NoColor == false,
	}
}

// NewWriter creates a Renderer writing to the given writer.
func NewWriter(w io.Writer) *Renderer {
	return &Renderer{w: w, color: false}
}

var (
	headerStyle  = color.New(color.Bold, color.FgCyan)
	labelStyle   = color.New(color.Bold)
	successStyle = color.New(color.FgGreen)
	errorStyle   = color.New(color.FgRed)
	dimStyle     = color.New(color.Faint)
)

// Header prints a bold section header.
func (r *Renderer) Header(title string) {
	line := strings.Repeat("─", len(title)+4)
	fmt.Fprintln(r.w)
	headerStyle.Fprintf(r.w, "  %s  \n", title)
	fmt.Fprintf(r.w, "  %s\n", line)
}

// Label prints a bold label with value.
func (r *Renderer) Label(label, value string) {
	labelStyle.Fprintf(r.w, "%-18s", label+":")
	fmt.Fprintln(r.w, value)
}

// Bullet prints a bullet list.
func (r *Renderer) Bullet(items []string) {
	for _, item := range items {
		fmt.Fprintf(r.w, "  • %s\n", item)
	}
}

// NumberedList prints a numbered list.
func (r *Renderer) NumberedList(items []string) {
	for i, item := range items {
		fmt.Fprintf(r.w, "  %d. %s\n", i+1, item)
	}
}

// Success prints a green success message.
func (r *Renderer) Success(msg string) {
	successStyle.Fprintln(r.w, "✓ "+msg)
}

// Error prints a red error message.
func (r *Renderer) Error(msg string) {
	errorStyle.Fprintln(r.w, "✗ "+msg)
}

// Dim prints a faint/dim message.
func (r *Renderer) Dim(msg string) {
	dimStyle.Fprintln(r.w, msg)
}

// Println writes a plain line.
func (r *Renderer) Println(s string) {
	fmt.Fprintln(r.w, s)
}

// PlanResult renders a full PlanResult to the terminal.
func (r *Renderer) PlanResult(result *model.PlanResult, compactOnly bool) {
	if !compactOnly {
		r.Header("SESSION TITLE")
		fmt.Fprintln(r.w, "  "+result.Title)

		r.Header("PRIMARY GOAL")
		fmt.Fprintln(r.w, "  "+result.Goal)

		if result.WhyNow != "" {
			r.Header("WHY NOW")
			fmt.Fprintln(r.w, "  "+result.WhyNow)
		}

		if len(result.CurrentContext) > 0 {
			r.Header("CURRENT CONTEXT")
			r.Bullet(result.CurrentContext)
		}

		if len(result.RelevantFiles) > 0 {
			r.Header("RELEVANT FILES")
			for _, f := range result.RelevantFiles {
				if f.Reason != "" {
					fmt.Fprintf(r.w, "  %-40s  %s\n", f.Path, f.Reason)
				} else {
					fmt.Fprintf(r.w, "  %s\n", f.Path)
				}
			}
		}

		if len(result.ExecutionPlan) > 0 {
			r.Header("EXECUTION PLAN")
			r.NumberedList(result.ExecutionPlan)
		}

		if len(result.Constraints) > 0 {
			r.Header("CONSTRAINTS / NON-GOALS")
			r.Bullet(result.Constraints)
		}

		if len(result.Validation) > 0 {
			r.Header("VALIDATION CHECKLIST")
			r.Bullet(result.Validation)
		}

		if result.ExpectedOutput != "" {
			r.Header("EXPECTED DELIVERABLE")
			fmt.Fprintln(r.w, "  "+result.ExpectedOutput)
		}
	}

	r.CopyBlock(result.CopyReadyPrompt)
}

// CopyBlock prints the copy-ready prompt block.
func (r *Renderer) CopyBlock(prompt string) {
	fmt.Fprintln(r.w)
	labelStyle.Fprintln(r.w, "COPY-READY PROMPT:")
	fmt.Fprintln(r.w, "<<<BEGIN PROMPT")
	fmt.Fprintln(r.w, prompt)
	fmt.Fprintln(r.w, "END PROMPT>>>")
	fmt.Fprintln(r.w)
	successStyle.Fprintln(r.w, "Ready to paste into your coding agent.")
}
