package main

import "slices"

import "fmt"

func Echo[T any](v T) T {
	return v
}

func First[T any](items []T) (T, bool) {
	var zero T
	if len(items) == 0 {
		return zero, false 
	}
	return items[0], true
}

type Number interface {
	~int | ~float64
}

func Sum[T Number](items []T) T {
	var result T
	for _, el := range items {
		result += el
	}
	return result
}

func IndexOf[T comparable](items []T, target T) int {
	for i, el := range items {
		if el == target {
			return i
		}
	}

	return -1
}

func Values[K comparable, V any](m map[K]V) []V {
	result := make([]V, 0, len(m))
	for _,v := range m {
		result = append(result, v)
	}
	return result
}

type Store[T any] struct {
	items []T
}

func (s *Store[T]) Add(item T) {
	s.items = append(s.items, item)
}

func (s *Store[T]) All() []T {
	return s.items
}

func Contains[T comparable](items []T, target T) bool {
	return slices.Contains(items, target)
}

func main() {
// Задача 3.5.1. Универсальный возврат

// Напиши функцию:
// func Echo[T any](v T) T
// Проверь ее для int, string, bool.

fmt.Println(Echo("string"))
fmt.Println(Echo(123))
fmt.Println(Echo(false))

// Задача 3.5.2. Первый элемент

// Напиши функцию:
// func First[T any](items []T) (T, bool)

// Требования:
// если срез пустой, возвращай нулевое значение T и false;
// иначе первый элемент и true.

fmt.Println(First([]int{}))
fmt.Println(First([]string{}))
fmt.Println(First([]bool{}))
fmt.Println(First([]int{1,3,5,7}))
fmt.Println(First([]string{"string", "boba"}))
fmt.Println(First([]bool{true, false}))

// Задача 3.5.3. Сумма чисел

// Создай ограничение Number для ~int | ~float64.
// Напиши функцию:
// func Sum[T Number](items []T) T
// Проверь на []int и []float64.

fmt.Println(Sum([]int{1,3,1})) // 5
fmt.Println(Sum([]float64{1.5,3.5,1.5})) // 6.5

// Задача 3.5.4. Поиск индекса

// Напиши функцию:
// func IndexOf[T comparable](items []T, target T) int

// Требования:
// вернуть индекс найденного элемента;
// если элемента нет, вернуть -1.

fmt.Println(IndexOf([]int{}, 2)) // -1
fmt.Println(IndexOf([]string{}, "string")) // -1
fmt.Println(IndexOf([]bool{}, true)) // -1
fmt.Println(IndexOf([]int{1,3,5,7}, 10)) // -1
fmt.Println(IndexOf([]string{"string", "boba"}, "string")) // 0
fmt.Println(IndexOf([]bool{true, false}, false)) // 1

// Задача 3.5.5. Обобщенная структура Pair

// Создай структуру:
// type Pair[K comparable, V any] struct {
// 	Key   K
// 	Value V
// }
// В main создай минимум два объекта Pair с разными типами.

type Pair[K comparable, V any] struct {
	Key   K
	Value V
}

p1 := Pair[int, string]{
	Key: 777,
	Value: "200",
}

p2 := Pair[bool, int]{
	Key: true,
	Value: 100,
}

fmt.Println(p1,p2) // {777 200} {true 100}

// Задача 3.5.6. Значения map в срез

// Напиши функцию:
// func Values[K comparable, V any](m map[K]V) []V
// Проверь на карте map[string]int.
fmt.Println(Values(map[string]int{"1": 1, "2": 2, "3": 3})) // [1 2 3]

// Задача 3.5.7. Интеграционная задача

// Спроектируй обобщенный контейнер Store[T any]:
// поле items []T;
// метод Add(item T);
// метод All() []T.
// Дополнительно реализуй функцию:
// func Contains[T comparable](items []T, target T) bool
// В main:
// создай Store[string] и Store[int],
// добавь элементы,
// выведи содержимое,
// проверь наличие нескольких значений через Contains.
// Пример с Store[string]
stringStore := Store[string]{}
stringStore.Add("Alice")
stringStore.Add("Bob")
fmt.Println("Содержимое stringStore:", stringStore.All()) // Содержимое stringStore: [Alice Bob]
fmt.Println("Содержит 'Alice'? ", Contains(stringStore.All(), "Alice")) // Содержит 'Alice'?  true
fmt.Println("Содержит 'Charlie'? ", Contains(stringStore.All(), "Charlie")) // Содержит 'Charlie'?  false

intStore := Store[int]{}
intStore.Add(10)
intStore.Add(20)
intStore.Add(30)
fmt.Println("Содержимое intStore:", intStore.All()) // Содержимое intStore: [10 20 30]
fmt.Println("Содержит 20? ", Contains(intStore.All(), 20)) // Содержит 20?  true
fmt.Println("Содержит 40? ", Contains(intStore.All(), 40)) // Содержит 40?  false
}