package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const ConfigFile = "config.yml"

type Config struct {
	Benchmark BenchmarkConfig `yaml:"benchmark"`

	// CLI-only.
	Agent string `yaml:"-"`
}

type BenchmarkConfig struct {
	ProjectPath string `yaml:"project_path"`
	Model       string `yaml:"model"`
	ResultsDir  string `yaml:"results_dir"`
	TaskFile    string `yaml:"task_file"`
	Runs        int    `yaml:"runs"`
	Reasoning   string `yaml:"reasoning"`
}

func Load() (Config, error) {
	benchmarkDir, err := os.Getwd()
	if err != nil {
		return Config{}, fmt.Errorf(
			"get working directory: %w",
			err,
		)
	}

	benchmarkDir, err = filepath.Abs(benchmarkDir)
	if err != nil {
		return Config{}, fmt.Errorf(
			"resolve working directory: %w",
			err,
		)
	}

	configPath := filepath.Join(
		benchmarkDir,
		ConfigFile,
	)

	data, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, fmt.Errorf(
			"read %s: %w",
			configPath,
			err,
		)
	}

	var cfg Config

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf(
			"parse %s: %w",
			configPath,
			err,
		)
	}

	if err := validate(cfg); err != nil {
		return Config{}, err
	}

	cfg.Benchmark.ProjectPath, err =
		filepath.Abs(
			cfg.Benchmark.ProjectPath,
		)
	if err != nil {
		return Config{}, fmt.Errorf(
			"resolve project_path: %w",
			err,
		)
	}

	if !filepath.IsAbs(
		cfg.Benchmark.ResultsDir,
	) {
		cfg.Benchmark.ResultsDir =
			filepath.Join(
				benchmarkDir,
				cfg.Benchmark.ResultsDir,
			)
	}

	cfg.Benchmark.ResultsDir, err =
		filepath.Abs(
			cfg.Benchmark.ResultsDir,
		)
	if err != nil {
		return Config{}, fmt.Errorf(
			"resolve results_dir: %w",
			err,
		)
	}

	return cfg, nil
}

func validate(cfg Config) error {
	b := cfg.Benchmark

	if b.ProjectPath == "" {
		return fmt.Errorf(
			"benchmark.project_path must not be empty",
		)
	}

	if b.Model == "" {
		return fmt.Errorf(
			"benchmark.model must not be empty",
		)
	}

	if b.ResultsDir == "" {
		return fmt.Errorf(
			"benchmark.results_dir must not be empty",
		)
	}

	if b.TaskFile == "" {
		return fmt.Errorf(
			"benchmark.task_file must not be empty",
		)
	}

	if b.Runs <= 0 {
		return fmt.Errorf(
			"benchmark.runs must be greater than 0",
		)
	}

	if b.Reasoning != "on" && b.Reasoning != "off" {
		return fmt.Errorf(
			"benchmark.reasoning must be \"on\" or \"off\"",
		)
	}

	return nil
}

func (c Config) Print() {
	fmt.Printf(
		"Project path:  %s\n",
		c.Benchmark.ProjectPath,
	)

	fmt.Printf(
		"Model:         %s\n",
		c.Benchmark.Model,
	)

	fmt.Printf(
		"Results dir:   %s\n",
		c.Benchmark.ResultsDir,
	)

	fmt.Printf(
		"Task file:     %s\n",
		c.Benchmark.TaskFile,
	)

	fmt.Printf(
		"Runs:          %d\n",
		c.Benchmark.Runs,
	)

	fmt.Printf(
		"Reasoning:     %s\n",
		c.Benchmark.Reasoning,
	)

	fmt.Printf(
		"Agent:         %s\n",
		c.Agent,
	)
}
