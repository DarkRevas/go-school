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
	"testing"
)

// === Вариант 1: через каналы ===

func producerConsumerChan(producers, consumers, n, bufSize int) []int {
	jobs := make(chan int, bufSize)
	results := make(chan int, n)

	// Производители
	var prodWg sync.WaitGroup
	for p := 0; p < producers; p++ {
		prodWg.Add(1)
		start := p * (n / producers)
		end := start + (n / producers)
		if p == producers-1 {
			end = n
		}
		go func(from, to int) {
			defer prodWg.Done()
			for i := from; i < to; i++ {
				jobs <- i
			}
		}(start, end)
	}

	// Закрываем jobs когда все производители закончили
	go func() {
		prodWg.Wait()
		close(jobs)
	}()

	// Потребители
	var consWg sync.WaitGroup
	for c := 0; c < consumers; c++ {
		consWg.Add(1)
		go func() {
			defer consWg.Done()
			for job := range jobs {
				results <- job * job
			}
		}()
	}

	// Закрываем results когда все потребители закончили
	go func() {
		consWg.Wait()
		close(results)
	}()

	var out []int
	for r := range results {
		out = append(out, r)
	}
	return out
}

func TestProducerConsumer(t *testing.T) {
	results := producerConsumerChan(3, 4, 20, 5)
	sort.Ints(results)

	if len(results) != 20 {
		t.Fatalf("ожидали 20 результатов, получили %d", len(results))
	}

	// Проверяем что это квадраты чисел 0..19
	for i, v := range results {
		want := i * i
		if v != want {
			t.Errorf("[%d] = %d, want %d", i, v, want)
		}
	}
}

func main() {
	results := producerConsumerChan(2, 3, 10, 3)
	sort.Ints(results)
	fmt.Println("Результаты:", results)
}
