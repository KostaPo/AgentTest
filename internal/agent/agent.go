package agent

type Agent interface {
	Name() string

	// Docker Compose service name.
	Service() string

	// CLI arguments passed after the Docker entrypoint.
	Args(task string) []string

	// Parse a single JSONL event.
	ParseEvent(event map[string]any) EventInfo
}

type EventInfo struct {
	IsToolCall bool
	ToolName   string
	IsFinal    bool
	Answer     string
}
