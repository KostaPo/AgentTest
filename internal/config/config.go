package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Compose-файл по конвенции лежит в директории бенчмарка рядом с .env
// и называется "docker-compose.yml". Переопределяется через
// -compose-file / COMPOSE_FILE, если файл когда-нибудь переименуют.
const defaultComposeFileName = "docker-compose.yml"

// Config — полностью разрешённая конфигурация запуска. Все пути здесь
// уже абсолютные, дальше по коду никто не должен вызывать filepath.Abs
// или думать про working directory.
type Config struct {
	BenchmarkDir string // директория с docker-compose.yml, .env, cmd/, internal/
	ComposeFile  string // абсолютный путь к compose-файлу

	TaskFile string
	Agent    string // "pi", "claude" или "both"
	Runs     int
	Model    string
	Timeout  time.Duration

	ProjectPath string // абсолютный путь к целевому репозиторию (монтируется как /workspace)
	ResultsDir  string
}

// Load разрешает конфигурацию в порядке возрастания приоритета:
// 1. дефолты в коде
// 2. .env в директории бенчмарка
// 3. реальные переменные окружения процесса
// 4. флаги командной строки
func Load() (Config, error) {
	benchmarkDir, err := os.Getwd()
	if err != nil {
		return Config{}, fmt.Errorf("get benchmark directory: %w", err)
	}

	benchmarkDir, err = filepath.Abs(benchmarkDir)
	if err != nil {
		return Config{}, fmt.Errorf("resolve benchmark directory: %w", err)
	}

	envPath := filepath.Join(benchmarkDir, ".env")
	if err := godotenv.Load(envPath); err != nil {
		fmt.Printf("warning: could not load %s: %v\n", envPath, err)
	}

	taskFile := flag.String("task-file", getenv("TASK_FILE", "tasks.yaml"), "path to tasks yaml (env: TASK_FILE)")
	agentName := flag.String("agent", getenv("AGENT", "both"), "pi, claude or both (env: AGENT)")
	runs := flag.Int("runs", getenvInt("RUNS", 1), "number of runs per task (env: RUNS)")
	projectPath := flag.String("project", getenv("PROJECT_PATH", ""), "absolute or relative path to target repository (env: PROJECT_PATH)")
	model := flag.String("model", getenv("MODEL", "deepseek/deepseek-v4-flash-0731"), "model name (env: MODEL)")
	timeout := flag.Duration("timeout", getenvDuration("TIMEOUT", 15*time.Minute), "maximum runtime for one agent invocation (env: TIMEOUT)")
	resultsDir := flag.String("results", getenv("RESULTS_DIR", "results"), "directory for benchmark results (env: RESULTS_DIR)")
	composeFileName := flag.String("compose-file", getenv("COMPOSE_FILE", defaultComposeFileName), "compose file name inside the benchmark directory (env: COMPOSE_FILE)")

	flag.Parse()

	targetPath := *projectPath
	if targetPath == "" {
		return Config{}, fmt.Errorf("PROJECT_PATH must be set (env or -project flag)")
	}

	targetPath, err = filepath.Abs(targetPath)
	if err != nil {
		return Config{}, fmt.Errorf("resolve project path: %w", err)
	}

	info, err := os.Stat(targetPath)
	if err != nil {
		return Config{}, fmt.Errorf("project path: %w", err)
	}
	if !info.IsDir() {
		return Config{}, fmt.Errorf("project path is not a directory: %s", targetPath)
	}

	resultsPath := *resultsDir
	if !filepath.IsAbs(resultsPath) {
		resultsPath = filepath.Join(benchmarkDir, resultsPath)
	}
	resultsPath, err = filepath.Abs(resultsPath)
	if err != nil {
		return Config{}, fmt.Errorf("resolve results directory: %w", err)
	}

	composeFile := filepath.Join(benchmarkDir, *composeFileName)
	if _, err := os.Stat(composeFile); err != nil {
		return Config{}, fmt.Errorf(
			"compose file: %w (expected %q in the benchmark directory - override with -compose-file or COMPOSE_FILE)",
			err, *composeFileName,
		)
	}

	return Config{
		BenchmarkDir: benchmarkDir,
		ComposeFile:  composeFile,

		TaskFile: *taskFile,
		Agent:    *agentName,
		Runs:     *runs,
		Model:    *model,
		Timeout:  *timeout,

		ProjectPath: targetPath,
		ResultsDir:  resultsPath,
	}, nil
}

// Print выводит резолвнутую конфигурацию — то, что раньше печаталось
// прямо из main.go.
func (c Config) Print() {
	fmt.Printf("Benchmark directory: %s\n", c.BenchmarkDir)
	fmt.Printf("Compose file:        %s\n", c.ComposeFile)
	fmt.Printf("Target project:      %s\n", c.ProjectPath)
	fmt.Printf("Results directory:   %s\n", c.ResultsDir)
	fmt.Printf("Model:               %s\n", c.Model)
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getenvDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
