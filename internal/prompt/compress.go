package prompt

import "github.com/sashabaranov/go-openai"

// BuildCompressPrompt builds the messages for summarizing a document or coding output.
func BuildCompressPrompt(text string) []openai.ChatCompletionMessage {
	system := `You are a technical analyst. Extract structured information from the provided text.

Respond with ONLY a JSON object with this exact schema:
{
  "summary": "one paragraph summary",
  "goal": "the main goal or objective mentioned (empty string if none)",
  "actions": ["list of actions taken or described"],
  "decisions": ["list of key decisions made"],
  "blockers": ["list of blockers, issues, or unresolved problems"],
  "next_actions": ["list of suggested next steps"],
  "files_mentioned": ["list of file paths mentioned"]
}

Rules:
- Do not hallucinate or invent facts not present in the text
- Use empty arrays [] for missing fields, not null
- Keep the summary concise (2-4 sentences max)
- Only include items explicitly mentioned in the text`

	return []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: system},
		{Role: openai.ChatMessageRoleUser, Content: text},
	}
}
