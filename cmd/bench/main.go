package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"agenttest/internal/agent"
	"agenttest/internal/config"
	"agenttest/internal/runner"
	"agenttest/internal/tasks"
)

func main() {
	agentName := flag.String(
		"agent",
		"",
		"agent to run: pi or claude",
	)

	flag.Parse()

	if *agentName == "" {
		log.Fatal(
			"--agent is required (pi or claude)",
		)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	cfg.Agent = *agentName

	taskList, err := tasks.Load(
		cfg.Benchmark.TaskFile,
	)
	if err != nil {
		log.Fatalf(
			"load tasks: %v",
			err,
		)
	}

	a, err := agent.Build(
		cfg.Agent,
		cfg.Benchmark.Model,
		cfg.Benchmark.Reasoning,
	)
	if err != nil {
		log.Fatal(err)
	}

	benchmarkDir, err := os.Getwd()
	if err != nil {
		log.Fatalf(
			"get working directory: %v",
			err,
		)
	}

	r := runner.New(
		runner.Config{
			ComposeDir:  benchmarkDir,
			ComposeFile: "docker-compose.yml",
			ProjectPath: cfg.Benchmark.ProjectPath,
			ResultsDir:  cfg.Benchmark.ResultsDir,
		},
	)

	cfg.Print()

	fmt.Printf(
		"\nAgent: %s\n",
		a.Name(),
	)

	fmt.Printf(
		"Tasks: %d\n",
		len(taskList),
	)

	fmt.Printf(
		"Runs per task: %d\n",
		cfg.Benchmark.Runs,
	)

	for _, task := range taskList {
		for run := 1; run <= cfg.Benchmark.Runs; run++ {
			runID := fmt.Sprintf(
				"%s-%s-%03d",
				task.ID,
				a.Name(),
				run,
			)

			fmt.Printf(
				"\n=== %s ===\n",
				runID,
			)

			result, err := r.Run(
				runner.RunRequest{
					RunID:     runID,
					TaskID:    task.ID,
					Tier:      task.Tier,
					Prompt:    task.Prompt,
					Agent:     a,
					Model:     cfg.Benchmark.Model,
					Reasoning: cfg.Benchmark.Reasoning,
				},
			)

			if err != nil {
				log.Printf(
					"run %s failed: %v",
					runID,
					err,
				)
				continue
			}

			fmt.Printf(
				"agent=%s wall=%dms tools=%d exit=%d",
				result.Agent,
				result.WallTimeMs,
				result.ToolCalls,
				result.ExitCode,
			)

			if result.TimeToAnswerMs != nil {
				fmt.Printf(
					" answer=%dms",
					*result.TimeToAnswerMs,
				)
			}

			fmt.Printf(
				" input=%dtok output=%dtok reasoning=%dtok process_ok=%t\n",
				result.InputTokens,
				result.OutputTokens,
				result.ReasoningTokens,
				result.ProcessOK,
			)
		}
	}
}
