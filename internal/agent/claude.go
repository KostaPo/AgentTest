package agent

const maxThinkingTokens = "32000"

type Claude struct {
	Model string
}

func NewClaude(model string) *Claude {
	return &Claude{
		Model: model,
	}
}

func (c *Claude) Name() string {
	return "claude"
}

func (c *Claude) Service() string {
	return "claude"
}

func (c *Claude) Args(task string) []string {
	return []string{
		"-p", task,
		"--output-format", "stream-json",
		"--verbose",
	}
}

func (c *Claude) Env(reasoning string) []string {
	if reasoning == "on" {
		return []string{
			"ANTHROPIC_BASE_URL=http://127.0.0.1:8080/api",
			"CLAUDE_PROXY_REASONING=on",
			"MAX_THINKING_TOKENS=" + maxThinkingTokens,
		}
	}

	return []string{
		"ANTHROPIC_BASE_URL=http://127.0.0.1:8080/api",
		"CLAUDE_PROXY_REASONING=off",
		"MAX_THINKING_TOKENS=0",
	}
}

func (c *Claude) ParseEvent(event map[string]any) EventInfo {
	eventType, _ := event["type"].(string)

	switch eventType {
	case "assistant":
		info := EventInfo{
			Answer: extractAssistantText(event),
		}

		// Claude Code обычно передаёт tool_use
		// внутри assistant.message.content,
		// а не отдельным event type.
		if hasToolUse(event) {
			info.IsToolCall = true
			info.ToolName = extractToolName(event)
		}

		return info

	case "result":
		answer, _ := event["result"].(string)

		return EventInfo{
			IsFinal: true,
			Answer:  answer,
			Usage:   extractClaudeUsage(event),
		}

	case "tool_use":
		return EventInfo{
			IsToolCall: true,
		}

	default:
		return EventInfo{}
	}
}

func extractAssistantText(event map[string]any) string {
	message, ok := event["message"].(map[string]any)
	if !ok {
		return ""
	}

	content, ok := message["content"].([]any)
	if !ok {
		return ""
	}

	result := ""

	for _, item := range content {
		block, ok := item.(map[string]any)
		if !ok {
			continue
		}

		blockType, _ := block["type"].(string)

		if blockType != "text" {
			continue
		}

		text, _ := block["text"].(string)
		if text == "" {
			continue
		}

		result += text
	}

	return result
}

func hasToolUse(event map[string]any) bool {
	message, ok := event["message"].(map[string]any)
	if !ok {
		return false
	}

	content, ok := message["content"].([]any)
	if !ok {
		return false
	}

	for _, item := range content {
		block, ok := item.(map[string]any)
		if !ok {
			continue
		}

		blockType, _ := block["type"].(string)

		if blockType == "tool_use" {
			return true
		}
	}

	return false
}

func extractToolName(event map[string]any) string {
	message, ok := event["message"].(map[string]any)
	if !ok {
		return ""
	}

	content, ok := message["content"].([]any)
	if !ok {
		return ""
	}

	for _, item := range content {
		block, ok := item.(map[string]any)
		if !ok {
			continue
		}

		blockType, _ := block["type"].(string)

		if blockType != "tool_use" {
			continue
		}

		name, _ := block["name"].(string)
		return name
	}

	return ""
}

func extractClaudeUsage(event map[string]any) *Usage {
	rawUsage, ok := event["usage"].(map[string]any)
	if !ok {
		return nil
	}

	usage := &Usage{}

	if value, ok := numberAsInt64(rawUsage["input_tokens"]); ok {
		usage.InputTokens = value
	}

	if value, ok := numberAsInt64(rawUsage["output_tokens"]); ok {
		usage.OutputTokens = value
	}

	if value, ok := numberAsInt64(rawUsage["reasoning_tokens"]); ok {
		usage.ReasoningTokens = value
	}

	if value, ok := numberAsInt64(rawUsage["cache_read_input_tokens"]); ok {
		usage.CacheReadTokens = value
	}

	if value, ok := numberAsInt64(rawUsage["cache_creation_input_tokens"]); ok {
		usage.CacheCreationTokens = value
	}

	return usage
}
