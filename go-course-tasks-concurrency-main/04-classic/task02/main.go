// ============================================================
// Задача: Producer-Consumer с bounded buffer  🟡 Middle
// ============================================================
//
// Классика на собесах Junior/Middle уровня.
//
// Реализуй через каналы:
//   - M производителей генерируют числа 0..N
//   - K потребителей читают, возводят в квадрат, пишут в results
//   - Буфер между ними ограничен (размер B)
//
// Требования:
//   - Потребители завершаются когда производители закончили И буфер пуст
//   - Нет утечек горутин
//   - Все числа должны быть обработаны ровно один раз
//
// Реализуй ДВА варианта:
//   1. Через каналы (идиоматично в Go)
//   2. Через sync.Cond (для понимания классических примитивов)
//
// Проверь:
//   go test -race -v ./...

package main

import (
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
)

// === Вариант 1: через каналы ===

// TODO: реализуй producerConsumerChan
// Подсказка: два буферизованных канала и два WaitGroup — для производителей и потребителей
func producerConsumerChan(producers, consumers, n, bufSize int) []int {
	jobs := make(chan int, bufSize)
	resultsCh := make(chan int, n)

	var next atomic.Int64
	var prodWg sync.WaitGroup
	for i := 0; i < producers; i++ {
		prodWg.Add(1)
		go func() {
			defer prodWg.Done()
			for {
				v := int(next.Add(1) - 1)
				if v >= n {
					return
				}
				jobs <- v
			}
		}()
	}
	go func() {
		prodWg.Wait()
		close(jobs)
	}()

	var consWg sync.WaitGroup
	for i := 0; i < consumers; i++ {
		consWg.Add(1)
		go func() {
			defer consWg.Done()
			for v := range jobs {
				resultsCh <- v * v
			}
		}()
	}
	go func() {
		consWg.Wait()
		close(resultsCh)
	}()

	out := make([]int, 0, n)
	for r := range resultsCh {
		out = append(out, r)
	}
	return out
}

// === Вариант 2: через sync.Cond ===

// TODO: реализуй producerConsumerCond
// Подсказка: буфер — обычный срез; производители ждут пока буфер полон, потребители — пока пуст
func producerConsumerCond(producers, consumers, n, bufSize int) []int {
	var mu sync.Mutex
	notEmpty := sync.NewCond(&mu)
	notFull := sync.NewCond(&mu)
	buf := make([]int, 0, bufSize)
	done := false

	var next atomic.Int64
	var resultsMu sync.Mutex
	results := make([]int, 0, n)

	var prodWg sync.WaitGroup
	for i := 0; i < producers; i++ {
		prodWg.Add(1)
		go func() {
			defer prodWg.Done()
			for {
				v := int(next.Add(1) - 1)
				if v >= n {
					return
				}
				mu.Lock()
				for len(buf) == bufSize {
					notFull.Wait()
				}
				buf = append(buf, v)
				notEmpty.Signal()
				mu.Unlock()
			}
		}()
	}

	go func() {
		prodWg.Wait()
		mu.Lock()
		done = true
		notEmpty.Broadcast()
		mu.Unlock()
	}()

	var consWg sync.WaitGroup
	for i := 0; i < consumers; i++ {
		consWg.Add(1)
		go func() {
			defer consWg.Done()
			for {
				mu.Lock()
				for len(buf) == 0 && !done {
					notEmpty.Wait()
				}
				if len(buf) == 0 {
					mu.Unlock()
					return
				}
				v := buf[0]
				buf = buf[1:]
				notFull.Signal()
				mu.Unlock()

				resultsMu.Lock()
				results = append(results, v*v)
				resultsMu.Unlock()
			}
		}()
	}
	consWg.Wait()
	return results
}

func main() {
	results := producerConsumerChan(2, 3, 10, 3)
	sort.Ints(results)
	fmt.Println("Результаты:", results)
}
