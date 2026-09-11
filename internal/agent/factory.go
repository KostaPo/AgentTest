package agent

import "fmt"

// Build собирает набор агентов для запуска по имени селектора.
func Build(name, model string) ([]Agent, error) {
	switch name {
	case "pi":
		return []Agent{NewPi(model)}, nil

	case "claude":
		return []Agent{NewClaude(model)}, nil

	case "both":
		return []Agent{NewPi(model), NewClaude(model)}, nil

	default:
		return nil, fmt.Errorf("unknown agent %q; expected pi, claude or both", name)
	}
}
