AgentTest

Benchmark для сравнения coding agents:

Pi

Claude Code

Оба агента запускаются в Docker и используют одну и ту же модель через OpenRouter:

deepseek/deepseek-v4-flash-0731

Как это работает

Для каждого запуска выбирается один агент:

go run ./cmd/bench --agent pi

или:

go run ./cmd/bench --agent claude

Дальше задачи выполняются последовательно.

Каждый run запускается в новом Docker-контейнере с чистым контекстом агента.

                     ┌─────────────────────┐
                     │     config.yml      │
                     │ project / model /   │
                     │ tasks / runs        │
                     └──────────┬──────────┘
                                │
                                ▼
                     ┌─────────────────────┐
                     │    cmd/bench        │
                     │  --agent pi/claude  │
                     └──────────┬──────────┘
                                │
                                ▼
                       ┌────────────────┐
                       │   Task 1       │
                       └───────┬────────┘
                               │
                         new container
                               │
                               ▼
                    ┌──────────────────────┐
                    │      Agent run       │
                    │   clean LLM context  │
                    └──────────┬───────────┘
                               │
                        container removed
                               │
                               ▼
                       ┌────────────────┐
                       │   Task 2       │
                       └───────┬────────┘
                               │
                         new container
                               │
                               ▼
                    ┌──────────────────────┐
                    │      Agent run       │
                    │   clean LLM context  │
                    └──────────────────────┘

Контейнер запускается через:

docker compose run --rm ...

Поэтому предыдущий контейнер удаляется после завершения run.

Сам target project не копируется в контейнер, а монтируется:

PROJECT_PATH:/workspace

Структура проекта

AgentTest/
├── .claude-agent/
│   └── Dockerfile
├── .pi-agent/
│   └── Dockerfile
├── cmd/
│   └── bench/
│       └── main.go
├── internal/
│   ├── agent/
│   │   ├── agent.go
│   │   ├── claude.go
│   │   ├── factory.go
│   │   └── pi.go
│   ├── config/
│   │   └── config.go
│   ├── runner/
│   │   └── runner.go
│   └── tasks/
│       ├── tasks.go
│       └── tasks.yaml
├── .env
├── .gitignore
├── config.yml
├── docker-compose.yml
├── go.mod
└── README.md

Configuration

Настройки находятся в config.yml:

benchmark:
project_path: /home/kostapo/IdeaProjects/AccessMarket
model: deepseek/deepseek-v4-flash-0731
results_dir: results
task_file: tasks.yaml
runs: 1

Все поля обязательны:

project_path — путь к target repository

model — модель OpenRouter

results_dir — директория с результатами

task_file — имя файла задач внутри internal/tasks

runs — количество повторов каждой задачи, должно быть > 0

Секрет OpenRouter находится в .env:

OPENROUTER_API_KEY=...

Tasks

Задачи находятся в:

internal/tasks/tasks.yaml

Пример:

tasks:
- id: smoke-1
  tier: 0
  prompt: "Кто ты?"

- id: smoke-2
  tier: 0
  prompt: "Что ты умеешь?"

Имя файла задаётся через:

task_file: tasks.yaml

Запуск

Pi:

go run ./cmd/bench --agent pi

Claude Code:

go run ./cmd/bench --agent claude

За один запуск сравнивается только один агент.

Чтобы сравнить агентов, benchmark запускается отдельно:

go run ./cmd/bench --agent pi
go run ./cmd/bench --agent claude

Execution model

Задачи выполняются строго последовательно:

task 1
run 1
run 2
...

task 2
run 1
run 2
...

task 3
...

Следующий run начинается только после полного завершения предыдущего.

Timeout для задач не используется.

Каждый run:

создаёт новый Docker-контейнер;

монтирует target repository в /workspace;

запускает агента с чистым LLM-контекстом;

собирает stdout/stderr и JSON events;

ждёт завершения процесса;

сохраняет результат;

удаляет контейнер (--rm).

Agent configuration

Pi

Pi запускается в JSON mode:

pi \
--mode json \
--provider openrouter \
--model deepseek/deepseek-v4-flash-0731 \
--thinking off \
-p "<TASK>"

Thinking отключён.

Claude Code

Claude Code запускается в structured streaming mode:

claude \
-p "<TASK>" \
--output-format stream-json \
--verbose

Для Claude Code thinking отключён через:

MAX_THINKING_TOKENS: "0"

Таким образом, benchmark не использует отдельный reasoning/thinking режим у одного агента и не использует его у другого.

Results

Результаты сохраняются в:

results/

Для каждого run создаётся отдельная директория:

results/
└── smoke-1-pi-001/
├── stdout.jsonl
├── stderr.log
├── events.jsonl
└── result.json

result.json содержит:

agent

task

tier

prompt

model

start/end time

wall time

time to first tool

time to answer

tool calls

input tokens

output tokens

reasoning tokens

answer

exit code

process status

paths к raw logs/events

Стоимость (cost) в benchmark больше не используется.

Console output

В консоли выводится только summary:

agent=pi wall=1234ms tools=2 exit=0 first_tool=450ms answer=1200ms input=1870tok output=113tok reasoning=0tok process_ok=true

Полный ответ агента в консоль не выводится. Он сохраняется в result.json.

Current benchmark scope

Сейчас benchmark ориентирован на сравнение:

времени выполнения;

скорости первого tool call;

времени до финального ответа;

количества tool calls;

input/output/reasoning tokens;

поведения агентов при работе с repository.

На текущем этапе нет:

Langfuse;

OpenTelemetry;

параллельного запуска;

общей истории между задачами;

общего контейнера между run;

timeout;

автоматической оценки качества ответа.

Главный принцип текущего MVP:

одна задача → один чистый agent run → один контейнер → raw trajectory + основные метрики.