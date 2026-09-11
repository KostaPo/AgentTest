package agent

type Pi struct {
	Provider string
	Model    string
	Thinking string
}

func NewPi(model string) *Pi {
	return &Pi{
		Provider: "openrouter",
		Model:    model,
		Thinking: "off",
	}
}

func (p *Pi) Name() string {
	return "pi"
}

func (p *Pi) Service() string {
	return "pi"
}

func (p *Pi) Args(task string) []string {
	return []string{
		"--mode", "json",
		"--provider", p.Provider,
		"--model", p.Model,
		"--thinking", p.Thinking,
		"-p", task,
	}
}

func (p *Pi) ParseEvent(event map[string]any) EventInfo {
	eventType, _ := event["type"].(string)

	switch eventType {
	case "tool_execution_start":
		toolName, _ := event["toolName"].(string)

		if toolName == "" {
			toolName, _ = event["tool"].(string)
		}

		return EventInfo{
			IsToolCall: true,
			ToolName:   toolName,
		}

	case "message_end":
		answer := extractTextFromMessage(event["message"])

		if answer == "" {
			return EventInfo{}
		}

		return EventInfo{
			IsFinal: true,
			Answer:  answer,
		}

	case "turn_end":
		answer := extractTextFromMessage(event["message"])

		if answer == "" {
			return EventInfo{}
		}

		return EventInfo{
			IsFinal: true,
			Answer:  answer,
		}

	default:
		return EventInfo{}
	}
}

func extractTextFromMessage(raw any) string {
	message, ok := raw.(map[string]any)
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
