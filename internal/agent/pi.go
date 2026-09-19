package agent

type Pi struct {
	Provider string
	Model    string
	Thinking string
}

func NewPi(
	model string,
	reasoning string,
) *Pi {
	thinking := "off"

	if reasoning == "on" {
		thinking = "max"
	}

	return &Pi{
		Provider: "openrouter",
		Model:    model,
		Thinking: thinking,
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

func (p *Pi) Env(reasoning string) []string {
	return nil
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

	case "message_update":
		// message_update.usage is a cumulative snapshot used for
		// streaming/progress updates. Do not add it to final totals.
		return EventInfo{}

	case "message_end":
		message, ok := event["message"].(map[string]any)
		if !ok {
			return EventInfo{}
		}

		info := EventInfo{
			Answer: extractTextFromMessage(message),
		}

		role, _ := message["role"].(string)

		// Each completed assistant message has its own usage.
		// The runner will accumulate these usages across the run.
		if role == "assistant" {
			info.Usage = extractPiUsage(message)
			info.UsageDelta = true
		}

		return info

	case "agent_settled":
		return EventInfo{
			IsFinal: true,
		}

	default:
		return EventInfo{}
	}
}

func extractTextFromMessage(
	message map[string]any,
) string {
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

func extractPiUsage(
	message map[string]any,
) *Usage {
	rawUsage, ok := message["usage"].(map[string]any)
	if !ok {
		return nil
	}

	usage := &Usage{}

	if value, ok := numberAsInt64(
		rawUsage["input"],
	); ok {
		usage.InputTokens = value
	}

	if value, ok := numberAsInt64(
		rawUsage["output"],
	); ok {
		usage.OutputTokens = value
	}

	if value, ok := numberAsInt64(
		rawUsage["reasoning"],
	); ok {
		usage.ReasoningTokens = value
	}

	if value, ok := numberAsInt64(
		rawUsage["cacheRead"],
	); ok {
		usage.CacheReadTokens = value
	}

	if value, ok := numberAsInt64(
		rawUsage["cacheWrite"],
	); ok {
		usage.CacheCreationTokens = value
	}

	return usage
}
