// ============================================================
// Задача: Retry Job Queue с экспоненциальным бэкоффом  ⚫ Expert
// ============================================================
//
// Реализуй очередь задач с повторами при ошибке:
//
//   type Queue struct { ... }
//
//   func NewQueue(workers int) *Queue
//   func (q *Queue) Submit(job Job) JobID
//   func (q *Queue) Result(id JobID) (any, error)
//   func (q *Queue) Shutdown(timeout time.Duration) error
//
//   type Job struct {
//       Fn          func(ctx context.Context) (any, error)
//       MaxRetries  int
//       BaseBackoff time.Duration // базовая задержка (экспонента: base, 2*base, 4*base, ...)
//       Timeout     time.Duration // per-attempt таймаут
//   }
//
// Требования:
//   - workers горутин обрабатывают задачи параллельно
//   - При ошибке задача планируется на retry через backoff = BaseBackoff * 2^attempt
//     (+ небольшой джиттер — бонус)
//   - После MaxRetries неудачных попыток — задача уходит в "dead letter" и Result
//     возвращает последнюю ошибку обёрнутую в ErrGivenUp
//   - Result блокируется до финального результата (или до Shutdown)
//   - Shutdown ждёт завершения текущих задач или до timeout
//
// Это усложнённая версия Scheduler из 05-hard/task04 с бэкоффом и retry.
//
// Проверь:
//   go test -race -v ./...

package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var ErrGivenUp = errors.New("retry-queue: исчерпаны попытки")

type JobID int64

type Job struct {
	Fn          func(ctx context.Context) (any, error)
	MaxRetries  int
	BaseBackoff time.Duration
	Timeout     time.Duration
}

type jobState struct {
	job     Job
	id      JobID
	attempt int
	done    chan struct{}
	value   any
	err     error
	readyAt time.Time // для отложенной обработки
}

type Queue struct {
	mu       sync.Mutex
	jobs     map[JobID]*jobState
	pending  []*jobState
	ready    chan *jobState // готовые к запуску
	wg       sync.WaitGroup
	nextID   atomic.Int64
	stopping atomic.Bool
	done     chan struct{}
}

// TODO: реализуй NewQueue
// Подсказка: нужен пул воркеров + отдельный "планировщик" который будит задачи
// когда наступит их readyAt. Можно хранить ожидающие в min-heap или просто в срезе.
func NewQueue(workers int) *Queue {
	q := &Queue{
		jobs:  make(map[JobID]*jobState),
		ready: make(chan *jobState, 64),
		done:  make(chan struct{}),
	}
	q.wg.Add(workers + 1)
	go q.scheduler()
	for i := 0; i < workers; i++ {
		go q.worker()
	}
	return q
}

func (q *Queue) scheduler() {
	defer q.wg.Done()
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-q.done:
			return
		case <-ticker.C:
			now := time.Now()
			q.mu.Lock()
			remain := q.pending[:0]
			for _, js := range q.pending {
				if !js.readyAt.After(now) {
					select {
					case q.ready <- js:
					default:
						remain = append(remain, js)
					}
				} else {
					remain = append(remain, js)
				}
			}
			q.pending = remain
			q.mu.Unlock()
		}
	}
}

func (q *Queue) worker() {
	defer q.wg.Done()
	for {
		select {
		case <-q.done:
			return
		case js, ok := <-q.ready:
			if !ok {
				return
			}
			q.runAttempt(js)
		}
	}
}

func (q *Queue) runAttempt(js *jobState) {
	ctx := context.Background()
	var cancel context.CancelFunc
	if js.job.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, js.job.Timeout)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}
	defer cancel()

	val, err := js.job.Fn(ctx)
	if err == nil {
		js.value = val
		js.err = nil
		close(js.done)
		return
	}

	if js.attempt >= js.job.MaxRetries {
		js.err = fmt.Errorf("%w: %v", ErrGivenUp, err)
		close(js.done)
		return
	}

	backoff := js.job.BaseBackoff * time.Duration(1<<js.attempt)
	js.attempt++
	js.readyAt = time.Now().Add(backoff)
	q.mu.Lock()
	q.pending = append(q.pending, js)
	q.mu.Unlock()
}

// TODO: реализуй Submit
// Подсказка: регистрируй jobState, отправь в ready (readyAt = сейчас)
func (q *Queue) Submit(job Job) JobID {
	if q.stopping.Load() {
		return 0
	}
	id := JobID(q.nextID.Add(1))
	js := &jobState{
		job:     job,
		id:      id,
		done:    make(chan struct{}),
		readyAt: time.Now(),
	}
	q.mu.Lock()
	q.jobs[id] = js
	q.mu.Unlock()
	select {
	case q.ready <- js:
	case <-q.done:
	}
	return id
}

// TODO: реализуй Result — блокирующее ожидание финального результата
func (q *Queue) Result(id JobID) (any, error) {
	q.mu.Lock()
	js, ok := q.jobs[id]
	q.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("job %d not found", id)
	}
	<-js.done
	return js.value, js.err
}

// TODO: реализуй Shutdown
// Подсказка: дай текущим задачам завершиться в рамках timeout;
// новые Submit после Shutdown должны сразу возвращать 0
func (q *Queue) Shutdown(timeout time.Duration) error {
	q.stopping.Store(true)
	finished := make(chan struct{})
	go func() {
		q.wg.Wait()
		close(finished)
	}()
	close(q.done)

	select {
	case <-finished:
		return nil
	case <-time.After(timeout):
		return context.DeadlineExceeded
	}
}

// TODO (внутреннее): worker читает из ready, делает попытку с Timeout через ctx
// При ошибке и attempt < MaxRetries — планируй retry с readyAt = now + BaseBackoff * 2^attempt
// Иначе фиксируй финальный результат (err обёрнутый в ErrGivenUp если попытки исчерпаны)

// TODO (внутреннее): scheduler читает ожидающие задачи и перекладывает в ready
// когда их readyAt наступил

func main() {
	q := NewQueue(3)

	attempts := atomic.Int32{}
	flaky := Job{
		Fn: func(ctx context.Context) (any, error) {
			n := attempts.Add(1)
			if n < 3 {
				return nil, errors.New("временная ошибка")
			}
			return fmt.Sprintf("успех с %d-й попытки", n), nil
		},
		MaxRetries:  5,
		BaseBackoff: 20 * time.Millisecond,
		Timeout:     100 * time.Millisecond,
	}

	id := q.Submit(flaky)
	v, err := q.Result(id)
	fmt.Println("result:", v, "err:", err)

	// Безнадёжная задача
	bad := Job{
		Fn:          func(ctx context.Context) (any, error) { return nil, errors.New("always fails") },
		MaxRetries:  3,
		BaseBackoff: 10 * time.Millisecond,
		Timeout:     50 * time.Millisecond,
	}
	id2 := q.Submit(bad)
	_, err2 := q.Result(id2)
	fmt.Println("dead letter:", errors.Is(err2, ErrGivenUp))

	_ = q.Shutdown(1 * time.Second)
}
