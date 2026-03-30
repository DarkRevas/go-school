// ============================================================
// Задача: Планировщик задач с приоритетами  ⚫ Expert
// ============================================================
//
// Вопрос с финальных этапов собеса уровня Staff+.
//
// Реализуй Scheduler — планировщик с:
//   - Приоритетами (High > Medium > Low)
//   - Отменой через context
//   - Зависимостями: задача B стартует только после завершения задачи A
//   - Дедлайнами: задача отменяется если не стартовала до дедлайна
//   - Метриками: время ожидания, время выполнения
//
//   type Scheduler struct { ... }
//
//   func NewScheduler(workers int) *Scheduler
//   func (s *Scheduler) Schedule(task Task) TaskID
//   func (s *Scheduler) Cancel(id TaskID) bool
//   func (s *Scheduler) Wait(id TaskID) error
//   func (s *Scheduler) Shutdown()
//   func (s *Scheduler) Stats() Stats
//
// Проверь:
//   go test -race -v ./...

package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Priority int

const (
	PriorityHigh   Priority = 3
	PriorityMedium Priority = 2
	PriorityLow    Priority = 1
)

type TaskID int64

type Task struct {
	ID       TaskID
	Priority Priority
	Deadline time.Time // нулевое = без дедлайна
	DependsOn []TaskID  // ID задач от которых зависим
	Fn       func(ctx context.Context) error
}

type taskState struct {
	task     Task
	ctx      context.Context
	cancel   context.CancelFunc
	done     chan struct{}
	err      error
	queuedAt time.Time
}

type Stats struct {
	Completed int64
	Failed    int64
	Cancelled int64
	Pending   int64
}

type Scheduler struct {
	mu      sync.Mutex
	tasks   map[TaskID]*taskState
	queue   []*taskState // упрощённо — обычный срез, сортируем по приоритету
	workers int
	jobs    chan *taskState
	wg      sync.WaitGroup
	stats   Stats
	nextID  atomic.Int64
	done    chan struct{}
}

// TODO: реализуй NewScheduler
func NewScheduler(workers int) *Scheduler {
	s := &Scheduler{
		tasks:   make(map[TaskID]*taskState),
		workers: workers,
		jobs:    make(chan *taskState, 100),
		done:    make(chan struct{}),
	}

	s.wg.Add(workers)
	for range workers {
		go s.worker()
	}

	return s
}

func (s *Scheduler) worker() {
	defer s.wg.Done()
	for {
		select {
		case ts, ok := <-s.jobs:
			if !ok {
				return
			}
			err := ts.task.Fn(ts.ctx)
			ts.err = err
			if err != nil {
				if ts.ctx.Err() != nil {
					atomic.AddInt64(&s.stats.Cancelled, 1)
				} else {
					atomic.AddInt64(&s.stats.Failed, 1)
				}
			} else {
				atomic.AddInt64(&s.stats.Completed, 1)
			}
			close(ts.done)
		case <-s.done:
			return
		}
	}
}

// TODO: реализуй Schedule — добавляет задачу в очередь
// Если у задачи есть DependsOn — ждём завершения всех зависимостей в горутине
func (s *Scheduler) Schedule(task Task) TaskID {
	if task.ID == 0 {
		task.ID = TaskID(s.nextID.Add(1))
	}

	ctx := context.Background()
	if !task.Deadline.IsZero() {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, task.Deadline)
		_ = cancel
	}

	ctx, cancel := context.WithCancel(ctx)
	ts := &taskState{
		task:     task,
		ctx:      ctx,
		cancel:   cancel,
		done:     make(chan struct{}),
		queuedAt: time.Now(),
	}

	s.mu.Lock()
	s.tasks[task.ID] = ts
	s.mu.Unlock()

	// Если есть зависимости — ждём в горутине
	if len(task.DependsOn) > 0 {
		go func() {
			for _, depID := range task.DependsOn {
				s.Wait(depID)
			}
			s.jobs <- ts
		}()
	} else {
		s.jobs <- ts
	}

	return task.ID
}

// TODO: реализуй Cancel
func (s *Scheduler) Cancel(id TaskID) bool {
	s.mu.Lock()
	ts, ok := s.tasks[id]
	s.mu.Unlock()
	if !ok {
		return false
	}
	ts.cancel()
	return true
}

// Wait блокируется до завершения задачи
func (s *Scheduler) Wait(id TaskID) error {
	s.mu.Lock()
	ts, ok := s.tasks[id]
	s.mu.Unlock()
	if !ok {
		return fmt.Errorf("задача %d не найдена", id)
	}
	<-ts.done
	return ts.err
}

// Shutdown останавливает планировщик
func (s *Scheduler) Shutdown() {
	close(s.done)
	close(s.jobs)
	s.wg.Wait()
}

func (s *Scheduler) Stats() Stats {
	return Stats{
		Completed: atomic.LoadInt64(&s.stats.Completed),
		Failed:    atomic.LoadInt64(&s.stats.Failed),
		Cancelled: atomic.LoadInt64(&s.stats.Cancelled),
	}
}

func main() {
	sched := NewScheduler(3)

	// Задача A
	idA := sched.Schedule(Task{
		Priority: PriorityHigh,
		Fn: func(ctx context.Context) error {
			time.Sleep(100 * time.Millisecond)
			fmt.Println("задача A выполнена")
			return nil
		},
	})

	// Задача B зависит от A
	idB := sched.Schedule(Task{
		Priority:  PriorityMedium,
		DependsOn: []TaskID{idA},
		Fn: func(ctx context.Context) error {
			fmt.Println("задача B выполнена (после A)")
			return nil
		},
	})

	// Задача C с дедлайном
	sched.Schedule(Task{
		Priority: PriorityLow,
		Deadline: time.Now().Add(50 * time.Millisecond),
		Fn: func(ctx context.Context) error {
			select {
			case <-time.After(200 * time.Millisecond):
				fmt.Println("задача C выполнена")
				return nil
			case <-ctx.Done():
				fmt.Println("задача C отменена по дедлайну")
				return ctx.Err()
			}
		},
	})

	sched.Wait(idB)
	time.Sleep(100 * time.Millisecond)

	stats := sched.Stats()
	fmt.Printf("Выполнено: %d, Ошибок: %d, Отменено: %d\n",
		stats.Completed, stats.Failed, stats.Cancelled)
}
