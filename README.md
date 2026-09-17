# AgentTest

Benchmark для сравнения coding agents:

- **Pi**
- **Claude Code**

Оба агента запускаются в Docker и используют одну и ту же модель через OpenRouter:

```
deepseek/deepseek-v4-flash-0731
```

---

## Как это работает

Для каждого запуска выбирается один агент:

```bash
go run ./cmd/bench --agent pi
```

или:

```bash
go run ./cmd/bench --agent claude
```

Дальше задачи выполняются **последовательно**.

Каждый run запускается в **новом Docker-контейнере** с чистым контекстом агента.

### Диаграмма выполнения

```
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
```

Контейнер запускается через:

```bash
docker compose run --rm ...
```

Поэтому предыдущий контейнер удаляется после завершения run.

**Важно:** Сам target project не копируется в контейнер, а монтируется:

```
PROJECT_PATH:/workspace
```

---

## Структура проекта

```
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
```

---

## Configuration

Настройки находятся в `config.yml`:

```yaml
benchmark:
  project_path: /home/kostapo/IdeaProjects/AccessMarket
  model: deepseek/deepseek-v4-flash-0731
  results_dir: results
  task_file: tasks.yaml
  runs: 1
```

### Обязательные поля:

| Поле | Описание |
|------|---------|
| `project_path` | Путь к target repository |
| `model` | Модель OpenRouter |
| `results_dir` | Директория с результатами |
| `task_file` | Имя файла задач внутри `internal/tasks` |
| `runs` | Количество повторов каждой задачи (должно быть > 0) |

### API ключ

Секрет OpenRouter находится в `.env`:

```
OPENROUTER_API_KEY=...
```

---

## Tasks

Задачи находятся в:

```
internal/tasks/tasks.yaml
```

### Пример:

```yaml
tasks:
  - id: smoke-1
    tier: 0
    prompt: "Кто ты?"

  - id: smoke-2
    tier: 0
    prompt: "Что ты умеешь?"
```

Имя файла задаётся через параметр `task_file` в конфигурации.

---

## Запуск

### Pi:

```bash
go run ./cmd/bench --agent pi
```

### Claude Code:

```bash
go run ./cmd/bench --agent claude
```

> **Важно:** За один запуск сравнивается только один агент.
>
> Чтобы сравнить агентов, benchmark запускается отдельно:
>
> ```bash
> go run ./cmd/bench --agent pi
> go run ./cmd/bench --agent claude
> ```

---

## Execution model

Задачи выполняются **строго последовательно**:

```
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
```

### Характеристики:

- Следующий run начинается только после **полного завершения** предыдущего
- **Timeout для задач не используется**
- Каждый run:
  - Создаёт новый Docker-контейнер
  - Монтирует target repository в `/workspace`
  - Запускает агента с чистым LLM-контекстом
  - Собирает stdout/stderr и JSON events
  - Ждёт завершения процесса
  - Сохраняет результат
  - Удаляет контейнер (`--rm`)

---

## Agent configuration

### Pi

Pi запускается в JSON mode:

```bash
pi \
  --mode json \
  --provider openrouter \
  --model deepseek/deepseek-v4-flash-0731 \
  --thinking off \
  -p "<TASK>"
```

**Thinking отключён.**

### Claude Code

Claude Code запускается в structured streaming mode:

```bash
claude \
  -p "<TASK>" \
  --output-format stream-json \
  --verbose
```

Для Claude Code thinking отключён через:

```
MAX_THINKING_TOKENS: "0"
```

### Справка:

Benchmark **не использует** отдельный reasoning/thinking режим у одного агента и **не использует** его у другого. Оба агента работают в одинаковых условиях.

---

## Results

Результаты сохраняются в директорию `results/`:

```
results/
└── smoke-1-pi-001/
    ├── stdout.jsonl
    ├── stderr.log
    ├── events.jsonl
    └── result.json
```

### result.json содержит:

- `agent` — используемый агент
- `task` — ID задачи
- `tier` — уровень сложности
- `prompt` — текст задачи
- `model` — модель OpenRouter
- `start/end time` — время начала/завершения
- `wall time` — полное время выполнения
- `time to first tool` — время до первого tool call
- `time to answer` — время до финального ответа
- `tool calls` — количество вызовов инструментов
- `input tokens` — использованные входные токены
- `output tokens` — использованные выходные токены
- `reasoning tokens` — использованные токены на рассуждение
- `answer` — финальный ответ агента
- `exit code` — код выхода процесса
- `process status` — статус процесса
- Пути к raw logs/events

> **Примечание:** Стоимость (cost) в benchmark больше не используется.

---

## Console output

В консоли выводится только **summary**:

```
agent=pi wall=1234ms tools=2 exit=0 first_tool=450ms answer=1200ms input=1870tok output=113tok reasoning=0tok process_ok=true
```

**Полный ответ агента в консоль не выводится.** Он сохраняется в `result.json`.

---

## Current benchmark scope

### Сравнивается:

- ⏱️ Время выполнения
- ⚡ Скорость первого tool call
- 🎯 Время до финального ответа
- 🔧 Количество tool calls
- 📊 Input/output/reasoning tokens
- 💻 Поведение агентов при работе с repository

### Главный принцип текущего MVP:

> **одна задача → один чистый agent run → один контейнер → raw trajectory + основные метрики**