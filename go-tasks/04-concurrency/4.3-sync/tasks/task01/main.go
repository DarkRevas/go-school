// Задание 1: Безопасный счётчик через Mutex
//
// Создай структуру SafeCounter с:
//   - полем mu типа sync.Mutex
//   - полем value типа int
//   - методом Increment() - увеличивает value на 1 (с блокировкой)
//   - методом Value() int - возвращает value (с блокировкой)
//
// Запусти 1000 горутин, каждая вызывает counter.Increment().
// После wg.Wait() выведи финальное значение.
//
// Проверь отсутствие гонок:
//   go run -race main.go
//
// Ожидаемый вывод:
//   Финальный счётчик: 1000
//   (и никаких WARNING: DATA RACE)
//
// Запусти: go run main.go

package main

import (
	"fmt"
	"sync"
)

// TODO: напиши структуру SafeCounter
// type SafeCounter struct { ... }
type SafeCounter struct {
	value int
	mu sync.Mutex
	wg sync.WaitGroup
}

// TODO: напиши метод Increment()
func (s *SafeCounter) Increment() {
	defer s.wg.Done()
	defer s.mu.Unlock()
	s.mu.Lock()
	s.value++
}

// TODO: напиши метод Value() int
func (s *SafeCounter) Value() int {
	defer s.mu.Unlock()
	s.mu.Lock()
	return s.value
}

func main() {
	// TODO: создай SafeCounter и запусти 1000 горутин
	// Каждая горутина вызывает counter.Increment()
	// После завершения всех горутин выведи counter.Value()
	counter := SafeCounter{
		value: 0,
		mu: sync.Mutex{},
		wg: sync.WaitGroup{},
	}

	for range 1000 {
		counter.wg.Add(1)
		go counter.Increment()
	}

	counter.wg.Wait()

	fmt.Printf("Финальный счётчик: %v", counter.value) // Финальный счётчик: 1000
}
