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
	case "tool_use":
		return EventInfo{
			IsToolCall: true,
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
