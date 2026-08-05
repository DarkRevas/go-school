// ============================================================
// Задача: Connection Pool  🔴 Senior
// ============================================================
//
// Вопрос с собесов уровня Senior.
//
// Реализуй пул соединений с базой данных:
//
//   type Pool struct { ... }
//
//   func NewPool(maxConn int, factory func() (Conn, error)) *Pool
//   func (p *Pool) Acquire(ctx context.Context) (Conn, error)
//   func (p *Pool) Release(conn Conn)
//   func (p *Pool) Close()
//   func (p *Pool) Stats() PoolStats
//
// Требования:
//   - Не более maxConn одновременных соединений
//   - Acquire блокируется пока нет свободного соединения
//   - Если ctx отменён во время ожидания — вернуть ctx.Err()
//   - Соединения переиспользуются (не создаём новое на каждый Acquire)
//   - Health check: если соединение "сломано" — создаём новое вместо него
//   - Close закрывает все незанятые соединения, ждёт возврата занятых
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

var ErrPoolClosed = errors.New("пул закрыт")
var ErrPoolExhausted = errors.New("пул исчерпан")

type Conn interface {
	Ping() error
	Close() error
	ID() int
}

type PoolStats struct {
	Total    int
	Idle     int
	InUse    int
	Acquired int64
	Released int64
}

type mockConn struct {
	id     int
	broken bool
}

func (c *mockConn) Ping() error {
	if c.broken {
		return errors.New("соединение сломано")
	}
	return nil
}
func (c *mockConn) Close() error {
	fmt.Printf("закрыто соединение %d\n", c.id)
	return nil
}
func (c *mockConn) ID() int { return c.id }

var connIDCounter atomic.Int32

type Pool struct {
	mu       sync.Mutex
	cond     *sync.Cond
	idle     []Conn
	inUse    int
	maxConn  int
	factory  func() (Conn, error)
	closed   bool
	acquired atomic.Int64
	released atomic.Int64
}

// TODO: реализуй NewPool
func NewPool(maxConn int, factory func() (Conn, error)) *Pool {
	p := &Pool{
		maxConn: maxConn,
		factory: factory,
	}
	p.cond = sync.NewCond(&p.mu)
	return p
}

func (p *Pool) takeIdleHealthy() (Conn, bool) {
	for len(p.idle) > 0 {
		conn := p.idle[len(p.idle)-1]
		p.idle = p.idle[:len(p.idle)-1]
		if conn.Ping() == nil {
			return conn, true
		}
		_ = conn.Close()
	}
	return nil, false
}

// TODO: реализуй Acquire
// Подсказка: три сценария: idle есть, можно создать, нужно ждать
// Отмена ctx должна разбудить ожидающего — подумай как
func (p *Pool) Acquire(ctx context.Context) (Conn, error) {
	for {
		p.mu.Lock()
		if p.closed {
			p.mu.Unlock()
			return nil, ErrPoolClosed
		}

		if conn, ok := p.takeIdleHealthy(); ok {
			p.inUse++
			p.mu.Unlock()
			p.acquired.Add(1)
			return conn, nil
		}

		if p.inUse+len(p.idle) < p.maxConn {
			p.inUse++
			factory := p.factory
			p.mu.Unlock()
			conn, err := factory()
			if err != nil {
				p.mu.Lock()
				p.inUse--
				p.cond.Signal()
				p.mu.Unlock()
				return nil, err
			}
			p.acquired.Add(1)
			return conn, nil
		}

		waitCh := make(chan struct{})
		go func() {
			select {
			case <-ctx.Done():
				p.mu.Lock()
				p.cond.Broadcast()
				p.mu.Unlock()
			case <-waitCh:
			}
		}()

		p.cond.Wait()
		close(waitCh)
		p.mu.Unlock()

		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
}

// TODO: реализуй Release
// Подсказка: после возврата нужно разбудить ожидающего
func (p *Pool) Release(conn Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		_ = conn.Close()
		p.inUse--
		p.cond.Broadcast()
		return
	}
	p.inUse--
	p.idle = append(p.idle, conn)
	p.released.Add(1)
	p.cond.Signal()
}

// TODO: реализуй Close
func (p *Pool) Close() {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.closed = true
	for _, c := range p.idle {
		_ = c.Close()
	}
	p.idle = nil
	p.cond.Broadcast()
	for p.inUse > 0 {
		p.cond.Wait()
	}
	p.mu.Unlock()
}

func (p *Pool) Stats() PoolStats {
	p.mu.Lock()
	defer p.mu.Unlock()
	return PoolStats{
		Total:    p.inUse + len(p.idle),
		Idle:     len(p.idle),
		InUse:    p.inUse,
		Acquired: p.acquired.Load(),
		Released: p.released.Load(),
	}
}

func main() {
	factory := func() (Conn, error) {
		id := int(connIDCounter.Add(1))
		fmt.Printf("создано соединение %d\n", id)
		return &mockConn{id: id}, nil
	}

	pool := NewPool(3, factory)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			conn, err := pool.Acquire(ctx)
			if err != nil {
				fmt.Printf("горутина %d: ошибка %v\n", n, err)
				return
			}
			fmt.Printf("горутина %d: соединение %d\n", n, conn.ID())
			time.Sleep(50 * time.Millisecond)
			pool.Release(conn)
		}(i)
	}

	wg.Wait()
	stats := pool.Stats()
	fmt.Printf("\nСтатистика: acquired=%d, released=%d, idle=%d\n",
		stats.Acquired, stats.Released, stats.Idle)
	pool.Close()
}
