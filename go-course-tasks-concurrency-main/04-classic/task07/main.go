// ============================================================
// Задача: Readers-Writers без голодания  🔴 Senior
// ============================================================
//
// Классика на собесах. Отличается от задачи "WriterPriorityRWMutex" тем
// что здесь надо реализовать ДВА разных варианта и сравнить поведение.
//
// 1. Readers-preferring:
//    Читатели проходят всегда когда возможно. Писатель может голодать.
//
// 2. Fair (FIFO):
//    Порядок прихода соблюдается: если писатель пришёл раньше последующих
//    читателей — он проходит первым.
//
// Реализуй оба варианта с одинаковым интерфейсом:
//
//   type Lock interface {
//       RLock()
//       RUnlock()
//       Lock()
//       Unlock()
//   }
//
//   func NewReaderPreferring() Lock
//   func NewFair() Lock
//
// Задача не в коде — а в понимании семантики.
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

type Lock interface {
	RLock()
	RUnlock()
	Lock()
	Unlock()
}

// === Reader-preferring ===

type readerPref struct {
	mu        sync.Mutex
	readers   int
	writing   bool
	readCond  *sync.Cond
	writeCond *sync.Cond
}

// TODO: реализуй NewReaderPreferring
func NewReaderPreferring() Lock {
	l := &readerPref{}
	l.readCond = sync.NewCond(&l.mu)
	l.writeCond = sync.NewCond(&l.mu)
	return l
}

// TODO
func (l *readerPref) RLock() {
	l.mu.Lock()
	for l.writing {
		l.readCond.Wait()
	}
	l.readers++
	l.mu.Unlock()
}
func (l *readerPref) RUnlock() {
	l.mu.Lock()
	l.readers--
	if l.readers == 0 {
		l.writeCond.Signal()
	}
	l.mu.Unlock()
}

// TODO
func (l *readerPref) Lock() {
	l.mu.Lock()
	for l.writing || l.readers > 0 {
		l.writeCond.Wait()
	}
	l.writing = true
	l.mu.Unlock()
}
func (l *readerPref) Unlock() {
	l.mu.Lock()
	l.writing = false
	l.writeCond.Signal()
	l.readCond.Broadcast()
	l.mu.Unlock()
}

// === Fair (FIFO) ===

type fairWaiter struct {
	writer bool
	ready  chan struct{}
}

type fair struct {
	mu            sync.Mutex
	activeReaders int
	activeWriter  bool
	waiters       []*fairWaiter
}

// TODO: реализуй NewFair
func NewFair() Lock {
	return &fair{}
}

func (l *fair) wakeNext() {
	if len(l.waiters) == 0 {
		return
	}
	if l.waiters[0].writer {
		if l.activeReaders > 0 || l.activeWriter {
			return
		}
		w := l.waiters[0]
		l.waiters = l.waiters[1:]
		l.activeWriter = true
		close(w.ready)
		return
	}
	for len(l.waiters) > 0 && !l.waiters[0].writer {
		w := l.waiters[0]
		l.waiters = l.waiters[1:]
		l.activeReaders++
		close(w.ready)
	}
}

// TODO
func (l *fair) RLock() {
	l.mu.Lock()
	if !l.activeWriter && len(l.waiters) == 0 {
		l.activeReaders++
		l.mu.Unlock()
		return
	}
	w := &fairWaiter{ready: make(chan struct{})}
	l.waiters = append(l.waiters, w)
	l.mu.Unlock()
	<-w.ready
}
func (l *fair) RUnlock() {
	l.mu.Lock()
	l.activeReaders--
	if l.activeReaders == 0 {
		l.wakeNext()
	}
	l.mu.Unlock()
}

// TODO
func (l *fair) Lock() {
	l.mu.Lock()
	if l.activeReaders == 0 && !l.activeWriter && len(l.waiters) == 0 {
		l.activeWriter = true
		l.mu.Unlock()
		return
	}
	w := &fairWaiter{writer: true, ready: make(chan struct{})}
	l.waiters = append(l.waiters, w)
	l.mu.Unlock()
	<-w.ready
}
func (l *fair) Unlock() {
	l.mu.Lock()
	l.activeWriter = false
	l.wakeNext()
	l.mu.Unlock()
}

// === Демо ===

func demo(name string, l Lock) {
	fmt.Printf("\n=== %s ===\n", name)
	var reads, writes atomic.Int32
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 5 {
				l.RLock()
				reads.Add(1)
				time.Sleep(3 * time.Millisecond)
				l.RUnlock()
			}
		}()
	}

	// Писатель посреди потока читателей
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(5 * time.Millisecond)
		t0 := time.Now()
		l.Lock()
		fmt.Printf("писатель зашёл через %v\n", time.Since(t0))
		time.Sleep(10 * time.Millisecond)
		writes.Add(1)
		l.Unlock()
	}()

	wg.Wait()
	fmt.Printf("reads=%d writes=%d\n", reads.Load(), writes.Load())
}

func main() {
	demo("reader-preferring", NewReaderPreferring())
	demo("fair", NewFair())
}
