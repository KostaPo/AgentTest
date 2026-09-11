package agent

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

func (c *Claude) ParseEvent(event map[string]any) EventInfo {
	eventType, _ := event["type"].(string)

	switch eventType {
	case "assistant":
		// Assistant text can appear here too.
		answer := extractAssistantText(event)

		if answer == "" {
			return EventInfo{}
		}

		return EventInfo{
			Answer: answer,
		}

	case "result":
		answer, _ := event["result"].(string)

		return EventInfo{
			IsFinal: true,
			Answer:  answer,
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

		if text != "" {
			return text
		}
	}

	return ""
}
