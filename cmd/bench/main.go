package main

import (
	"fmt"
	"log"

	"agenttest/internal/agent"
	"agenttest/internal/config"
	"agenttest/internal/runner"
	"agenttest/internal/tasks"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	taskList, err := tasks.Load(cfg.TaskFile)
	if err != nil {
		log.Fatalf("load tasks: %v", err)
	}

	agents, err := agent.Build(cfg.Agent, cfg.Model)
	if err != nil {
		log.Fatal(err)
	}

	r := runner.New(runner.Config{
		ComposeDir:  cfg.BenchmarkDir,
		ComposeFile: cfg.ComposeFile,
		ProjectPath: cfg.ProjectPath,
		ResultsDir:  cfg.ResultsDir,
		Timeout:     cfg.Timeout,
	})

	cfg.Print()

	for _, task := range taskList {
		for _, a := range agents {
			for run := 1; run <= cfg.Runs; run++ {
				runID := fmt.Sprintf("%s-%s-%03d", task.ID, a.Name(), run)
				fmt.Printf("\n=== %s ===\n", runID)

				result, err := r.Run(runner.RunRequest{
					RunID:  runID,
					TaskID: task.ID,
					Tier:   task.Tier,
					Prompt: task.Prompt,

					Agent: a,
					Model: cfg.Model,
				})
				if err != nil {
					log.Printf("run %s failed: %v", runID, err)
					continue
				}

				fmt.Printf(
					"agent=%s wall=%dms tools=%d exit=%d",
					result.Agent, result.WallTimeMs, result.ToolCalls, result.ExitCode,
				)
				if result.TimeToFirstToolMs != nil {
					fmt.Printf(" first_tool=%dms", *result.TimeToFirstToolMs)
				}
				fmt.Printf(" process_ok=%t\n", result.ProcessOK)
			}
		}
	}
}
