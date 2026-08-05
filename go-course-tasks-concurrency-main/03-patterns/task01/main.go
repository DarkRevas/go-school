// ============================================================
// Задача: Worker Pool  🟡 Middle
// ============================================================
//
// Один из самых частых вопросов на собесах в любой Go-компании.
//
// Реализуй WorkerPool:
//
//   type WorkerPool struct { ... }
//
//   func NewWorkerPool(workers int) *WorkerPool
//   func (p *WorkerPool) Submit(task func()) bool  // false если пул остановлен
//   func (p *WorkerPool) Stop()                    // graceful: ждёт завершения текущих задач
//   func (p *WorkerPool) StopNow()                 // немедленная остановка
//   func (p *WorkerPool) Running() int             // количество активных воркеров
//
// Требования:
//   - Фиксированный пул из workers горутин
//   - Submit не блокируется (очередь задач буферизована, размер 100)
//   - Stop ждёт пока все отправленные задачи завершатся
//   - StopNow отменяет незапущенные задачи, ждёт текущие
//   - Нет утечек горутин
//
// Проверь:
//   go test -race -v ./...

package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type WorkerPool struct {
	jobs    chan func()
	wg      sync.WaitGroup
	once    sync.Once
	running atomic.Int32
	stopped atomic.Bool
}

// TODO: реализуй NewWorkerPool
func NewWorkerPool(workers int) *WorkerPool {
	p := &WorkerPool{
		jobs: make(chan func(), 100),
	}
	p.wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer p.wg.Done()
			for task := range p.jobs {
				p.running.Add(1)
				task()
				p.running.Add(-1)
			}
		}()
	}
	return p
}

// TODO: реализуй Submit
// Подсказка: что если очередь уже полна или пул остановлен?
func (p *WorkerPool) Submit(task func()) bool {
	if p.stopped.Load() {
		return false
	}
	defer func() { recover() }()
	select {
	case p.jobs <- task:
		return true
	default:
		return false
	}
}

// TODO: Stop ждёт завершения всех задач
func (p *WorkerPool) Stop() {
	p.once.Do(func() {
		p.stopped.Store(true)
		close(p.jobs)
	})
	p.wg.Wait()
}

// TODO: StopNow немедленная остановка — дропает незапущенные задачи вместо ожидания
func (p *WorkerPool) StopNow() {
	p.once.Do(func() {
		p.stopped.Store(true)
		close(p.jobs)
	})
	for range p.jobs {
	}
	p.wg.Wait()
}

func (p *WorkerPool) Running() int {
	return int(p.running.Load())
}

func main() {
	pool := NewWorkerPool(3)

	var mu sync.Mutex
	var results []int

	for i := 0; i < 10; i++ {
		n := i
		pool.Submit(func() {
			time.Sleep(50 * time.Millisecond)
			mu.Lock()
			results = append(results, n)
			mu.Unlock()
			fmt.Printf("задача %d выполнена\n", n)
		})
	}

	pool.Stop()
	fmt.Printf("Выполнено задач: %d\n", len(results))
}
