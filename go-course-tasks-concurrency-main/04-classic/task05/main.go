// ============================================================
// Задача: Building H2O — LeetCode 1117  🔴 Senior
// ============================================================
//
// Задача с LeetCode. Частый вопрос на собесах уровня Senior.
//
// Есть два типа горутин: "водород" H и "кислород" O.
// Они должны формировать молекулы воды H2O строго:
//   - каждая молекула = 2 атома H + 1 атом O
//   - все три атома должны "встретиться" прежде чем какой-либо из них пройдёт дальше
//
// Реализуй:
//   type H2O struct { ... }
//   func NewH2O() *H2O
//   func (w *H2O) Hydrogen(fn func())   // fn() = "связаться в молекулу"
//   func (w *H2O) Oxygen(fn func())
//
// Ожидаемый вывод (буквы в группах по 3, каждая группа = HOO нет, = HHO или OHH):
//   OHH HHO OHH...  (по 2 H и 1 O в каждой тройке)
//
// Проверь:
//   go test -race -v ./...

package main

import (
	"fmt"
	"strings"
	"sync"
)

type H2O struct {
	hSem chan struct{} // пропускает до 2 водородов
	oSem chan struct{} // пропускает 1 кислород
	bar  *barrier
}

// barrier — встреча трёх горутин перед продолжением
type barrier struct {
	mu         sync.Mutex
	cond       *sync.Cond
	count      int
	total      int
	generation int
}

func newBarrier(n int) *barrier {
	b := &barrier{total: n}
	b.cond = sync.NewCond(&b.mu)
	return b
}

// TODO: реализуй Wait
// Подсказка: последний пришедший разбуждает всех, остальные ждут
func (b *barrier) Wait() {
	b.mu.Lock()
	gen := b.generation
	b.count++
	if b.count == b.total {
		b.count = 0
		b.generation++
		b.cond.Broadcast()
		b.mu.Unlock()
		return
	}
	for gen == b.generation {
		b.cond.Wait()
	}
	b.mu.Unlock()
}

// TODO: реализуй NewH2O
// Подсказка: семафоры ограничивают сколько атомов каждого типа собирается в одну "встречу",
// а барьер синхронизирует их — подумай какие ёмкости нужны для H и O
func NewH2O() *H2O {
	h := &H2O{
		hSem: make(chan struct{}, 2),
		oSem: make(chan struct{}, 1),
		bar:  newBarrier(3),
	}
	h.hSem <- struct{}{}
	h.hSem <- struct{}{}
	h.oSem <- struct{}{}
	return h
}

// TODO: реализуй Hydrogen
func (w *H2O) Hydrogen(fn func()) {
	<-w.hSem
	w.bar.Wait()
	fn()
	w.hSem <- struct{}{}
}

// TODO: реализуй Oxygen
func (w *H2O) Oxygen(fn func()) {
	<-w.oSem
	w.bar.Wait()
	fn()
	w.oSem <- struct{}{}
}

func main() {
	h2o := NewH2O()
	var mu sync.Mutex
	var result strings.Builder
	var wg sync.WaitGroup

	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h2o.Hydrogen(func() { mu.Lock(); result.WriteRune('H'); mu.Unlock() })
		}()
	}
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h2o.Oxygen(func() { mu.Lock(); result.WriteRune('O'); mu.Unlock() })
		}()
	}

	wg.Wait()
	fmt.Println(result.String()) // должно быть 2 молекулы H2O
}
