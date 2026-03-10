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

// TODO: напиши метод Increment()

// TODO: напиши метод Value() int

func main() {
	// TODO: создай SafeCounter и запусти 1000 горутин
	// Каждая горутина вызывает counter.Increment()
	// После завершения всех горутин выведи counter.Value()

	_ = fmt.Println
	_ = sync.Mutex{}
}
