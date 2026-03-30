// ============================================================
// Задача: Очередь с ожиданием через sync.Cond  🔴 Senior
// ============================================================
//
// Вопрос с собесов уровня Senior.
//
// Реализуй блокирующую очередь (blocking queue):
//
//   type BlockingQueue[T any] struct { ... }
//
//   func NewBlockingQueue[T any](capacity int) *BlockingQueue[T]
//   func (q *BlockingQueue[T]) Put(item T)   // блокируется если очередь полна
//   func (q *BlockingQueue[T]) Take() T      // блокируется если очередь пуста
//   func (q *BlockingQueue[T]) PutTimeout(item T, d time.Duration) bool
//   func (q *BlockingQueue[T]) TakeTimeout(d time.Duration) (T, bool)
//   func (q *BlockingQueue[T]) Len() int
//   func (q *BlockingQueue[T]) Close()       // пробуждает все заблокированные горутины
//
// Используй sync.Cond (не каналы!).
//
// Зачем sync.Cond вместо каналов?
//   - Каналы имеют фиксированный тип и не поддерживают broadcast
//   - sync.Cond позволяет гибко управлять условиями пробуждения
//   - Используется в стандартной библиотеке (sync.Pool, etc.)
//
// Проверь:
//   go test -race -v ./...

package main

import (
	"fmt"
	"sync"
	"time"
)

type BlockingQueue[T any] struct {
	mu       sync.Mutex
	notFull  *sync.Cond
	notEmpty *sync.Cond
	items    []T
	cap      int
	closed   bool
}

// TODO: реализуй NewBlockingQueue
func NewBlockingQueue[T any](capacity int) *BlockingQueue[T] {
	q := &BlockingQueue[T]{
		items: make([]T, 0, capacity),
		cap:   capacity,
	}
	q.notFull = sync.NewCond(&q.mu)
	q.notEmpty = sync.NewCond(&q.mu)
	return q
}

// TODO: реализуй Put — блокируется пока len(items) == cap
func (q *BlockingQueue[T]) Put(item T) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for len(q.items) == q.cap && !q.closed {
		q.notFull.Wait()
	}
	if q.closed {
		return
	}
	q.items = append(q.items, item)
	q.notEmpty.Signal()
}

// TODO: реализуй Take — блокируется пока len(items) == 0
func (q *BlockingQueue[T]) Take() (zero T) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for len(q.items) == 0 && !q.closed {
		q.notEmpty.Wait()
	}
	if len(q.items) == 0 {
		return zero
	}
	item := q.items[0]
	q.items = q.items[1:]
	q.notFull.Signal()
	return item
}

// TODO: реализуй PutTimeout
// Подсказка: запусти горутину с таймером которая вызывает notFull.Broadcast()
func (q *BlockingQueue[T]) PutTimeout(item T, d time.Duration) bool {
	timer := time.AfterFunc(d, func() {
		q.mu.Lock()
		q.notFull.Broadcast()
		q.mu.Unlock()
	})
	defer timer.Stop()

	q.mu.Lock()
	defer q.mu.Unlock()

	deadline := time.Now().Add(d)
	for len(q.items) == q.cap && !q.closed {
		if time.Now().After(deadline) {
			return false
		}
		q.notFull.Wait()
	}
	if q.closed {
		return false
	}
	q.items = append(q.items, item)
	q.notEmpty.Signal()
	return true
}

// TODO: реализуй TakeTimeout аналогично
func (q *BlockingQueue[T]) TakeTimeout(d time.Duration) (zero T, ok bool) {
	timer := time.AfterFunc(d, func() {
		q.mu.Lock()
		q.notEmpty.Broadcast()
		q.mu.Unlock()
	})
	defer timer.Stop()

	q.mu.Lock()
	defer q.mu.Unlock()

	deadline := time.Now().Add(d)
	for len(q.items) == 0 && !q.closed {
		if time.Now().After(deadline) {
			return zero, false
		}
		q.notEmpty.Wait()
	}
	if len(q.items) == 0 {
		return zero, false
	}
	item := q.items[0]
	q.items = q.items[1:]
	q.notFull.Signal()
	return item, true
}

func (q *BlockingQueue[T]) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}

// Close закрывает очередь и пробуждает все заблокированные горутины
func (q *BlockingQueue[T]) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.closed = true
	q.notFull.Broadcast()
	q.notEmpty.Broadcast()
}

func main() {
	q := NewBlockingQueue[int](3)

	// Производитель
	go func() {
		for i := 0; i < 10; i++ {
			q.Put(i)
			fmt.Printf("Положено: %d, в очереди: %d\n", i, q.Len())
			time.Sleep(50 * time.Millisecond)
		}
		q.Close()
	}()

	// Потребитель (медленный)
	for {
		v, ok := q.TakeTimeout(500 * time.Millisecond)
		if !ok {
			fmt.Println("Очередь закрыта или таймаут")
			break
		}
		fmt.Printf("Взято: %d\n", v)
		time.Sleep(100 * time.Millisecond)
	}
}
