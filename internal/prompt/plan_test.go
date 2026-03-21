package prompt

import (
	"strings"
	"testing"

	"github.com/vule022/swallow/internal/model"
	"github.com/vule022/swallow/internal/storage"
)

func TestBuildPlanPrompt_ContainsGoal(t *testing.T) {
	req := PlanRequest{
		Goal:        "fix the authentication flow",
		DetailLevel: DetailStandard,
		Context: &storage.RecentContext{
			Project: &model.Project{
				Name:    "testproject",
				Summary: "A test app",
			},
		},
	}

	msgs := BuildPlanPrompt(req)
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}

	userMsg := msgs[1].Content
	if !strings.Contains(userMsg, "fix the authentication flow") {
		t.Error("user message does not contain the goal")
	}
	if !strings.Contains(userMsg, "testproject") {
		t.Error("user message does not contain project name")
	}
}

func TestBuildPlanPrompt_DetailLevelAffectsSchema(t *testing.T) {
	base := PlanRequest{Goal: "test", Context: &storage.RecentContext{}}

	compact := BuildPlanPrompt(PlanRequest{Goal: "test", Context: &storage.RecentContext{}, DetailLevel: DetailCompact})
	detailed := BuildPlanPrompt(PlanRequest{Goal: "test", Context: &storage.RecentContext{}, DetailLevel: DetailDetailed})

	compactSchema := compact[0].Content
	detailedSchema := detailed[0].Content

	// Compact should NOT have validation field.
	if strings.Contains(compactSchema, `"validation"`) {
		t.Error("compact schema should not include validation")
	}
	// Detailed should have validation field.
	if !strings.Contains(detailedSchema, `"validation"`) {
		t.Error("detailed schema should include validation")
	}

	_ = base
}

func TestDetailLevelContextLimits(t *testing.T) {
	tests := []struct {
		level       DetailLevel
		wantOutputs int
		wantDocs    int
	}{
		{DetailCompact, 2, 3},
		{DetailStandard, 5, 8},
		{DetailDetailed, 10, 15},
	}
	for _, tt := range tests {
		outputs, _, docs := tt.level.ContextLimits()
		if outputs != tt.wantOutputs {
			t.Errorf("%s: outputs = %d, want %d", tt.level, outputs, tt.wantOutputs)
		}
		if docs != tt.wantDocs {
			t.Errorf("%s: docs = %d, want %d", tt.level, docs, tt.wantDocs)
		}
	}
}

func TestParseDetailLevel(t *testing.T) {
	tests := []struct {
		input string
		want  DetailLevel
		isErr bool
	}{
		{"compact", DetailCompact, false},
		{"standard", DetailStandard, false},
		{"detailed", DetailDetailed, false},
		{"full", DetailDetailed, false},
		{"", DetailStandard, false},
		{"bogus", DetailStandard, true},
	}
	for _, tt := range tests {
		got, err := Parse(tt.input)
		if tt.isErr && err == nil {
			t.Errorf("Parse(%q): expected error", tt.input)
		}
		if !tt.isErr && err != nil {
			t.Errorf("Parse(%q): unexpected error: %v", tt.input, err)
		}
		if got != tt.want {
			t.Errorf("Parse(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
