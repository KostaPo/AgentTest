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

	// UsageDelta означает, что поле Usage содержит данные об использовании
	// для конкретного сообщения ассистента и должно быть добавлено к общим показателям.
	//
	// Если значение равно false, то Usage представляет собой итоговый показатель
	// использования для данного выполнения и должно заменить текущие итоговые значения.
	UsageDelta bool
}

type Usage struct {
	InputTokens         int64
	OutputTokens        int64
	ReasoningTokens     int64
	CacheReadTokens     int64
	CacheCreationTokens int64
}
