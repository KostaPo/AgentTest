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
			"MAX_THINKING_TOKENS=" + maxThinkingTokens,
		}
	}

	return []string{
		"MAX_THINKING_TOKENS=0",
	}
}

func (c *Claude) ParseEvent(
	event map[string]any,
) EventInfo {
	eventType, _ := event["type"].(string)

	switch eventType {
	case "assistant":
		return EventInfo{
			Answer: extractAssistantText(event),
		}

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

func extractAssistantText(
	event map[string]any,
) string {
	message, ok := event["message"].(map[string]any)
	if !ok {
		return ""
	}

	content, ok := message["content"].([]any)
	if !ok {
		return ""
	}

	var result string

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

func extractClaudeUsage(
	event map[string]any,
) *Usage {
	usageRaw, ok := event["usage"].(map[string]any)
	if !ok {
		return nil
	}

	usage := &Usage{}

	if value, ok := numberAsInt64(
		usageRaw["input_tokens"],
	); ok {
		usage.InputTokens = value
	}

	if value, ok := numberAsInt64(
		usageRaw["output_tokens"],
	); ok {
		usage.OutputTokens = value
	}

	if value, ok := numberAsInt64(
		usageRaw["cache_read_input_tokens"],
	); ok {
		usage.CacheReadTokens = value
	}

	if value, ok := numberAsInt64(
		usageRaw["cache_creation_input_tokens"],
	); ok {
		usage.CacheCreationTokens = value
	}

	if details, ok :=
		usageRaw["output_tokens_details"].(map[string]any); ok {
		if value, ok := numberAsInt64(
			details["thinking_tokens"],
		); ok {
			usage.ReasoningTokens = value
		}
	}

	return usage
}
