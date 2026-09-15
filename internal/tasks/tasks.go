package tasks

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Task struct {
	ID     string `yaml:"id"`
	Tier   int    `yaml:"tier"`
	Prompt string `yaml:"prompt"`
}

type file struct {
	Tasks []Task `yaml:"tasks"`
}

func Load(filename string) ([]Task, error) {
	if filename == "" {
		return nil, fmt.Errorf(
			"task filename must not be empty",
		)
	}

	path := filepath.Join(
		"internal",
		"tasks",
		filename,
	)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf(
			"read task file %s: %w",
			path,
			err,
		)
	}

	var f file

	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf(
			"parse task file %s: %w",
			path,
			err,
		)
	}

	if len(f.Tasks) == 0 {
		return nil, fmt.Errorf(
			"task file %s contains no tasks",
			path,
		)
	}

	for i, task := range f.Tasks {
		if task.ID == "" {
			return nil, fmt.Errorf(
				"task %d: id must not be empty",
				i+1,
			)
		}

		if task.Prompt == "" {
			return nil, fmt.Errorf(
				"task %q: prompt must not be empty",
				task.ID,
			)
		}
	}

	return f.Tasks, nil
}
