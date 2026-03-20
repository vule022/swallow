package planner

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/vule022/swallow/internal/config"
	"github.com/vule022/swallow/internal/model"
	"github.com/vule022/swallow/internal/prompt"
)

// LLMPlanner calls an OpenAI-compatible API to generate plans and summaries.
type LLMPlanner struct {
	client *openai.Client
	cfg    *config.Config
}

// New creates a new LLMPlanner.
func New(cfg *config.Config, apiKey string) *LLMPlanner {
	clientCfg := openai.DefaultConfig(apiKey)
	if cfg.BaseURL != "" {
		clientCfg.BaseURL = cfg.BaseURL
	}
	return &LLMPlanner{
		client: openai.NewClientWithConfig(clientCfg),
		cfg:    cfg,
	}
}

func (p *LLMPlanner) modelFor(req PlanRequest) string {
	if req.ModelOverride != "" {
		return req.ModelOverride
	}
	return p.cfg.Model
}

// Plan generates a structured session brief.
func (p *LLMPlanner) Plan(ctx context.Context, req PlanRequest) (*model.PlanResult, error) {
	messages := prompt.BuildPlanPrompt(prompt.PlanRequest{
		Goal:        req.Goal,
		Context:     req.Context,
		DetailLevel: req.DetailLevel,
	})

	resp, err := p.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       p.modelFor(req),
		Messages:    messages,
		MaxTokens:   p.cfg.MaxTokens,
		Temperature: float32(p.cfg.Temperature),
	})
	if err != nil {
		return nil, fmt.Errorf("planner: LLM call failed: %w\n\nTip: check SWALLOW_API_KEY and run 'swallow doctor'", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("planner: LLM returned no choices")
	}

	raw := resp.Choices[0].Message.Content
	result, err := parsePlanResult(raw)
	if err != nil {
		// Fallback: wrap raw response.
		return &model.PlanResult{
			Title:           "Session Brief",
			Goal:            req.Goal,
			CopyReadyPrompt: raw,
		}, nil
	}
	return result, nil
}

// Compress summarises a document or coding output.
func (p *LLMPlanner) Compress(ctx context.Context, text string) (*CompressResult, error) {
	messages := prompt.BuildCompressPrompt(text)

	resp, err := p.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       p.cfg.Model,
		Messages:    messages,
		MaxTokens:   1024,
		Temperature: 0.1,
	})
	if err != nil {
		return nil, fmt.Errorf("planner: compress call failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("planner: compress returned no choices")
	}

	raw := resp.Choices[0].Message.Content
	return parseCompressResult(raw)
}

// extractJSON strips markdown fences and finds the outermost JSON object.
func extractJSON(s string) string {
	s = strings.TrimSpace(s)

	// Strip ```json ... ``` or ``` ... ``` fences.
	for _, fence := range []string{"```json", "```JSON", "```"} {
		if strings.Contains(s, fence) {
			start := strings.Index(s, fence)
			if start >= 0 {
				inner := s[start+len(fence):]
				end := strings.Index(inner, "```")
				if end >= 0 {
					s = strings.TrimSpace(inner[:end])
					break
				}
			}
		}
	}

	// Find first { and last } to extract the JSON object.
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start == -1 || end == -1 || end <= start {
		return s
	}
	return s[start : end+1]
}

func parsePlanResult(raw string) (*model.PlanResult, error) {
	jsonStr := extractJSON(raw)

	var result model.PlanResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("planner: parse plan result: %w", err)
	}
	if result.Goal == "" && result.CopyReadyPrompt == "" {
		return nil, fmt.Errorf("planner: plan result missing required fields")
	}
	// Ensure slices are never nil.
	if result.CurrentContext == nil {
		result.CurrentContext = []string{}
	}
	if result.RelevantFiles == nil {
		result.RelevantFiles = []model.FileReference{}
	}
	if result.ExecutionPlan == nil {
		result.ExecutionPlan = []string{}
	}
	if result.Constraints == nil {
		result.Constraints = []string{}
	}
	if result.Validation == nil {
		result.Validation = []string{}
	}
	return &result, nil
}

func parseCompressResult(raw string) (*CompressResult, error) {
	jsonStr := extractJSON(raw)

	var data struct {
		Summary        string   `json:"summary"`
		Goal           string   `json:"goal"`
		Actions        []string `json:"actions"`
		Decisions      []string `json:"decisions"`
		Blockers       []string `json:"blockers"`
		NextActions    []string `json:"next_actions"`
		FilesMentioned []string `json:"files_mentioned"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return &CompressResult{Summary: raw}, nil
	}
	result := &CompressResult{
		Summary:        data.Summary,
		Goal:           data.Goal,
		Actions:        orEmpty(data.Actions),
		Decisions:      orEmpty(data.Decisions),
		Blockers:       orEmpty(data.Blockers),
		NextActions:    orEmpty(data.NextActions),
		FilesMentioned: orEmpty(data.FilesMentioned),
	}
	return result, nil
}

func orEmpty(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
