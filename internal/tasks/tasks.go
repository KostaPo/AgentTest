package tasks

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type File struct {
	Tasks []Task `yaml:"tasks"`
}

type Task struct {
	ID     string `yaml:"id"`
	Tier   int    `yaml:"tier"`
	Prompt string `yaml:"prompt"`
}

// Load читает и парсит tasks.yaml.
func Load(path string) ([]Task, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read task file: %w", err)
	}

	var file File
	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("parse task file: %w", err)
	}

	if len(file.Tasks) == 0 {
		return nil, fmt.Errorf("no tasks found in %s", path)
	}

	return file.Tasks, nil
}
