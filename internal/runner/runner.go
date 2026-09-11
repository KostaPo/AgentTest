package runner

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"agenttest/internal/agent"
)

type Config struct {
	// Directory containing compose.yaml.
	ComposeDir string

	// Absolute path to compose.yaml.
	ComposeFile string

	// Target repository that will be mounted into /workspace.
	ProjectPath string

	// Where benchmark results are written.
	ResultsDir string

	// Maximum allowed runtime for one agent invocation.
	Timeout time.Duration
}

type Result struct {
	RunID string `json:"run_id"`

	Agent  string `json:"agent"`
	TaskID string `json:"task_id"`
	Tier   int    `json:"tier"`
	Prompt string `json:"prompt"`

	Model string `json:"model"`

	StartedAt  string `json:"started_at"`
	FinishedAt string `json:"finished_at"`

	WallTimeMs        int64  `json:"wall_time_ms"`
	TimeToFirstToolMs *int64 `json:"time_to_first_tool_ms,omitempty"`
	ToolCalls         int    `json:"tool_calls"`

	// Final answer produced by the agent.
	Answer string `json:"answer,omitempty"`

	ExitCode  int  `json:"exit_code"`
	ProcessOK bool `json:"process_ok"`

	StdoutFile string `json:"stdout_file"`
	StderrFile string `json:"stderr_file"`
	EventsFile string `json:"events_file"`
}

type RunRequest struct {
	RunID  string
	TaskID string
	Tier   int
	Prompt string

	Agent agent.Agent
	Model string
}

type Runner struct {
	cfg Config
}

func New(cfg Config) *Runner {
	return &Runner{
		cfg: cfg,
	}
}

func (r *Runner) Run(req RunRequest) (Result, error) {
	started := time.Now()

	runDir := filepath.Join(
		r.cfg.ResultsDir,
		req.RunID,
	)

	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return Result{}, fmt.Errorf(
			"create run directory: %w",
			err,
		)
	}

	stdoutPath := filepath.Join(
		runDir,
		"stdout.jsonl",
	)

	stderrPath := filepath.Join(
		runDir,
		"stderr.log",
	)

	eventsPath := filepath.Join(
		runDir,
		"events.jsonl",
	)

	resultPath := filepath.Join(
		runDir,
		"result.json",
	)

	stdoutFile, err := os.Create(stdoutPath)
	if err != nil {
		return Result{}, fmt.Errorf(
			"create stdout file: %w",
			err,
		)
	}
	defer stdoutFile.Close()

	stderrFile, err := os.Create(stderrPath)
	if err != nil {
		return Result{}, fmt.Errorf(
			"create stderr file: %w",
			err,
		)
	}
	defer stderrFile.Close()

	eventsFile, err := os.Create(eventsPath)
	if err != nil {
		return Result{}, fmt.Errorf(
			"create events file: %w",
			err,
		)
	}
	defer eventsFile.Close()

	ctx := context.Background()

	if r.cfg.Timeout > 0 {
		var cancel context.CancelFunc

		ctx, cancel = context.WithTimeout(
			ctx,
			r.cfg.Timeout,
		)

		defer cancel()
	}

	args := []string{
		"compose",
		"-f",
		r.cfg.ComposeFile,

		"run",
		"--rm",

		"-v",
		fmt.Sprintf(
			"%s:/workspace",
			r.cfg.ProjectPath,
		),

		req.Agent.Service(),
	}

	args = append(
		args,
		req.Agent.Args(req.Prompt)...,
	)

	cmd := exec.CommandContext(
		ctx,
		"docker",
		args...,
	)

	// Run docker-compose from the benchmark project,
	// not from the target repository.
	cmd.Dir = r.cfg.ComposeDir

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return Result{}, fmt.Errorf(
			"stdout pipe: %w",
			err,
		)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return Result{}, fmt.Errorf(
			"stderr pipe: %w",
			err,
		)
	}

	if err := cmd.Start(); err != nil {
		return Result{}, fmt.Errorf(
			"start docker compose: %w",
			err,
		)
	}

	encoder := json.NewEncoder(eventsFile)

	var (
		toolCalls   int
		firstToolMs *int64
		answer      string
	)

	stdoutDone := make(chan error, 1)
	stderrDone := make(chan error, 1)

	// ------------------------------------------------------------
	// stdout
	// ------------------------------------------------------------

	go func() {
		scanner := bufio.NewScanner(stdoutPipe)

		scanner.Buffer(
			make([]byte, 64*1024),
			10*1024*1024,
		)

		for scanner.Scan() {
			line := scanner.Text()

			// Always preserve raw stdout.
			_, _ = fmt.Fprintln(
				stdoutFile,
				line,
			)

			var event map[string]any

			// Preserve non-JSON output in stdout,
			// but don't attempt to parse it as an event.
			if err := json.Unmarshal(
				[]byte(line),
				&event,
			); err != nil {
				continue
			}

			info := req.Agent.ParseEvent(event)

			if info.IsToolCall {
				toolCalls++

				if firstToolMs == nil {
					elapsed := time.Since(started).Milliseconds()
					firstToolMs = &elapsed
				}
			}

			// Keep the latest non-empty answer.
			//
			// This is useful because:
			// - Claude can emit a final "result"
			// - Pi can emit message_end / turn_end
			if strings.TrimSpace(info.Answer) != "" {
				answer = info.Answer
			}

			normalized := map[string]any{
				"t_ms":  time.Since(started).Milliseconds(),
				"agent": req.Agent.Name(),
				"event": event,
			}

			if err := encoder.Encode(normalized); err != nil {
				// Raw stdout is already preserved, so don't fail
				// the entire benchmark because normalized logging failed.
				continue
			}
		}

		stdoutDone <- scanner.Err()
	}()

	// ------------------------------------------------------------
	// stderr
	// ------------------------------------------------------------

	go func() {
		scanner := bufio.NewScanner(stderrPipe)

		scanner.Buffer(
			make([]byte, 64*1024),
			10*1024*1024,
		)

		for scanner.Scan() {
			_, _ = fmt.Fprintln(
				stderrFile,
				scanner.Text(),
			)
		}

		stderrDone <- scanner.Err()
	}()

	// ------------------------------------------------------------
	// Wait
	// ------------------------------------------------------------

	waitErr := cmd.Wait()

	stdoutErr := <-stdoutDone
	stderrErr := <-stderrDone

	if stdoutErr != nil {
		return Result{}, fmt.Errorf(
			"read stdout: %w",
			stdoutErr,
		)
	}

	if stderrErr != nil {
		return Result{}, fmt.Errorf(
			"read stderr: %w",
			stderrErr,
		)
	}

	finished := time.Now()

	exitCode := 0

	if waitErr != nil {
		if exitErr, ok := waitErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	processOK := exitCode == 0

	if ctx.Err() != nil {
		processOK = false
	}

	// ------------------------------------------------------------
	// Result
	// ------------------------------------------------------------

	result := Result{
		RunID:  req.RunID,
		Agent:  req.Agent.Name(),
		TaskID: req.TaskID,
		Tier:   req.Tier,
		Prompt: req.Prompt,

		Model: req.Model,

		StartedAt:  started.UTC().Format(time.RFC3339Nano),
		FinishedAt: finished.UTC().Format(time.RFC3339Nano),

		WallTimeMs:        finished.Sub(started).Milliseconds(),
		TimeToFirstToolMs: firstToolMs,
		ToolCalls:         toolCalls,

		Answer: answer,

		ExitCode:  exitCode,
		ProcessOK: processOK,

		StdoutFile: relativePath(
			r.cfg.ResultsDir,
			stdoutPath,
		),

		StderrFile: relativePath(
			r.cfg.ResultsDir,
			stderrPath,
		),

		EventsFile: relativePath(
			r.cfg.ResultsDir,
			eventsPath,
		),
	}

	data, err := json.MarshalIndent(
		result,
		"",
		"  ",
	)
	if err != nil {
		return Result{}, fmt.Errorf(
			"marshal result: %w",
			err,
		)
	}

	if err := os.WriteFile(
		resultPath,
		data,
		0o644,
	); err != nil {
		return Result{}, fmt.Errorf(
			"write result: %w",
			err,
		)
	}

	return result, nil
}

func relativePath(base, path string) string {
	rel, err := filepath.Rel(
		base,
		path,
	)

	if err != nil {
		return strings.TrimPrefix(
			path,
			base,
		)
	}

	return rel
}
