// ============================================================
// Задача: Circuit Breaker  🔴 Senior
// ============================================================
//
// Паттерн для защиты от лавинообразных падений при проблемах
// у внешнего сервиса.
//
// Три состояния:
//   Closed    — всё ок, запросы пропускаются, считаем последовательные ошибки
//   Open      — запросы моментально фейлятся (не идём в бэкенд) до таймаута
//   HalfOpen  — пробный запрос: если успешен — возвращаемся в Closed;
//               если падает — снова Open
//
// Интерфейс:
//
//   type CircuitBreaker struct { ... }
//
//   func New(failureThreshold int, openTimeout time.Duration) *CircuitBreaker
//   func (cb *CircuitBreaker) Call(fn func() error) error
//   func (cb *CircuitBreaker) State() State
//
// Требования:
//   - Потокобезопасен (много параллельных Call)
//   - В состоянии Open возвращает ErrOpen сразу, не вызывая fn
//   - В HalfOpen пропускает ТОЛЬКО один пробный Call одновременно
//   - Успех в HalfOpen → Closed, счётчик ошибок сбрасывается
//   - Ошибка в HalfOpen → снова Open, таймер открытия продлевается
//
// Проверь:
//   go test -race -v ./...

package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var ErrOpen = errors.New("circuit breaker: открыт")

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

func (s State) String() string {
	return [...]string{"Closed", "Open", "HalfOpen"}[s]
}

type CircuitBreaker struct {
	mu               sync.Mutex
	state            State
	failures         int
	failureThreshold int
	openTimeout      time.Duration
	openedAt         time.Time
	probeInFlight    bool
}

// TODO: реализуй конструктор
func New(failureThreshold int, openTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:            StateClosed,
		failureThreshold: failureThreshold,
		openTimeout:      openTimeout,
	}
}

func (cb *CircuitBreaker) maybeTransitionLocked() {
	if cb.state == StateOpen && time.Since(cb.openedAt) >= cb.openTimeout {
		cb.state = StateHalfOpen
		cb.probeInFlight = false
	}
}

// TODO: реализуй Call
// Подсказка: проверь state перед вызовом fn; после вызова — обнови состояние по результату
// Отдельная сложность — переход Open → HalfOpen по времени (не через таймер, а лениво при Call)
func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.Lock()
	cb.maybeTransitionLocked()

	switch cb.state {
	case StateOpen:
		cb.mu.Unlock()
		return ErrOpen
	case StateHalfOpen:
		if cb.probeInFlight {
			cb.mu.Unlock()
			return ErrOpen
		}
		cb.probeInFlight = true
	}
	cb.mu.Unlock()

	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateHalfOpen:
		cb.probeInFlight = false
		if err != nil {
			cb.state = StateOpen
			cb.openedAt = time.Now()
			return err
		}
		cb.state = StateClosed
		cb.failures = 0
		return nil
	case StateClosed:
		if err != nil {
			cb.failures++
			if cb.failures >= cb.failureThreshold {
				cb.state = StateOpen
				cb.openedAt = time.Now()
			}
			return err
		}
		cb.failures = 0
		return nil
	default:
		return err
	}
}

// TODO: реализуй State — просто читаем текущее состояние (не забудь про ленивый переход Open → HalfOpen)
func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.maybeTransitionLocked()
	return cb.state
}

func main() {
	cb := New(3, 100*time.Millisecond)

	failingCall := func() error { return errors.New("backend error") }
	goodCall := func() error { return nil }

	// Роняем до Open
	for i := 0; i < 5; i++ {
		err := cb.Call(failingCall)
		fmt.Printf("попытка %d: state=%s err=%v\n", i+1, cb.State(), err)
	}

	// В Open — мгновенный ErrOpen
	err := cb.Call(goodCall)
	fmt.Printf("в Open: state=%s err=%v\n", cb.State(), err)

	// Ждём openTimeout — следующий Call должен быть probe (HalfOpen)
	time.Sleep(150 * time.Millisecond)
	err = cb.Call(goodCall)
	fmt.Printf("probe: state=%s err=%v\n", cb.State(), err) // должен быть Closed, err=nil
}
