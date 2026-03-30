// ============================================================
// Задача: Найди и исправь дедлоки  🔴 Senior
// ============================================================
//
// Вопрос-ловушка на собесах уровня Senior.
// "Что не так в этом коде?"
//
// Здесь 4 независимых сценария с дедлоками. Найди каждый и исправь.
//
// Запуск:
//   go run main.go           ← укажи номер сценария: go run main.go 1
//   go run -race main.go 4   ← для сценария с гонкой

package main

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

// ============================================================
// Сценарий 1: Взаимный захват мьютексов
// ============================================================
// Горутина A: Lock(mu1) → Lock(mu2)
// Горутина B: Lock(mu2) → Lock(mu1)
// → дедлок!
//
// TODO: исправь порядок захвата мьютексов

func scenario1() {
	var mu1, mu2 sync.Mutex
	done := make(chan struct{})

	go func() {
		mu1.Lock()
		time.Sleep(1 * time.Millisecond) // даём горутине B время захватить mu2
		mu2.Lock() // ← дедлок
		fmt.Println("A: захватил оба мьютекса")
		mu2.Unlock()
		mu1.Unlock()
		close(done)
	}()

	mu2.Lock()
	time.Sleep(1 * time.Millisecond)
	mu1.Lock() // ← дедлок
	fmt.Println("B: захватил оба мьютекса")
	mu1.Unlock()
	mu2.Unlock()

	<-done
}

// ============================================================
// Сценарий 2: Горутина пишет в небуферизованный канал, читателя нет
// ============================================================
// TODO: исправь (буферизованный канал или читатель в горутине)

func scenario2() {
	ch := make(chan int) // ← небуферизованный
	ch <- 42            // ← блокируется навсегда — никто не читает
	fmt.Println("sent:", <-ch)
}

// ============================================================
// Сценарий 3: Lock внутри Lock (рекурсивный мьютекс)
// ============================================================
// sync.Mutex НЕ рекурсивный! Повторный Lock из той же горутины → дедлок.
// TODO: исправь через unlock перед рекурсивным вызовом или через отдельный метод

type SafeCounter struct {
	mu    sync.Mutex
	count int
}

func (c *SafeCounter) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count++
}

func (c *SafeCounter) IncAndLog() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count++
	// TODO: эта строка вызывает дедлок! Inc() тоже берёт mu
	c.Inc() // ← дедлок: рекурсивный Lock
}

func scenario3() {
	c := &SafeCounter{}
	c.IncAndLog() // дедлок
	fmt.Println("count:", c.count)
}

// ============================================================
// Сценарий 4: WaitGroup — Add после запуска горутин
// ============================================================
// TODO: перенеси wg.Add(1) ДО go func()

func scenario4() {
	var wg sync.WaitGroup
	results := make([]int, 5)

	for i := 0; i < 5; i++ {
		n := i
		go func() {
			wg.Add(1) // ← НЕПРАВИЛЬНО: Add после старта горутины
			defer wg.Done()
			time.Sleep(10 * time.Millisecond)
			results[n] = n * n
		}()
	}

	wg.Wait() // может завершиться до того как все горутины добавились
	fmt.Println(results)
}

func main() {
	scenario := 1
	if len(os.Args) > 1 {
		scenario, _ = strconv.Atoi(os.Args[1])
	}

	fmt.Printf("=== Сценарий %d ===\n", scenario)
	fmt.Println("ВНИМАНИЕ: этот код содержит дедлок. Запусти и изучи трейсбек.")
	fmt.Println("Затем найди и исправь проблему в коде.")
	fmt.Println()

	switch scenario {
	case 1:
		scenario1()
	case 2:
		scenario2()
	case 3:
		scenario3()
	case 4:
		scenario4()
	default:
		fmt.Println("Укажи номер сценария: go run main.go [1-4]")
	}
}
