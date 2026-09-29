package agent

import "fmt"

func Build(
	name string,
	model string,
	reasoning string,
) (Agent, error) {
	switch name {
	case "pi":
		return NewPi(
			model,
			reasoning,
		), nil

	case "claude":
		return NewClaude(
			model,
		), nil

	default:
		return nil, fmt.Errorf(
			"unknown agent %q; expected pi or claude",
			name,
		)
	}
}
