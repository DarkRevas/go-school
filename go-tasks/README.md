# go-tasks

Задания из курса по Go. Каждое задание - отдельный модуль, можно запускать независимо.

Курс: [go-course](../go-course)

---

## Структура


```
go-tasks/
├── 01-intro/
│   └── 1.1-setup/tasks/
│       ├── task01-hello/         — первая программа, go run
│       └── task02-workspace/     — go work, несколько модулей
│
├── 02-basics/
│   ├── 2.1-variables/tasks/
│   │   ├── task01/               — объявление переменных и типов
│   │   └── task02/               — iota и перечисления
│   ├── 2.2-control-flow/tasks/
│   │   ├── task01/               — FizzBuzz через switch
│   │   └── task02/               — подсчёт символов через for range
│   ├── 2.3-collections/tasks/
│   │   ├── task01/               — рост вместимости среза (len/cap)
│   │   └── task02/               — инвертировать словарь
│   ├── 2.4-functions/tasks/
│   │   ├── task01/               — замер времени через defer
│   │   └── task02/               — счётчик с состоянием (замыкание)
│   └── 2.6-context/tasks/
│       ├── task01/               — таймаут запроса
│       └── task02/               — идентификатор запроса через контекст
│
├── 03-types/
│   ├── 3.1-structs/              — структуры и композиция
│   ├── 3.2-methods/              — методы и указатели
│   ├── 3.3-interfaces/           — интерфейсы
│   ├── 3.4-errors/               — обработка ошибок
│   ├── 3.5-generics/             — дженерики
│   └── 3.7-project/              — практическое задание
│
├── 04-concurrency/
│   ├── 4.1-goroutines/tasks/
│   │   ├── task01/               — параллельная обработка с WaitGroup
│   │   └── task02/               — обнаружить и исправить гонку (-race)
│   ├── 4.2-channels/tasks/
│   │   ├── task01/               — Worker Pool
│   │   └── task02/               — Fan-In: слить два канала
│   ├── 4.3-sync/tasks/
│   │   ├── task01/               — безопасный счётчик на Mutex
│   │   └── task02/               — параллельные запросы с errgroup
│   └── 4.4-iterators/tasks/
│       ├── task01/               — итератор для стека
│       └── task02/               — генератор чисел Фибоначчи
│
├── 05-infrastructure/
│   ├── 5.1-http/                 — HTTP-сервер: net/http
│   ├── 5.2-middleware/           — middleware: цепочка обработчиков
│   ├── 5.3-data/                 — работа с данными
│   ├── 5.4-observability/        — slog, метрики, трейсинг
│   ├── 5.5-grpc/                 — Protobuf & gRPC
│   └── 5.6-project/              — практическое задание
│
└── 06-architecture/
    ├── 6.2-clean-arch/tasks/
    │   └── task01/               — Repository + Service для товаров
    └── 6.3-testing/tasks/
        └── task01/               — табличные тесты и фаззинг парсера
```

---

## Как запускать

Каждое задание - отдельный Go-модуль. Переходи в папку и запускай:

```bash
cd 02-basics/2.1-variables/tasks/task01
go run main.go
```

Для заданий с тестами:
```bash
cd 06-architecture/6.3-testing/tasks/task01
go test -v ./...
```

Проверка на гонки:
```bash
cd 04-concurrency/4.1-goroutines/tasks/task02
go run -race main.go
```

Задание с внешней зависимостью (errgroup):
```bash
cd 04-concurrency/4.3-sync/tasks/task02
go get golang.org/x/sync
go run main.go
```

---

## Как работать с заданиями

Внутри каждого `main.go` есть комментарии `// TODO:` - они показывают что именно нужно реализовать. Описание задачи и ожидаемый вывод - в заголовке файла.

Решения лежат в курсе: `../go-course/<раздел>/solutions/`
