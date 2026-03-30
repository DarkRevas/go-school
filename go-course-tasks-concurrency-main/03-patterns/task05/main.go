// ============================================================
// Задача: Singleflight — дедупликация запросов  🔴 Senior
// ============================================================
//
// Вопрос с собесов уровня Senior.
//
// Thundering Herd проблема: 1000 горутин одновременно идут за одним ключом в кеш.
// Кеш пуст — 1000 запросов идут в БД. БД падает.
//
// Решение: Singleflight — если есть параллельные запросы с одним ключом,
// выполняется только ОДИН, остальные ждут его результата.
//
// Реализуй без singleflight из стандартной библиотеки:
//
//   type Group[T any] struct { ... }
//
//   func (g *Group[T]) Do(key string, fn func() (T, error)) (T, error, bool)
//   // bool = true если этот вызов был "дедуплицирован" (получил чужой результат)
//
// Проверь что реально выполняется только 1 вызов fn:
//   go test -race -v ./...

package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type call[T any] struct {
	wg  sync.WaitGroup
	val T
	err error
}

type Group[T any] struct {
	mu    sync.Mutex
	calls map[string]*call[T]
}

// TODO: реализуй Do
// Алгоритм:
//   1. Lock, проверь есть ли уже вызов с этим ключом
//   2. Если есть — Unlock, wg.Wait(), вернуть результат с shared=true
//   3. Если нет — создай call, добавь в map, Unlock
//   4. Выполни fn()
//   5. Сохрани результат, wg.Done(), удали из map
//   6. Вернуть результат с shared=false
func (g *Group[T]) Do(key string, fn func() (T, error)) (T, error, bool) {
	g.mu.Lock()

	if g.calls == nil {
		g.calls = make(map[string]*call[T])
	}

	if c, ok := g.calls[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err, true // shared
	}

	c := &call[T]{}
	c.wg.Add(1)
	g.calls[key] = c
	g.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.calls, key)
	g.mu.Unlock()

	return c.val, c.err, false
}

func main() {
	var group Group[string]
	var actualCalls atomic.Int32
	const concurrency = 100

	fetch := func() (string, error) {
		actualCalls.Add(1)
		time.Sleep(50 * time.Millisecond) // имитируем запрос к БД
		return "данные из БД", nil
	}

	var wg sync.WaitGroup
	shared := atomic.Int32{}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			val, _, isShared := group.Do("user:42", fetch)
			if isShared {
				shared.Add(1)
			}
			_ = val
		}()
	}
	wg.Wait()

	fmt.Printf("Запросов к БД: %d (из %d горутин)\n", actualCalls.Load(), concurrency)
	fmt.Printf("Дедуплицировано: %d запросов\n", shared.Load())
	// Ожидаем: Запросов к БД: 1
}
