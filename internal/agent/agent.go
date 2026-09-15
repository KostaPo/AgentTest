package agent

type Agent interface {
	Name() string
	Service() string

	Args(task string) []string
	Env(reasoning string) []string

	ParseEvent(event map[string]any) EventInfo
}

type EventInfo struct {
	IsToolCall bool
	ToolName   string

	IsFinal bool
	Answer  string

	Usage *Usage
}

type Usage struct {
	InputTokens         int64
	OutputTokens        int64
	ReasoningTokens     int64
	CacheReadTokens     int64
	CacheCreationTokens int64
}
