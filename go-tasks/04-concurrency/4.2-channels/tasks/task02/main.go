// Задание 2: Fan-In - слить два канала
//
// Создай две функции-генератора:
//   - evenNumbers() <-chan int  - генерирует чётные числа 2, 4, 6, 8, 10
//   - oddNumbers() <-chan int   - генерирует нечётные числа 1, 3, 5, 7, 9
//
// Каждая функция запускает горутину, которая кладёт числа в канал и закрывает его.
//
// Напиши функцию merge(ch1, ch2 <-chan int) <-chan int,
// которая читает из обоих каналов и возвращает один объединённый канал.
//
// В main() слей оба генератора и выведи все числа в одну строку.
//
// Ожидаемый вывод (порядок может быть любым):
//   2 1 4 3 6 5 8 7 10 9
//   (или любой другой порядок - главное все 10 чисел)
//
// Запусти: go run main.go

package main

import (
	"fmt"
	"sync"
)

// TODO: напиши функцию evenNumbers() <-chan int
func evenNumbers() <-chan int {
	ch := make(chan int, 5)

	go func() {
		for i := 2; i <= 10; i += 2 {
			ch <- i
		}
		close(ch)
	}()
	
	return ch
}

// TODO: напиши функцию oddNumbers() <-chan int
func oddNumbers() <-chan int {
	ch := make(chan int, 5)

	go func() {
		for i := 1; i <= 10; i += 2 {
			ch <- i
		}
		close(ch)
	}()

	return ch
}

// TODO: напиши функцию merge(ch1, ch2 <-chan int) <-chan int
// Подсказка: используй WaitGroup и отдельную горутину для закрытия merged

func merge(ch1, ch2 <-chan int) <-chan int {
	var wg sync.WaitGroup
	ch := make(chan int, 10)

	wg.Add(2)

	go func() {
		defer wg.Done() 
		for el := range ch1 {
			ch <- el
		}
	}()

	go func() {
		defer wg.Done() 
		for el := range ch2 {
			ch <- el
		}
	}()

	go func() {
		wg.Wait()
		close(ch)
	}()

	return ch
}

func main() {
	// TODO: создай два канала через генераторы
	// TODO: слей их через merge
	// TODO: выведи все числа в одну строку
	ch := merge(oddNumbers(),evenNumbers()) // Вывод всегда 1 3 5 7 9 2 4 6 8 10 - это правильно?

	for el := range ch {
		fmt.Print(el, " ")
	}
}
