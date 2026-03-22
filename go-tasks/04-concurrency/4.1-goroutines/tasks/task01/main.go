// Задание 1: Параллельная обработка задач
//
// Есть список задач (числа от 1 до 10).
// Запусти каждую задачу в отдельной горутине.
// Каждая горутина должна:
//   1. Подождать случайное время от 50 до 200мс (имитация работы)
//   2. Вывести: "Задача N выполнена за Xms"
//
// Дождись завершения ВСЕХ горутин через WaitGroup, потом выведи:
//   "Все 10 задач выполнены!"
//
// Запусти несколько раз - порядок вывода будет меняться, это нормально.
//
// Запусти: go run main.go

package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	// TODO: объяви WaitGroup
	var wg sync.WaitGroup

	// Тут рандомный вывод

	for i := 1; i <= 10; i++ {
		// TODO: увеличь счётчик wg на 1
		wg.Add(1)
		// TODO: запусти горутину, передав i как параметр
		// Внутри горутины:
		//   - вызови defer wg.Done()
		//   - сгенерируй случайное время: rand.Intn(150)+50 миллисекунд
		//   - подожди это время через time.Sleep
		//   - выведи результат
		go func(i int) {
			defer wg.Done()
			start := time.Now()
			time.Sleep(time.Duration(rand.Intn(150)+50))
			fmt.Printf("Задача %v выполнена за %vms \n", i, time.Since(start))
		}(i)
	}

	// TODO: дождись всех горутин через wg.Wait()
	wg.Wait()

	fmt.Printf("----------------------------------------\n")

	// Тут гарантированный порядок задач

	numGorutines := 10
	sl := make([]string, numGorutines)

	for i := range numGorutines {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			start := time.Now()
			time.Sleep(time.Duration(rand.Intn(150)+50))
			sl[i] = fmt.Sprintf("Задача %v выполнена за %vms \n", i, time.Since(start))
		}(i)
	}
	
	wg.Wait()

	for _, el := range sl {
		fmt.Println(el)
	}
}
