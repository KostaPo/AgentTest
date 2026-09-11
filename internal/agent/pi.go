package agent

import "fmt"

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
		toolName := ""

		if tool, ok := event["toolName"].(string); ok {
			toolName = tool
		}

		if toolName == "" {
			if tool, ok := event["tool"].(string); ok {
				toolName = tool
			}
		}

		return EventInfo{
			IsToolCall: true,
			ToolName:   toolName,
		}

	case "agent_end":
		return EventInfo{
			IsFinal: true,
		}

	default:
		return EventInfo{}
	}
}

func (p *Pi) String() string {
	return fmt.Sprintf(
		"Pi(provider=%s, model=%s, thinking=%s)",
		p.Provider,
		p.Model,
		p.Thinking,
	)
}
